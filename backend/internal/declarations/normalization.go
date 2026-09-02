// Deterministic normalization of raw ML/OCR field values into structured
// declarations. This is pure, testable logic — NO ML and NO legal decisions
// happen here. It converts messy raw strings into
// {normalizedValue, numericValue, unit} that the deterministic rule engine
// can consume. The caller always preserves rawValue/confidence/bbox as
// evidence; nothing here discards OCR evidence.
package declarations

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/raunak0400/metrology-backend/internal/ml"
)

var knownTypes = map[string]bool{
	"MRP": true, "NET_QUANTITY": true, "MANUFACTURER": true, "PACKER": true,
	"IMPORTER": true, "MANUFACTURE_DATE": true, "PACKING_DATE": true,
	"EXPIRY_DATE": true, "CONSUMER_CARE": true, "COUNTRY_OF_ORIGIN": true,
	"DIMENSIONS": true, "UNIT_SALE_PRICE": true, "GENERIC_NAME": true,
	"BEST_BEFORE": true, "USE_BY_DATE": true,
}

var nonWordRunRe = regexp.MustCompile(`\s+`)
var moneyRe = regexp.MustCompile(`(\d+(?:\.\d+)?)`)
var quantityRe = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*(kg|g|mg|l|ml|cl|pcs|n|no\.?|units?)?`)
var commaRe = regexp.MustCompile(`,`)

func coerceType(field string) string {
	upper := nonWordRunRe.ReplaceAllString(strings.ToUpper(strings.TrimSpace(field)), "_")
	if knownTypes[upper] {
		return upper
	}
	return "OTHER"
}

// parseMoney extracts the first monetary amount from a string, e.g.
// "MRP Rs. 199.00" -> 199.
func parseMoney(raw string) *float64 {
	clean := commaRe.ReplaceAllString(raw, "")
	m := moneyRe.FindStringSubmatch(clean)
	if m == nil {
		return nil
	}
	n, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return nil
	}
	return &n
}

// parseQuantity extracts a quantity + unit, e.g. "Net Qty: 500 g" -> (500, "g").
func parseQuantity(raw string) (*float64, *string) {
	clean := commaRe.ReplaceAllString(raw, "")
	m := quantityRe.FindStringSubmatch(clean)
	if m == nil {
		return nil, nil
	}
	n, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return nil, nil
	}
	var unit *string
	if m[2] != "" {
		u := strings.ToLower(m[2])
		unit = &u
	}
	return &n, unit
}

// Normalized is the deterministic normalization output.
type Normalized struct {
	Type            string
	NormalizedValue *string
	NumericValue    *float64
	Unit            *string
}

func sp(s string) *string { return &s }

// Normalize converts one raw extracted field into a normalized declaration.
func Normalize(field ml.ExtractedField) Normalized {
	t := coerceType(field.Field)
	raw := strings.TrimSpace(field.RawValue)

	switch t {
	case "MRP", "UNIT_SALE_PRICE":
		value := parseMoney(raw)
		var normalized *string
		if value != nil {
			normalized = sp(strconv.FormatFloat(*value, 'f', 2, 64))
		} else if raw != "" {
			normalized = sp(raw)
		}
		return Normalized{Type: t, NormalizedValue: normalized, NumericValue: value, Unit: sp("INR")}

	case "NET_QUANTITY":
		value, unit := parseQuantity(raw)
		var normalized *string
		if value != nil {
			s := strconv.FormatFloat(*value, 'f', -1, 64)
			if unit != nil {
				s += " " + *unit
			}
			normalized = sp(s)
		} else if raw != "" {
			normalized = sp(raw)
		}
		return Normalized{Type: t, NormalizedValue: normalized, NumericValue: value, Unit: unit}

	case "MANUFACTURE_DATE", "PACKING_DATE", "EXPIRY_DATE", "BEST_BEFORE", "USE_BY_DATE":
		normalized := nonWordRunRe.ReplaceAllString(raw, " ")
		var np *string
		if normalized != "" {
			np = sp(normalized)
		}
		return Normalized{Type: t, NormalizedValue: np}

	default:
		var np *string
		if raw != "" {
			np = sp(raw)
		}
		return Normalized{Type: t, NormalizedValue: np}
	}
}
