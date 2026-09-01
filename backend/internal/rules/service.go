package rules

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/raunak0400/metrology-backend/internal/apperror"
	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/dbutil"
	"github.com/raunak0400/metrology-backend/internal/pagination"
)

type Service struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewService(q *db.Queries, pool *pgxpool.Pool) *Service { return &Service{q: q, pool: pool} }

type Detail struct {
	db.LegalRule
	Requirements []db.RuleRequirement       `json:"requirements"`
	Categories   []db.ListRuleCategoriesRow `json:"categories"`
}

func (s *Service) detail(ctx context.Context, rule db.LegalRule) (Detail, error) {
	reqs, err := s.q.ListRuleRequirements(ctx, rule.ID)
	if err != nil {
		return Detail{}, err
	}
	cats, err := s.q.ListRuleCategories(ctx, rule.ID)
	if err != nil {
		return Detail{}, err
	}
	return Detail{LegalRule: rule, Requirements: reqs, Categories: cats}, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (Detail, error) {
	rule, err := s.q.GetLegalRule(ctx, id)
	if err != nil {
		return Detail{}, apperror.NotFound("Rule")
	}
	return s.detail(ctx, rule)
}

func (s *Service) List(ctx context.Context, q ListRulesQuery, page pagination.Query) ([]db.ListLegalRulesRow, int64, error) {
	params := db.ListLegalRulesParams{Limit: page.LimitInt32(), Offset: page.Offset()}
	countParams := db.CountLegalRulesParams{}
	if q.IsActive != "" {
		v := q.IsActive == "true"
		params.IsActive = &v
		countParams.IsActive = &v
	}
	if q.Search != "" {
		params.Search = &q.Search
		countParams.Search = &q.Search
	}

	items, err := s.q.ListLegalRules(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountLegalRules(ctx, countParams)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func parseTime(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return nil
	}
	return &t
}

func (s *Service) Create(ctx context.Context, req CreateRuleRequest, createdByID string) (Detail, error) {
	var severity *db.ViolationSeverity
	if req.SeverityDefault != "" {
		v := db.ViolationSeverity(req.SeverityDefault)
		severity = &v
	}

	rule, err := s.q.CreateLegalRule(ctx, db.CreateLegalRuleParams{
		Code: req.Code, Title: req.Title, Description: dbutil.StrPtr(req.Description),
		ActReference: dbutil.StrPtr(req.ActReference), Section: dbutil.StrPtr(req.Section),
		SeverityDefault: severity, EffectiveFrom: parseTime(req.EffectiveFrom), EffectiveTo: parseTime(req.EffectiveTo),
		IsActive: req.IsActive, Params: dbutil.JSONBytes(req.Params), CreatedByID: dbutil.NullUUID(createdByID),
	})
	if err != nil {
		return Detail{}, err
	}

	for i, r := range req.Requirements {
		if err := s.createRequirement(ctx, s.q, rule.ID, r, i); err != nil {
			return Detail{}, err
		}
	}
	for _, catID := range req.CategoryIDs {
		if id, err := uuid.Parse(catID); err == nil {
			_ = s.q.CreateRuleCategory(ctx, db.CreateRuleCategoryParams{RuleID: rule.ID, CategoryID: id})
		}
	}

	return s.detail(ctx, rule)
}

func (s *Service) createRequirement(ctx context.Context, q *db.Queries, ruleID uuid.UUID, r RequirementInput, orderIndex int) error {
	var declType *db.DeclarationType
	if r.DeclarationType != "" {
		v := db.DeclarationType(r.DeclarationType)
		declType = &v
	}
	var severity *db.ViolationSeverity
	if r.Severity != "" {
		v := db.ViolationSeverity(r.Severity)
		severity = &v
	}
	isMandatory := true
	if r.IsMandatory != nil {
		isMandatory = *r.IsMandatory
	}
	idx := int32(orderIndex)
	if r.OrderIndex != nil {
		idx = int32(*r.OrderIndex)
	}

	_, err := q.CreateRuleRequirement(ctx, db.CreateRuleRequirementParams{
		RuleID: ruleID, Key: r.Key, Operator: db.RequirementOperator(r.Operator),
		IsMandatory: isMandatory, OrderIndex: idx, DeclarationType: declType,
		Expected: dbutil.JSONBytes(r.Expected), Unit: dbutil.StrPtr(r.Unit), Severity: severity, Message: dbutil.StrPtr(r.Message),
	})
	return err
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateRuleRequest) (Detail, error) {
	rule, err := s.q.GetLegalRule(ctx, id)
	if err != nil {
		return Detail{}, apperror.NotFound("Rule")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Detail{}, err
	}
	defer tx.Rollback(ctx)
	qtx := s.q.WithTx(tx)

	if req.Requirements != nil {
		if err := qtx.DeleteRuleRequirements(ctx, id); err != nil {
			return Detail{}, err
		}
		for i, r := range req.Requirements {
			if err := s.createRequirement(ctx, qtx, id, r, i); err != nil {
				return Detail{}, err
			}
		}
	}
	if req.CategoryIDs != nil {
		if err := qtx.DeleteRuleCategories(ctx, id); err != nil {
			return Detail{}, err
		}
		for _, catID := range req.CategoryIDs {
			if cid, err := uuid.Parse(catID); err == nil {
				_ = qtx.CreateRuleCategory(ctx, db.CreateRuleCategoryParams{RuleID: id, CategoryID: cid})
			}
		}
	}

	params := db.UpdateLegalRuleParams{
		ID: id, Title: dbutil.StrPtr(req.Title), Description: dbutil.StrPtr(req.Description),
		ActReference: dbutil.StrPtr(req.ActReference), Section: dbutil.StrPtr(req.Section),
		EffectiveFrom: parseTime(req.EffectiveFrom), EffectiveTo: parseTime(req.EffectiveTo),
		IsActive: req.IsActive, Params: dbutil.JSONBytes(req.Params),
	}
	if req.SeverityDefault != "" {
		v := db.ViolationSeverity(req.SeverityDefault)
		params.SeverityDefault = &v
	}
	updated, err := qtx.UpdateLegalRule(ctx, params)
	if err != nil {
		return Detail{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Detail{}, err
	}
	_ = rule
	return s.detail(ctx, updated)
}
