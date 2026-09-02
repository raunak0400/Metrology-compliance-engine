package inspections

import (
	"context"

	"github.com/google/uuid"
	"github.com/raunak0400/metrology-backend/internal/apperror"
	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/dbutil"
	"github.com/raunak0400/metrology-backend/internal/pagination"
)

type Service struct {
	q *db.Queries
}

func NewService(q *db.Queries) *Service { return &Service{q: q} }

// Create creates an inspection. If a product is supplied and no category
// given, the category is inherited from the product.
func (s *Service) Create(ctx context.Context, req CreateInspectionRequest, inspectorID uuid.UUID) (db.Inspection, error) {
	categoryID := dbutil.NullUUID(req.CategoryID)
	if !categoryID.Valid && req.ProductID != "" {
		if pid, err := uuid.Parse(req.ProductID); err == nil {
			if product, err := s.q.GetProductByID(ctx, pid); err == nil {
				categoryID = uuid.NullUUID{UUID: product.CategoryID, Valid: true}
			}
		}
	}

	return s.q.CreateInspection(ctx, db.CreateInspectionParams{
		Code:        dbutil.GenerateCode("INSP"),
		InspectorID: inspectorID,
		ProductID:   dbutil.NullUUID(req.ProductID),
		CategoryID:  categoryID,
		Source:      dbutil.StrPtr(req.Source),
		Notes:       dbutil.StrPtr(req.Notes),
		Metadata:    dbutil.JSONBytes(req.Metadata),
		Location:    dbutil.StrPtr(req.Location),
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
	})
}

func (s *Service) List(ctx context.Context, q ListInspectionsQuery, page pagination.Query) ([]db.ListInspectionsRow, int64, error) {
	var status *db.InspectionStatus
	if q.Status != "" {
		v := db.InspectionStatus(q.Status)
		status = &v
	}
	var outcome *db.InspectionOutcome
	if q.Outcome != "" {
		v := db.InspectionOutcome(q.Outcome)
		outcome = &v
	}
	inspectorID := dbutil.NullUUID(q.InspectorID)
	productID := dbutil.NullUUID(q.ProductID)

	items, err := s.q.ListInspections(ctx, db.ListInspectionsParams{
		Limit: page.LimitInt32(), Offset: page.Offset(),
		Status: status, Outcome: outcome, InspectorID: inspectorID, ProductID: productID,
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountInspections(ctx, db.CountInspectionsParams{
		Status: status, Outcome: outcome, InspectorID: inspectorID, ProductID: productID,
	})
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (db.GetInspectionDetailRow, error) {
	row, err := s.q.GetInspectionDetail(ctx, id)
	if err != nil {
		return db.GetInspectionDetailRow{}, apperror.NotFound("Inspection")
	}
	return row, nil
}

// Ensure throws NotFoundError if the inspection does not exist; returns the
// bare row otherwise. Used by every other module that hangs off an inspection.
func (s *Service) Ensure(ctx context.Context, id uuid.UUID) (db.Inspection, error) {
	insp, err := s.q.GetInspection(ctx, id)
	if err != nil {
		return db.Inspection{}, apperror.NotFound("Inspection")
	}
	return insp, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateInspectionRequest) (db.Inspection, error) {
	insp, err := s.Ensure(ctx, id)
	if err != nil {
		return db.Inspection{}, err
	}
	if insp.IsFinalized {
		return db.Inspection{}, apperror.Conflict("Inspection is finalized; reopen it before modifying")
	}

	params := db.UpdateInspectionParams{
		ID:         id,
		ReviewerID: dbutil.NullUUID(req.ReviewerID),
		Notes:      dbutil.StrPtr(req.Notes),
		Metadata:   dbutil.JSONBytes(req.Metadata),
		Location:   dbutil.StrPtr(req.Location),
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
	}
	if req.Status != "" {
		v := db.InspectionStatus(req.Status)
		params.Status = &v
	}
	if req.Outcome != "" {
		v := db.InspectionOutcome(req.Outcome)
		params.Outcome = &v
	}

	return s.q.UpdateInspection(ctx, params)
}

// Finalize locks an inspection against further modification. Reviewers must
// explicitly Reopen it before any further edit (declarations, violations, or
// the inspection itself) is allowed.
func (s *Service) Finalize(ctx context.Context, id uuid.UUID, finalizedByID string) (db.Inspection, error) {
	if _, err := s.Ensure(ctx, id); err != nil {
		return db.Inspection{}, err
	}
	insp, err := s.q.FinalizeInspection(ctx, db.FinalizeInspectionParams{ID: id, FinalizedByID: dbutil.NullUUID(finalizedByID)})
	if err != nil {
		return db.Inspection{}, apperror.Conflict("Inspection is already finalized")
	}
	return insp, nil
}

// Reopen clears the finalized lock, explicitly re-opening the inspection for
// correction. Callers must audit-log this action.
func (s *Service) Reopen(ctx context.Context, id uuid.UUID) (db.Inspection, error) {
	if _, err := s.Ensure(ctx, id); err != nil {
		return db.Inspection{}, err
	}
	insp, err := s.q.ReopenInspection(ctx, id)
	if err != nil {
		return db.Inspection{}, apperror.Conflict("Inspection is not finalized")
	}
	return insp, nil
}

// EnsureNotFinalized returns a Conflict error if the inspection is finalized.
// Other modules (declarations, violations) call this before allowing an edit
// on data that belongs to a finalized inspection.
func (s *Service) EnsureNotFinalized(ctx context.Context, id uuid.UUID) error {
	insp, err := s.Ensure(ctx, id)
	if err != nil {
		return err
	}
	if insp.IsFinalized {
		return apperror.Conflict("Inspection is finalized; reopen it before modifying")
	}
	return nil
}
