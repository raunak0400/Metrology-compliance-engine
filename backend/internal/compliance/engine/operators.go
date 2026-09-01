package engine

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// manualReviewPrefix marks a reason string as "not a definite failure — route
// to manual review" without adding another field to OperatorResult (mirrors
// the "MANUAL_REVIEW:" string-prefix convention used by the original engine).
const manualReviewPrefix = "MANUAL_REVIEW:"

// OperatorResult is a pure, deterministic verdict. Operators never "throw" for
// normal failing input — they always return a result.
type OperatorResult struct {
	Pass     bool
	Actual   *string
	Expected *string
	Reason   *string
}

func ok(actual *string) OperatorResult {
	return OperatorResult{Pass: true, Actual: actual}
}

func fail(actual, expected *string, reason string) OperatorResult {
	return OperatorResult{Pass: false, Actual: actual, Expected: expected, Reason: &reason}
}

func manualReview(actual, expected *string, reason string) OperatorResult {
	r := manualReviewPrefix + reason
	return OperatorResult{Pass: false, Actual: actual, Expected: expected, Reason: &r}
}

func sp(s string) *string { return &s }

type OperatorInput struct {
	Declaration *Declaration
	Measurement *Measurement
	Expected    map[string]interface{}
	Context     Context
}

func expectedNumber(expected map[string]interface{}, key string) (float64, bool) {
	if expected == nil {
		return 0, false
	}
	switch v := expected[key].(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case string:
		n, err := strconv.ParseFloat(v, 64)
		return n, err == nil
	default:
		return 0, false
	}
}

func expectedString(expected map[string]interface{}, key string) string {
	if expected == nil {
		return ""
	}
	if v, ok := expected[key].(string); ok {
		return v
	}
	return ""
}

func expectedStringSlice(expected map[string]interface{}, key string) []string {
	if expected == nil {
		return nil
	}
	raw, ok := expected[key].([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// ApplyOperator evaluates one requirement operator against the declaration
// (and, for measurement operators, the measurement) currently in scope.
func ApplyOperator(operator Operator, in OperatorInput) OperatorResult {
	decl := in.Declaration
	meas := in.Measurement
	expected := in.Expected
	ctx := in.Context

	var value *string
	if decl != nil {
		value = decl.NormalizedValue
	}

	switch operator {
	case OpExists:
		if decl != nil {
			return ok(value)
		}
		return fail(nil, sp("present"), "Declaration is absent")

	case OpNotEmpty:
		if value != nil && strings.TrimSpace(*value) != "" {
			return ok(value)
		}
		return fail(value, sp("non-empty"), "Declaration is empty")

	case OpRegexMatch:
		pattern := expectedString(expected, "pattern")
		if pattern == "" {
			return fail(value, nil, "No regex pattern configured")
		}
		if value == nil {
			return fail(nil, sp(pattern), "No value to match")
		}
		re, err := regexp.Compile(goRegex(pattern))
		if err != nil {
			return fail(value, sp(pattern), "Invalid regex pattern configured")
		}
		if re.MatchString(*value) {
			return ok(value)
		}
		return fail(value, sp(pattern), "Value does not match required pattern")

	case OpNumericMin:
		min, _ := expectedNumber(expected, "min")
		if decl == nil || decl.NumericValue == nil {
			return fail(value, sp(fmt.Sprintf(">= %v", min)), "No numeric value")
		}
		n := *decl.NumericValue
		if n >= min {
			return ok(sp(formatNumber(n)))
		}
		return fail(sp(formatNumber(n)), sp(fmt.Sprintf(">= %v", min)), fmt.Sprintf("Value below minimum %v", min))

	case OpNumericMax:
		max, _ := expectedNumber(expected, "max")
		if decl == nil || decl.NumericValue == nil {
			return fail(value, sp(fmt.Sprintf("<= %v", max)), "No numeric value")
		}
		n := *decl.NumericValue
		if n <= max {
			return ok(sp(formatNumber(n)))
		}
		return fail(sp(formatNumber(n)), sp(fmt.Sprintf("<= %v", max)), fmt.Sprintf("Value above maximum %v", max))

	case OpNumericRange:
		min, _ := expectedNumber(expected, "min")
		max, _ := expectedNumber(expected, "max")
		if decl == nil || decl.NumericValue == nil {
			return fail(value, sp(fmt.Sprintf("%v..%v", min, max)), "No numeric value")
		}
		n := *decl.NumericValue
		if n >= min && n <= max {
			return ok(sp(formatNumber(n)))
		}
		return fail(sp(formatNumber(n)), sp(fmt.Sprintf("%v..%v", min, max)), fmt.Sprintf("Value outside range %v..%v", min, max))

	case OpMinTextHeightMM:
		minMm, _ := expectedNumber(expected, "minMm")
		if meas == nil || !meas.HasValidScale || meas.HeightMm == nil {
			actual := "unknown"
			if meas != nil && meas.HeightMm != nil {
				actual = fmt.Sprintf("%vmm", *meas.HeightMm)
			}
			return manualReview(sp(actual), sp(fmt.Sprintf(">= %vmm", minMm)), "No valid scale reference for pixel-to-mm conversion")
		}
		if *meas.HeightMm >= minMm {
			return ok(sp(fmt.Sprintf("%vmm", *meas.HeightMm)))
		}
		return fail(sp(fmt.Sprintf("%vmm", *meas.HeightMm)), sp(fmt.Sprintf(">= %vmm", minMm)), fmt.Sprintf("Text height below minimum %vmm", minMm))

	case OpEnumIn:
		allowed := expectedStringSlice(expected, "values")
		if value == nil {
			return fail(nil, sp(strings.Join(allowed, "|")), "No value")
		}
		for _, a := range allowed {
			if strings.EqualFold(a, *value) {
				return ok(value)
			}
		}
		return fail(value, sp(strings.Join(allowed, "|")), "Value not in allowed set")

	case OpDateValid:
		if value == nil {
			return fail(nil, sp("valid date"), "No date value")
		}
		if _, ok := parseDate(*value); !ok {
			return fail(value, sp("valid date"), "Not a valid date")
		}
		return ok(value)

	case OpDateNotFuture:
		if value == nil {
			return fail(nil, sp("<= today"), "No date value")
		}
		d, okParse := parseDate(*value)
		if !okParse {
			return fail(value, sp("<= today"), "Not a valid date")
		}
		if !d.After(time.Now()) {
			return ok(value)
		}
		return fail(value, sp("<= today"), "Date is in the future")

	case OpRequiredIfImported:
		if !ctx.IsImported {
			return ok(value) // requirement not applicable
		}
		if decl != nil && value != nil && strings.TrimSpace(*value) != "" {
			return ok(value)
		}
		return fail(value, sp("present (imported package)"), "Required declaration missing for imported package")

	case OpCustom:
		fallthrough
	default:
		// Unknown/custom operators are not auto-decided — route to manual review.
		return manualReview(value, nil, "Operator requires manual evaluation")
	}
}

// goRegex translates the small set of inline-flag syntax the seeded rule
// templates use ("(?i)...") which Go's RE2 also supports natively, so this is
// currently a passthrough kept as an explicit seam for future translation.
func goRegex(pattern string) string { return pattern }

func formatNumber(n float64) string {
	if n == float64(int64(n)) {
		return strconv.FormatInt(int64(n), 10)
	}
	return strconv.FormatFloat(n, 'f', -1, 64)
}

func parseDate(s string) (time.Time, bool) {
	layouts := []string{time.RFC3339, "2006-01-02", "2006-01-02T15:04:05", "01/02/2006", "2006/01/02"}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
