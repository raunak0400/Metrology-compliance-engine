package declarations

import (
	"context"

	"github.com/google/uuid"
	"github.com/raunak0400/metrology-backend/internal/apperror"
	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/dbutil"
	"github.com/raunak0400/metrology-backend/internal/inspections"
	"github.com/raunak0400/metrology-backend/internal/ml"
)

type Service struct {
	q       *db.Queries
	inspSvc *inspections.Service
}

func NewService(q *db.Queries, inspSvc *inspections.Service) *Service {
	return &Service{q: q, inspSvc: inspSvc}
}

func (s *Service) ListByInspection(ctx context.Context, inspectionID uuid.UUID) ([]db.ExtractedDeclaration, error) {
	if _, err := s.inspSvc.Ensure(ctx, inspectionID); err != nil {
		return nil, err
	}
	return s.q.ListDeclarationsByInspection(ctx, inspectionID)
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (db.ExtractedDeclaration, error) {
	d, err := s.q.GetDeclaration(ctx, id)
	if err != nil {
		return db.ExtractedDeclaration{}, apperror.NotFound("Declaration")
	}
	return d, nil
}

// Update is a reviewer edit. Records that a human corrected the value and by
// whom, without discarding the original raw OCR value (RawValue is untouched).
func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateDeclarationRequest, editedByID string) (db.ExtractedDeclaration, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return db.ExtractedDeclaration{}, err
	}

	params := db.UpdateDeclarationManualParams{
		ID:              id,
		NormalizedValue: dbutil.StrPtr(req.NormalizedValue),
		NumericValue:    db.Float64ToNumeric(req.NumericValue),
		Unit:            dbutil.StrPtr(req.Unit),
		EditedByID:      dbutil.NullUUID(editedByID),
	}
	if req.Type != "" {
		v := db.DeclarationType(req.Type)
		params.Type = &v
	}

	return s.q.UpdateDeclarationManual(ctx, params)
}

// PersistExtracted persists normalized declarations from raw ML
// field-extraction output. Called by the scan pipeline. Preserves rawValue,
// confidence and bbox as evidence and links each declaration back to its
// source image and the raw Mongo document.
func (s *Service) PersistExtracted(ctx context.Context, inspectionID, imageID uuid.UUID, fields []ml.ExtractedField, mongoRefID string) error {
	if len(fields) == 0 {
		return nil
	}

	// Idempotency: replace prior ML-sourced declarations for this image.
	if err := s.q.DeleteMLDeclarationsForImage(ctx, db.DeleteMLDeclarationsForImageParams{
		InspectionID: inspectionID, SourceImageID: uuid.NullUUID{UUID: imageID, Valid: true},
	}); err != nil {
		return err
	}

	for _, f := range fields {
		n := Normalize(f)
		var bbox []byte
		if f.BBox != nil {
			bbox = dbutil.JSONBytes(f.BBox)
		}
		confidence := f.Confidence
		if _, err := s.q.CreateDeclaration(ctx, db.CreateDeclarationParams{
			InspectionID:    inspectionID,
			SourceImageID:   uuid.NullUUID{UUID: imageID, Valid: true},
			Type:            db.DeclarationType(n.Type),
			Source:          db.DeclarationSourceMLEXTRACTION,
			NormalizedValue: n.NormalizedValue,
			RawValue:        dbutil.StrPtr(f.RawValue),
			Unit:            n.Unit,
			NumericValue:    db.Float64ToNumeric(n.NumericValue),
			Confidence:      &confidence,
			Bbox:            bbox,
			MongoRefID:      dbutil.StrPtr(mongoRefID),
		}); err != nil {
			return err
		}
	}
	return nil
}
