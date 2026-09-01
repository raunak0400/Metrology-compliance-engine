// Dataset management for the product-label ML dataset. The backend does NOT
// train models — it provides dataset + annotation storage/APIs so the Python
// ML team can consume them. Binary images live in object storage; only keys
// are stored here.
package datasets

import (
	"context"
	"regexp"
	"strings"

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

var nonAlnumRe = regexp.MustCompile(`[^a-z0-9]+`)
var trimDashRe = regexp.MustCompile(`(^-|-$)`)

func slugify(name string) string {
	s := nonAlnumRe.ReplaceAllString(strings.ToLower(strings.TrimSpace(name)), "-")
	return trimDashRe.ReplaceAllString(s, "")
}

func (s *Service) Create(ctx context.Context, req CreateDatasetRequest, createdByID string) (db.Dataset, error) {
	return s.q.CreateDataset(ctx, db.CreateDatasetParams{
		Name: req.Name, Slug: slugify(req.Name), Description: dbutil.StrPtr(req.Description),
		Purpose: dbutil.StrPtr(req.Purpose), CreatedByID: dbutil.NullUUID(createdByID),
	})
}

func (s *Service) List(ctx context.Context, page pagination.Query) ([]db.ListDatasetsRow, int64, error) {
	items, err := s.q.ListDatasets(ctx, db.ListDatasetsParams{Limit: page.LimitInt32(), Offset: page.Offset()})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountDatasets(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *Service) ensureDataset(ctx context.Context, id uuid.UUID) error {
	if _, err := s.q.GetDataset(ctx, id); err != nil {
		return apperror.NotFound("Dataset")
	}
	return nil
}

func (s *Service) AddItem(ctx context.Context, datasetID uuid.UUID, req AddItemRequest) (db.DatasetItem, error) {
	if err := s.ensureDataset(ctx, datasetID); err != nil {
		return db.DatasetItem{}, err
	}
	var split *db.DatasetSplit
	if req.Split != "" {
		v := db.DatasetSplit(req.Split)
		split = &v
	}
	return s.q.CreateDatasetItem(ctx, db.CreateDatasetItemParams{
		DatasetID: datasetID, StorageKey: req.StorageKey, SourceImageID: dbutil.NullUUID(req.SourceImageID),
		MimeType: dbutil.StrPtr(req.MimeType), Split: split, Metadata: dbutil.JSONBytes(req.Metadata),
	})
}

func (s *Service) ListItems(ctx context.Context, datasetID uuid.UUID, q ListQuery, page pagination.Query) ([]db.ListDatasetItemsRow, int64, error) {
	if err := s.ensureDataset(ctx, datasetID); err != nil {
		return nil, 0, err
	}
	var split *db.DatasetSplit
	if q.Split != "" {
		v := db.DatasetSplit(q.Split)
		split = &v
	}
	var annotationStatus *db.AnnotationStatus
	if q.AnnotationStatus != "" {
		v := db.AnnotationStatus(q.AnnotationStatus)
		annotationStatus = &v
	}

	items, err := s.q.ListDatasetItems(ctx, db.ListDatasetItemsParams{
		DatasetID: datasetID, Limit: page.LimitInt32(), Offset: page.Offset(), Split: split, AnnotationStatus: annotationStatus,
	})
	if err != nil {
		return nil, 0, err
	}
	total, err := s.q.CountDatasetItems(ctx, db.CountDatasetItemsParams{DatasetID: datasetID, Split: split, AnnotationStatus: annotationStatus})
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *Service) AddAnnotation(ctx context.Context, itemID uuid.UUID, req AddAnnotationRequest, annotatorID string) (db.DatasetAnnotation, error) {
	if _, err := s.q.GetDatasetItem(ctx, itemID); err != nil {
		return db.DatasetAnnotation{}, apperror.NotFound("Dataset item")
	}

	var status *db.AnnotationStatus
	if req.Status != "" {
		v := db.AnnotationStatus(req.Status)
		status = &v
	}

	annotation, err := s.q.CreateDatasetAnnotation(ctx, db.CreateDatasetAnnotationParams{
		ItemID: itemID, AnnotatorID: dbutil.NullUUID(annotatorID), FieldLabel: dbutil.StrPtr(req.FieldLabel),
		ClassificationLabel: dbutil.StrPtr(req.ClassificationLabel), Bbox: dbutil.JSONBytes(req.BBox),
		Value: dbutil.StrPtr(req.Value), Confidence: req.Confidence, Status: status, Metadata: dbutil.JSONBytes(req.Metadata),
	})
	if err != nil {
		return db.DatasetAnnotation{}, err
	}

	// Reflect progress on the item.
	_ = s.q.UpdateDatasetItemAnnotationStatus(ctx, db.UpdateDatasetItemAnnotationStatusParams{
		ID: itemID, AnnotationStatus: db.AnnotationStatusANNOTATED,
	})
	return annotation, nil
}
