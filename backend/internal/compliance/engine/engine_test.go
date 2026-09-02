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

func TestWarning_NonMandatoryRequirementFails(t *testing.T) {
	rule := mrpRule()
	rule.Requirements = []Requirement{
		{
			ID:              "req-optional-note",
			Key:             "declaration.NET_QUANTITY.recommendedFormat",
			DeclarationType: str("NET_QUANTITY"),
			Operator:        OpRegexMatch,
			Expected:        map[string]interface{}{"pattern": `^\d+\s?(g|kg|ml|l)$`},
			Message:         str("Recommended format not followed"),
			IsMandatory:     false,
		},
	}
	out := Evaluate([]Rule{rule}, baseCtx([]Declaration{
		{DeclarationID: "d1", Type: "NET_QUANTITY", NormalizedValue: str("about 500 grams"), Confidence: f64(0.95)},
	}))
	if out.Checks[0].Status != StatusWarning {
		t.Fatalf("expected WARNING for a failing non-mandatory requirement, got %s", out.Checks[0].Status)
	}
	// A WARNING must never dominate the outcome the way a mandatory FAIL does.
	if out.Outcome != OutcomeCompliant {
		t.Fatalf("expected COMPLIANT outcome (non-mandatory warnings don't fail the inspection), got %s", out.Outcome)
	}
}

func TestNotApplicable_WhenNoRulesApply(t *testing.T) {
	out := Evaluate(nil, baseCtx(nil))
	if out.Outcome != OutcomeNotApplicable {
		t.Fatalf("expected NOT_APPLICABLE when no rules apply, got %s", out.Outcome)
	}
	if len(out.Checks) != 0 {
		t.Fatalf("expected zero checks when no rules apply, got %d", len(out.Checks))
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
