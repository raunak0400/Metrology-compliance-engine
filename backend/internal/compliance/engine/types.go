// Package engine is the DETERMINISTIC legal rule engine.
//
// It is a pure function: (declarations + measurements + applicable rules) ->
// (compliance checks + overall outcome). It performs NO database access, NO
// network calls, and contains NO ML. An LLM/ML model must never make the
// compliance decision — only this engine does, from structured inputs.
package engine

type ComplianceStatus string

const (
	StatusPass          ComplianceStatus = "PASS"
	StatusFail          ComplianceStatus = "FAIL"
	StatusManualReview  ComplianceStatus = "MANUAL_REVIEW"
	StatusNotApplicable ComplianceStatus = "NOT_APPLICABLE"
)

type Outcome string

const (
	OutcomeCompliant    Outcome = "COMPLIANT"
	OutcomeNonCompliant Outcome = "NON_COMPLIANT"
	OutcomeManualReview Outcome = "MANUAL_REVIEW"
)

type Severity string

const (
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

type Operator string

const (
	OpExists             Operator = "EXISTS"
	OpNotEmpty           Operator = "NOT_EMPTY"
	OpRegexMatch         Operator = "REGEX_MATCH"
	OpNumericMin         Operator = "NUMERIC_MIN"
	OpNumericMax         Operator = "NUMERIC_MAX"
	OpNumericRange       Operator = "NUMERIC_RANGE"
	OpMinTextHeightMM    Operator = "MIN_TEXT_HEIGHT_MM"
	OpEnumIn             Operator = "ENUM_IN"
	OpDateValid          Operator = "DATE_VALID"
	OpDateNotFuture      Operator = "DATE_NOT_FUTURE"
	OpRequiredIfImported Operator = "REQUIRED_IF_IMPORTED"
	OpCustom             Operator = "CUSTOM"
)

// Declaration is the normalized declaration as the engine sees it (a subset
// of the extracted_declarations row).
type Declaration struct {
	DeclarationID   string
	Type            string
	NormalizedValue *string
	NumericValue    *float64
	Unit            *string
	Confidence      *float64
}

// Measurement is a measurement result the engine consumes (produced by the
// measurement package). The engine never computes measurements itself.
type Measurement struct {
	DeclarationID string
	Type          string
	// HeightMm is the estimated real-world height in mm, or nil when no valid
	// scale reference exists.
	HeightMm      *float64
	HasValidScale bool
}

type Requirement struct {
	ID              string
	Key             string
	DeclarationType *string
	Operator        Operator
	Expected        map[string]interface{}
	Unit            *string
	Severity        *Severity
	Message         *string
	IsMandatory     bool
}

type Rule struct {
	ID              string
	Code            string
	Title           string
	ActReference    *string
	SeverityDefault Severity
	Requirements    []Requirement
}

// Context is everything the engine needs to evaluate a set of rules against
// one inspection.
type Context struct {
	IsImported   bool
	Declarations []Declaration
	Measurements []Measurement
	// ManualReviewThreshold: below this confidence, a check is routed to
	// MANUAL_REVIEW, never auto-FAIL.
	ManualReviewThreshold float64
}

type CheckResult struct {
	RuleID               string
	RuleCode             string
	RequirementID        string
	RequirementKey       string
	DeclarationID        *string
	Status               ComplianceStatus
	Severity             Severity
	ActualValue          *string
	ExpectedValue        *string
	Reason               *string
	Confidence           *float64
	RequiresManualReview bool
	LegalReference       *string
	Title                string
}

type Output struct {
	Checks  []CheckResult
	Outcome Outcome
}
