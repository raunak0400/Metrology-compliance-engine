// Scan/upload orchestration.
//
// The API path here does the CHEAP work only: store the image in object
// storage, create the authoritative DB rows, and enqueue an Asynq task. All
// expensive OCR/ML work happens later in the workers. The API returns a job
// id immediately.
package scans

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/dbutil"
	"github.com/raunak0400/metrology-backend/internal/inspections"
	"github.com/raunak0400/metrology-backend/internal/taskqueue"
	"github.com/raunak0400/metrology-backend/pkg/storage"
)

type Service struct {
	q       *db.Queries
	storage storage.ObjectStorage
	asynq   *asynq.Client
	inspSvc *inspections.Service
}

func NewService(q *db.Queries, store storage.ObjectStorage, asynqClient *asynq.Client, inspSvc *inspections.Service) *Service {
	return &Service{q: q, storage: store, asynq: asynqClient, inspSvc: inspSvc}
}

type UploadResult struct {
	Image db.InspectionImage `json:"image"`
	JobID string             `json:"jobId"`
}

// AddImage stores an uploaded image and enqueues it for processing.
func (s *Service) AddImage(ctx context.Context, inspectionID uuid.UUID, buf []byte, filename, contentType string) (UploadResult, error) {
	if _, err := s.inspSvc.Ensure(ctx, inspectionID); err != nil {
		return UploadResult{}, err
	}

	key := storage.BuildObjectKey("inspections/"+inspectionID.String()+"/original", filename)
	if err := s.storage.Put(ctx, key, buf, contentType); err != nil {
		return UploadResult{}, err
	}

	checksum := storage.SHA256Hex(buf)
	size := int32(len(buf))
	image, err := s.q.CreateInspectionImage(ctx, db.CreateInspectionImageParams{
		InspectionID:     inspectionID,
		StorageKey:       key,
		MimeType:         contentType,
		OriginalFilename: dbutil.StrPtr(filename),
		SizeBytes:        &size,
		Checksum:         &checksum,
	})
	if err != nil {
		return UploadResult{}, err
	}

	jobID, err := s.enqueueImage(ctx, inspectionID, image.ID)
	if err != nil {
		return UploadResult{}, err
	}

	_ = s.q.UpdateInspectionImageStatus(ctx, db.UpdateInspectionImageStatusParams{
		ID: image.ID, Status: db.ImageProcessingStatusQUEUED,
	})
	image.Status = db.ImageProcessingStatusQUEUED

	return UploadResult{Image: image, JobID: jobID}, nil
}

// ProcessInspection (re)processes every original image on an inspection.
func (s *Service) ProcessInspection(ctx context.Context, inspectionID uuid.UUID) (int, error) {
	if _, err := s.inspSvc.Ensure(ctx, inspectionID); err != nil {
		return 0, err
	}

	imageIDs, err := s.q.ListOriginalImagesByInspection(ctx, inspectionID)
	if err != nil {
		return 0, err
	}

	for _, imgID := range imageIDs {
		if _, err := s.enqueueImage(ctx, inspectionID, imgID); err != nil {
			return 0, err
		}
	}

	_ = s.q.SetInspectionProcessing(ctx, inspectionID)
	return len(imageIDs), nil
}

func (s *Service) ListJobs(ctx context.Context, inspectionID uuid.UUID) ([]db.MlJob, error) {
	if _, err := s.inspSvc.Ensure(ctx, inspectionID); err != nil {
		return nil, err
	}
	return s.q.ListMLJobsByInspection(ctx, inspectionID)
}

// enqueueImage creates an MLJob row + Asynq task for a single image.
func (s *Service) enqueueImage(ctx context.Context, inspectionID, imageID uuid.UUID) (string, error) {
	queueName := taskqueue.QueueScan
	mlJob, err := s.q.CreateMLJob(ctx, db.CreateMLJobParams{
		InspectionID: inspectionID,
		Type:         db.MlJobTypeANALYZE,
		QueueName:    &queueName,
		ImageID:      uuid.NullUUID{UUID: imageID, Valid: true},
	})
	if err != nil {
		return "", err
	}

	payload, err := json.Marshal(taskqueue.ScanPayload{
		InspectionID: inspectionID.String(), ImageID: imageID.String(), MLJobID: mlJob.ID.String(),
	})
	if err != nil {
		return "", err
	}

	task := asynq.NewTask(taskqueue.TaskScanProcess, payload)
	info, err := s.asynq.Enqueue(task, asynq.Queue(taskqueue.QueueScan), asynq.TaskID(mlJob.ID.String()), asynq.MaxRetry(3))
	if err != nil {
		return "", err
	}

	_ = s.q.SetMLJobAsynqID(ctx, db.SetMLJobAsynqIDParams{ID: mlJob.ID, BullJobID: &info.ID})
	return mlJob.ID.String(), nil
}
