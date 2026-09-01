package engine

import "strings"

func findDeclarationByType(items []Declaration, t *string) *Declaration {
	if t == nil {
		return nil
	}
	for i := range items {
		if items[i].Type == *t {
			return &items[i]
		}
	}
	return nil
}

func findMeasurementByType(items []Measurement, t *string) *Measurement {
	if t == nil {
		return nil
	}
	for i := range items {
		if items[i].Type == *t {
			return &items[i]
		}
	}
	return nil
}

// Evaluate runs every requirement of every rule against the context and
// returns one CheckResult per requirement plus an overall outcome. Low-
// confidence extractions and ambiguous measurement cases are routed to
// MANUAL_REVIEW — never auto-converted into legal violations.
func Evaluate(rules []Rule, ctx Context) Output {
	checks := make([]CheckResult, 0)

	for _, rule := range rules {
		for _, req := range rule.Requirements {
			declaration := findDeclarationByType(ctx.Declarations, req.DeclarationType)
			measurement := findMeasurementByType(ctx.Measurements, req.DeclarationType)

			result := ApplyOperator(req.Operator, OperatorInput{
				Declaration: declaration,
				Measurement: measurement,
				Expected:    req.Expected,
				Context:     ctx,
			})

			isManualByOperator := result.Reason != nil && strings.HasPrefix(*result.Reason, manualReviewPrefix)
			lowConfidence := declaration != nil && declaration.Confidence != nil && *declaration.Confidence < ctx.ManualReviewThreshold

			var status ComplianceStatus
			requiresManualReview := false
			reason := result.Reason

			switch {
			case isManualByOperator:
				status = StatusManualReview
				requiresManualReview = true
				trimmed := strings.TrimPrefix(*result.Reason, manualReviewPrefix)
				reason = &trimmed
			case result.Pass:
				status = StatusPass
			case lowConfidence:
				// A failing check on a low-confidence extraction is ambiguous —
				// do NOT auto-create a violation; send it for human review instead.
				status = StatusManualReview
				requiresManualReview = true
				base := "Failed check"
				if result.Reason != nil {
					base = *result.Reason
				}
				combined := base + " (low OCR confidence " + formatConfidence(*declaration.Confidence) + ")"
				reason = &combined
			default:
				status = StatusFail
			}

			severity := rule.SeverityDefault
			if req.Severity != nil {
				severity = *req.Severity
			}

			title := rule.Title
			if req.Message != nil {
				title = *req.Message
			}

			var declID *string
			if declaration != nil {
				declID = sp(declaration.DeclarationID)
			}
			var confidence *float64
			if declaration != nil {
				confidence = declaration.Confidence
			}

			checks = append(checks, CheckResult{
				RuleID:               rule.ID,
				RuleCode:             rule.Code,
				RequirementID:        req.ID,
				RequirementKey:       req.Key,
				DeclarationID:        declID,
				Status:               status,
				Severity:             severity,
				ActualValue:          result.Actual,
				ExpectedValue:        result.Expected,
				Reason:               reason,
				Confidence:           confidence,
				RequiresManualReview: requiresManualReview,
				LegalReference:       rule.ActReference,
				Title:                title,
			})
		}
	}

	return Output{Checks: checks, Outcome: deriveOutcome(rules, checks)}
}

// deriveOutcome: any manual review dominates; else any fail => non-compliant.
func deriveOutcome(rules []Rule, checks []CheckResult) Outcome {
	mandatory := make(map[string]bool)
	for _, r := range rules {
		for _, req := range r.Requirements {
			if req.IsMandatory {
				mandatory[req.ID] = true
			}
		}
	}

	hasManualReview := false
	hasFail := false
	for _, c := range checks {
		if !mandatory[c.RequirementID] {
			continue
		}
		if c.Status == StatusManualReview {
			hasManualReview = true
		}
		if c.Status == StatusFail {
			hasFail = true
		}
	}

	if hasManualReview {
		return OutcomeManualReview
	}
	if hasFail {
		return OutcomeNonCompliant
	}
	return OutcomeCompliant
}

func formatConfidence(c float64) string {
	return formatNumber(c)
}
