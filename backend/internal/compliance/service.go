// Compliance orchestration.
//
// Loads the normalized declarations + measurements + applicable legal rules
// from PostgreSQL, runs the DETERMINISTIC rule engine, and persists the
// resulting compliance checks and violations. Then it updates the inspection
// outcome.
//
// Pipeline position:
//
//	declarations + measurements + rules -> engine -> checks -> violations
//
// The engine (not any ML/LLM) decides compliance.
package compliance

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raunak0400/metrology-backend/internal/compliance/engine"
	"github.com/raunak0400/metrology-backend/internal/constants"
	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/dbutil"
	"github.com/raunak0400/metrology-backend/internal/inspections"
	"github.com/raunak0400/metrology-backend/internal/taskqueue"
)

type Service struct {
	q       *db.Queries
	pool    *pgxpool.Pool
	asynq   *asynq.Client
	inspSvc *inspections.Service
}

func NewService(q *db.Queries, pool *pgxpool.Pool, asynqClient *asynq.Client, inspSvc *inspections.Service) *Service {
	return &Service{q: q, pool: pool, asynq: asynqClient, inspSvc: inspSvc}
}

// loadApplicableRules loads active legal rules applicable to a category
// (category-mapped or global) with their requirements, in engine shape.
func (s *Service) loadApplicableRules(ctx context.Context, categoryID uuid.NullUUID) ([]engine.Rule, error) {
	rows, err := s.q.ListApplicableRules(ctx, categoryID)
	if err != nil {
		return nil, err
	}

	rules := make([]engine.Rule, 0, len(rows))
	for _, r := range rows {
		reqRows, err := s.q.ListRuleRequirements(ctx, r.ID)
		if err != nil {
			return nil, err
		}
		requirements := make([]engine.Requirement, 0, len(reqRows))
		for _, q := range reqRows {
			var declType *string
			if q.DeclarationType != nil {
				v := string(*q.DeclarationType)
				declType = &v
			}
			var expected map[string]interface{}
			if len(q.Expected) > 0 {
				_ = json.Unmarshal(q.Expected, &expected)
			}
			var severity *engine.Severity
			if q.Severity != nil {
				v := engine.Severity(*q.Severity)
				severity = &v
			}
			requirements = append(requirements, engine.Requirement{
				ID: q.ID.String(), Key: q.Key, DeclarationType: declType,
				Operator: engine.Operator(q.Operator), Expected: expected, Unit: q.Unit,
				Severity: severity, Message: q.Message, IsMandatory: q.IsMandatory,
			})
		}
		rules = append(rules, engine.Rule{
			ID: r.ID.String(), Code: r.Code, Title: r.Title, ActReference: r.ActReference,
			SeverityDefault: engine.Severity(r.SeverityDefault), Requirements: requirements,
		})
	}
	return rules, nil
}

func (s *Service) loadDeclarations(ctx context.Context, inspectionID uuid.UUID) ([]engine.Declaration, error) {
	rows, err := s.q.ListDeclarationsByInspection(ctx, inspectionID)
	if err != nil {
		return nil, err
	}
	out := make([]engine.Declaration, 0, len(rows))
	for _, d := range rows {
		id := d.ID.String()
		out = append(out, engine.Declaration{
			DeclarationID: id, Type: string(d.Type), NormalizedValue: d.NormalizedValue,
			NumericValue: db.NumericToFloat64(d.NumericValue), Unit: d.Unit, Confidence: d.Confidence,
		})
	}
	return out, nil
}

func (s *Service) loadMeasurements(ctx context.Context, inspectionID uuid.UUID) ([]engine.Measurement, error) {
	rows, err := s.q.ListDeclarationsWithScaleByInspection(ctx, inspectionID)
	if err != nil {
		return nil, err
	}
	scaleRows := make([]DeclarationScaleRow, 0, len(rows))
	for _, r := range rows {
		scaleRows = append(scaleRows, DeclarationScaleRow{
			DeclarationID: r.ID.String(), Type: string(r.Type), BBoxJSON: r.Bbox, ScaleReference: r.ScaleReference,
		})
	}
	return ComputeMeasurements(scaleRows), nil
}

type RunSummary struct {
	InspectionID string         `json:"inspectionId"`
	Outcome      engine.Outcome `json:"outcome"`
	TotalChecks  int            `json:"totalChecks"`
	Passed       int            `json:"passed"`
	Failed       int            `json:"failed"`
	ManualReview int            `json:"manualReview"`
	Violations   int            `json:"violations"`
}

// Run executes the full deterministic compliance evaluation and persists the results.
func (s *Service) Run(ctx context.Context, inspectionID uuid.UUID) (RunSummary, error) {
	insp, err := s.inspSvc.Ensure(ctx, inspectionID)
	if err != nil {
		return RunSummary{}, err
	}

	declarations, err := s.loadDeclarations(ctx, inspectionID)
	if err != nil {
		return RunSummary{}, err
	}
	measurements, err := s.loadMeasurements(ctx, inspectionID)
	if err != nil {
		return RunSummary{}, err
	}
	rules, err := s.loadApplicableRules(ctx, uuidToNull(insp.CategoryID))
	if err != nil {
		return RunSummary{}, err
	}

	isImported := false
	for _, d := range declarations {
		if d.Type == "IMPORTER" && d.NormalizedValue != nil && *d.NormalizedValue != "" {
			isImported = true
			break
		}
	}

	output := engine.Evaluate(rules, engine.Context{
		IsImported: isImported, Declarations: declarations, Measurements: measurements,
		ManualReviewThreshold: constants.ManualReviewConfidenceThreshold,
	})

	if err := s.persistResults(ctx, inspectionID, output.Checks, output.Outcome); err != nil {
		return RunSummary{}, err
	}

	summary := RunSummary{InspectionID: inspectionID.String(), Outcome: output.Outcome, TotalChecks: len(output.Checks)}
	for _, c := range output.Checks {
		switch c.Status {
		case engine.StatusPass:
			summary.Passed++
		case engine.StatusFail:
			summary.Failed++
			summary.Violations++
		case engine.StatusManualReview:
			summary.ManualReview++
		}
	}
	return summary, nil
}

