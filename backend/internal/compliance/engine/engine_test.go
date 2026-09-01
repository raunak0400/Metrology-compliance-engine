package engine

import "testing"

func f64(v float64) *float64 { return &v }
func str(v string) *string   { return &v }

func mrpRule() Rule {
	return Rule{
		ID:              "rule-mrp",
		Code:            "MRP-PRESENCE",
		Title:           "MRP present",
		ActReference:    str("PCR 2011 Rule 6"),
		SeverityDefault: SeverityHigh,
		Requirements: []Requirement{
			{
				ID:              "req-mrp-exists",
				Key:             "declaration.MRP.present",
				DeclarationType: str("MRP"),
				Operator:        OpExists,
				Message:         str("MRP missing"),
				IsMandatory:     true,
			},
		},
	}
}

func baseCtx(decls []Declaration) Context {
	return Context{Declarations: decls, Measurements: nil, IsImported: false, ManualReviewThreshold: 0.7}
}

func TestCompliant_WhenMandatoryDeclarationPresentHighConfidence(t *testing.T) {
	out := Evaluate([]Rule{mrpRule()}, baseCtx([]Declaration{
		{DeclarationID: "d1", Type: "MRP", NormalizedValue: str("199.00"), NumericValue: f64(199), Unit: str("INR"), Confidence: f64(0.95)},
	}))
	if out.Outcome != OutcomeCompliant {
		t.Fatalf("expected COMPLIANT, got %s", out.Outcome)
	}
	if out.Checks[0].Status != StatusPass {
		t.Fatalf("expected PASS, got %s", out.Checks[0].Status)
	}
}

func TestNonCompliant_WhenMandatoryDeclarationAbsent(t *testing.T) {
	out := Evaluate([]Rule{mrpRule()}, baseCtx(nil))
	if out.Outcome != OutcomeNonCompliant {
		t.Fatalf("expected NON_COMPLIANT, got %s", out.Outcome)
	}
	if out.Checks[0].Status != StatusFail {
		t.Fatalf("expected FAIL, got %s", out.Checks[0].Status)
	}
}

func TestManualReview_LowConfidenceFailingCheck(t *testing.T) {
	rule := mrpRule()
	rule.Requirements = []Requirement{
		{
			ID:              "req-mrp-format",
			Key:             "declaration.MRP.format",
			DeclarationType: str("MRP"),
			Operator:        OpRegexMatch,
			Expected:        map[string]interface{}{"pattern": `^\d+\.\d{2}$`},
			Message:         str("bad format"),
			IsMandatory:     true,
		},
	}
	out := Evaluate([]Rule{rule}, baseCtx([]Declaration{
		{DeclarationID: "d1", Type: "MRP", NormalizedValue: str("abc"), Confidence: f64(0.4)},
	}))
	if out.Outcome != OutcomeManualReview {
		t.Fatalf("expected MANUAL_REVIEW, got %s", out.Outcome)
	}
	if out.Checks[0].Status != StatusManualReview || !out.Checks[0].RequiresManualReview {
		t.Fatalf("expected manual review check, got %+v", out.Checks[0])
	}
}

func TestManualReview_MinTextHeightWithoutValidScale(t *testing.T) {
	rule := mrpRule()
	rule.Requirements = []Requirement{
		{
			ID:              "req-height",
			Key:             "measurement.minHeight",
			DeclarationType: str("MRP"),
			Operator:        OpMinTextHeightMM,
			Expected:        map[string]interface{}{"minMm": 1.0},
			Unit:            str("mm"),
			Message:         str("too small"),
			IsMandatory:     true,
		},
	}
	ctx := Context{
		Declarations: []Declaration{
			{DeclarationID: "d1", Type: "MRP", NormalizedValue: str("199"), NumericValue: f64(199), Unit: str("INR"), Confidence: f64(0.99)},
		},
		Measurements:          []Measurement{{DeclarationID: "d1", Type: "MRP", HeightMm: nil, HasValidScale: false}},
		IsImported:            false,
		ManualReviewThreshold: 0.7,
	}
	out := Evaluate([]Rule{rule}, ctx)
	if out.Outcome != OutcomeManualReview {
		t.Fatalf("expected MANUAL_REVIEW, got %s", out.Outcome)
	}
}
