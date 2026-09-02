// Package ruleengine is the ONLY place the Go backend talks to the Python
// Legal Metrology Rule Engine service (POST /validate, GET /rules,
// GET /health). It mirrors internal/ml's role for the OCR/NER service: the
// backend never evaluates legal rules itself — it builds a request from
// structured declarations, sends it here, and persists whatever this service
// returns. Types below mirror rule_engine/schemas.py field-for-field
// (snake_case JSON tags match the Python service's contract exactly).
package ruleengine

// BoundingBox mirrors rule_engine.schemas.BoundingBox (x, y, w, h — NOT
// width/height, unlike internal/ml.BBox and internal/compliance.BBox).
type BoundingBox struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	W float64 `json:"w"`
	H float64 `json:"h"`
}

// ExtractedField mirrors rule_engine.schemas.ExtractedField.
type ExtractedField struct {
	FieldType     string      `json:"field_type"`
	Text          string      `json:"text"`
	Confidence    float64     `json:"confidence"`
	BBox          BoundingBox `json:"bbox"`
	FontHeightMM  *float64    `json:"font_height_mm,omitempty"`
	PanelLocation string      `json:"panel_location"`
}

// TextMeasurementEntry mirrors rule_engine.schemas.TextMeasurementEntry.
type TextMeasurementEntry struct {
	Text             string  `json:"text"`
	Confidence       float64 `json:"confidence"`
	HeightMM         float64 `json:"height_mm"`
	MinRequiredMM    float64 `json:"min_required_mm"`
	IsRule7Compliant bool    `json:"is_rule7_compliant"`
}

// ValidationRequest mirrors rule_engine.schemas.ValidationRequest — the exact
// body POSTed to /validate. Unknown extra keys are ignored by the service
// (model_config = ConfigDict(extra="ignore")), so this only needs to carry
// the fields the rule engine actually documents.
type ValidationRequest struct {
	ProductID          string                 `json:"product_id"`
	Category           string                 `json:"category"`
	PackageShape       string                 `json:"package_shape"`
	PackageSurfaceType string                 `json:"package_surface_type"`
	SaleChannel        string                 `json:"sale_channel"`
	NetWeightKg        *float64               `json:"net_weight_kg,omitempty"`
	NetVolumeL         *float64               `json:"net_volume_l,omitempty"`
	IsImported         bool                   `json:"is_imported"`
	ExtractedFields    []ExtractedField       `json:"extracted_fields,omitempty"`
	ExtractedTags      map[string]interface{} `json:"extracted_tags,omitempty"`
	TextMeasurements   []TextMeasurementEntry `json:"text_measurements,omitempty"`
}

// ViolationResult mirrors rule_engine.schemas.ViolationResult.
type ViolationResult struct {
	RuleID        string  `json:"rule_id"`
	Clause        string  `json:"clause"`
	Severity      string  `json:"severity"` // major | minor | advisory
	Description   string  `json:"description"`
	FieldEvidence *string `json:"field_evidence,omitempty"`
	FieldType     *string `json:"field_type,omitempty"`
}

// ApplicabilityNote mirrors rule_engine.schemas.ApplicabilityNote.
type ApplicabilityNote struct {
	RuleID           string  `json:"rule_id"`
	Applied          bool    `json:"applied"`
	Reason           string  `json:"reason"`
	SeverityOverride *string `json:"severity_override,omitempty"`
}

// ValidationResponse mirrors rule_engine.schemas.ValidationResponse — the
// exact body returned by /validate. The Go backend never recomputes any of
// these fields; it only maps them onto compliance_checks/violations rows.
type ValidationResponse struct {
	ProductID                string                   `json:"product_id"`
	RulesetVersion           string                   `json:"ruleset_version"`
	OverallStatus            string                   `json:"overall_status"` // compliant | non_compliant | exempt | requires_manual_review
	ExemptReason             *string                  `json:"exempt_reason,omitempty"`
	Violations               []ViolationResult        `json:"violations"`
	ManualReviewRequired     []string                 `json:"manual_review_required"`
	ApplicabilityNotes       []ApplicabilityNote      `json:"applicability_notes"`
	TotalViolations          int                      `json:"total_violations"`
	MajorViolations          int                      `json:"major_violations"`
	MinorViolations          int                      `json:"minor_violations"`
	AdvisoryViolations       int                      `json:"advisory_violations"`
	SynthesizedFieldsSummary []map[string]interface{} `json:"synthesized_fields_summary"`
}

// RuleRegistryItem mirrors one entry of GET /rules.
type RuleRegistryItem struct {
	RuleID      string `json:"rule_id"`
	Clause      string `json:"clause"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	RuleClass   string `json:"rule_class"`
}

// RuleRegistryResponse mirrors the GET /rules response — the authoritative
// rule registry. The Go backend mirrors this into its local `legal_rules`
// table (see compliance.Service.syncRuleRegistry) purely so compliance_checks
// / violations can carry a foreign key; it is never used to evaluate rules.
type RuleRegistryResponse struct {
	RulesetVersion     string             `json:"ruleset_version"`
	RulesetDescription string             `json:"ruleset_description"`
	Rules              []RuleRegistryItem `json:"rules"`
}

// HealthResult mirrors the GET /health response.
type HealthResult struct {
	Status             string `json:"status"`
	Service            string `json:"service"`
	RulesetVersion     string `json:"ruleset_version"`
	RulesetDescription string `json:"ruleset_description"`
}