func uuidToNull(id uuid.NullUUID) uuid.NullUUID { return id }

type checksDetails struct {
	RuleCode       string `json:"ruleCode"`
	RequirementKey string `json:"requirementKey"`
}

// persistResults persists checks + violations and updates the inspection
// outcome inside a single PostgreSQL transaction — either the full evaluation
// (checks, violations, outcome) lands, or none of it does, so a mid-write
// failure never leaves the inspection with a stale/partial result. Idempotent
// re-run: clears the prior evaluation first.
func (s *Service) persistResults(ctx context.Context, inspectionID uuid.UUID, checks []engine.CheckResult, outcome engine.Outcome) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := s.q.WithTx(tx)

	if err := q.DeleteViolationsByInspection(ctx, inspectionID); err != nil {
		return err
	}
	if err := q.DeleteComplianceChecksByInspection(ctx, inspectionID); err != nil {
		return err
	}

	for _, c := range checks {
		ruleID, err := uuid.Parse(c.RuleID)
		if err != nil {
			continue
		}
		requirementID, _ := uuid.Parse(c.RequirementID)
		var declarationID uuid.NullUUID
		if c.DeclarationID != nil {
			if id, err := uuid.Parse(*c.DeclarationID); err == nil {
				declarationID = uuid.NullUUID{UUID: id, Valid: true}
			}
		}
		details := dbutil.JSONBytes(checksDetails{RuleCode: c.RuleCode, RequirementKey: c.RequirementKey})

		check, err := q.CreateComplianceCheck(ctx, db.CreateComplianceCheckParams{
			InspectionID: inspectionID, RuleID: ruleID,
			RequirementID: uuid.NullUUID{UUID: requirementID, Valid: requirementID != uuid.Nil},
			DeclarationID: declarationID, Status: db.ComplianceStatus(c.Status),
			ActualValue: c.ActualValue, ExpectedValue: c.ExpectedValue, Reason: c.Reason,
			Confidence: c.Confidence, RequiresManualReview: c.RequiresManualReview, Details: details,
		})
		if err != nil {
			return err
		}

		// Only definite failures of a MANDATORY requirement become violations.
		// WARNING (non-mandatory fail) and MANUAL_REVIEW do NOT.
		if c.Status == engine.StatusFail {
			if _, err := q.CreateViolation(ctx, db.CreateViolationParams{
				InspectionID: inspectionID, RuleID: uuid.NullUUID{UUID: ruleID, Valid: true},
				CheckID: uuid.NullUUID{UUID: check.ID, Valid: true}, Severity: db.ViolationSeverity(c.Severity),
				Title: c.Title, Reason: c.Reason, LegalReference: c.LegalReference,
			}); err != nil {
				return err
			}
		}
	}

	status := db.InspectionStatusCOMPLETED
	if outcome == engine.OutcomeManualReview {
		status = db.InspectionStatusAWAITINGREVIEW
	}
	if err := q.SetInspectionOutcome(ctx, db.SetInspectionOutcomeParams{
		ID: inspectionID, Status: status, Outcome: db.InspectionOutcome(outcome),
	}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Enqueue schedules a compliance evaluation on the compliance-check queue. The
// engine is deterministic and fast, but it still runs via the worker so
// status transitions and retries are consistent with the rest of the pipeline.
func (s *Service) Enqueue(ctx context.Context, inspectionID uuid.UUID, triggeredByUserID string) (string, error) {
	if _, err := s.inspSvc.Ensure(ctx, inspectionID); err != nil {
		return "", err
	}

	payload, err := json.Marshal(taskqueue.CompliancePayload{InspectionID: inspectionID.String(), TriggeredByUserID: triggeredByUserID})
	if err != nil {
		return "", err
	}
	info, err := s.asynq.Enqueue(asynq.NewTask(taskqueue.TaskComplianceCheck, payload), asynq.Queue(taskqueue.QueueCompliance))
	if err != nil {
		return "", err
	}
	_, _ = s.q.UpdateInspection(ctx, db.UpdateInspectionParams{ID: inspectionID, Status: statusPtr(db.InspectionStatusPROCESSING)})
	return info.ID, nil
}

func statusPtr(s db.InspectionStatus) *db.InspectionStatus { return &s }

func (s *Service) ListChecks(ctx context.Context, inspectionID uuid.UUID) ([]db.ListComplianceChecksByInspectionRow, error) {
	if _, err := s.inspSvc.Ensure(ctx, inspectionID); err != nil {
		return nil, err
	}
	return s.q.ListComplianceChecksByInspection(ctx, inspectionID)
}
