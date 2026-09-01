package evidence

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/dbutil"
	"github.com/raunak0400/metrology-backend/internal/inspections"
	"github.com/raunak0400/metrology-backend/pkg/storage"
)

type Service struct {
	q       *db.Queries
	storage storage.ObjectStorage
	inspSvc *inspections.Service
}

func NewService(q *db.Queries, store storage.ObjectStorage, inspSvc *inspections.Service) *Service {
	return &Service{q: q, storage: store, inspSvc: inspSvc}
}

type ViewWithURL struct {
	db.Evidence
	DownloadURL *string `json:"downloadUrl"`
}

// Add attaches evidence to an inspection. Evidence may be an uploaded file
// (stored in object storage) and/or a reference to an existing inspection
// image + bbox.
func (s *Service) Add(ctx context.Context, inspectionID uuid.UUID, req AddEvidenceRequest, createdByID string, fileBuf []byte, filename, contentType string) (db.Evidence, error) {
	if _, err := s.inspSvc.Ensure(ctx, inspectionID); err != nil {
		return db.Evidence{}, err
	}

	var storageKey *string
	if fileBuf != nil {
		key := storage.BuildObjectKey("inspections/"+inspectionID.String()+"/evidence", filename)
		if err := s.storage.Put(ctx, key, fileBuf, contentType); err != nil {
			return db.Evidence{}, err
		}
		storageKey = &key
	}

	evType := db.EvidenceTypeORIGINALIMAGE
	if req.Type != "" {
		evType = db.EvidenceType(req.Type)
	}

	return s.q.CreateEvidence(ctx, db.CreateEvidenceParams{
		InspectionID: inspectionID,
		ViolationID:  dbutil.NullUUID(req.ViolationID),
		ImageID:      dbutil.NullUUID(req.ImageID),
		Type:         evType,
		StorageKey:   storageKey,
		Bbox:         dbutil.JSONBytes(req.BBox),
		Caption:      dbutil.StrPtr(req.Caption),
		Metadata:     dbutil.JSONBytes(req.Metadata),
		CreatedByID:  dbutil.NullUUID(createdByID),
	})
}

// ListByInspection attaches short-lived signed download URLs for any stored artifacts.
func (s *Service) ListByInspection(ctx context.Context, inspectionID uuid.UUID) ([]ViewWithURL, error) {
	if _, err := s.inspSvc.Ensure(ctx, inspectionID); err != nil {
		return nil, err
	}
	items, err := s.q.ListEvidenceByInspection(ctx, inspectionID)
	if err != nil {
		return nil, err
	}

	out := make([]ViewWithURL, 0, len(items))
	for _, e := range items {
		v := ViewWithURL{Evidence: e}
		if e.StorageKey != nil {
			if url, err := s.storage.SignedDownloadURL(ctx, *e.StorageKey, 15*time.Minute); err == nil {
				v.DownloadURL = &url
			}
		}
		out = append(out, v)
	}
	return out, nil
}
