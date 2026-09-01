package violations

import (
	"context"

	"github.com/google/uuid"
	"github.com/raunak0400/metrology-backend/internal/apperror"
	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/dbutil"
	"github.com/raunak0400/metrology-backend/internal/inspections"
)

type Service struct {
	q       *db.Queries
	inspSvc *inspections.Service
}

func NewService(q *db.Queries, inspSvc *inspections.Service) *Service {
	return &Service{q: q, inspSvc: inspSvc}
}

func (s *Service) ListByInspection(ctx context.Context, inspectionID uuid.UUID) ([]db.ListViolationsByInspectionRow, error) {
	if _, err := s.inspSvc.Ensure(ctx, inspectionID); err != nil {
		return nil, err
	}
	return s.q.ListViolationsByInspection(ctx, inspectionID)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (db.Violation, error) {
	v, err := s.q.GetViolation(ctx, id)
	if err != nil {
		return db.Violation{}, apperror.NotFound("Violation")
	}
	return v, nil
}

// Update is a reviewer action on a violation: confirm, dismiss or resolve.
// Records who acted. This is a human decision layered on top of the
// deterministic engine output — it does not alter the original check.
func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateViolationRequest, resolvedByID string) (db.Violation, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return db.Violation{}, err
	}

	terminal := req.Status == "DISMISSED" || req.Status == "RESOLVED"
	params := db.UpdateViolationParams{
		ID: id, ReviewNotes: dbutil.StrPtr(req.ReviewNotes), Terminal: &terminal,
	}
	if terminal {
		params.ResolvedByID = dbutil.NullUUID(resolvedByID)
	}
	if req.Status != "" {
		v := db.ViolationStatus(req.Status)
		params.Status = &v
	}
	if req.Severity != "" {
		v := db.ViolationSeverity(req.Severity)
		params.Severity = &v
	}

	return s.q.UpdateViolation(ctx, params)
}
