// Scan processing pipeline (worker-side orchestration).
//
// Enforces the mandated separation:
//
//	IMAGE -> (Python ML: preprocess/detect/OCR/extract) -> RAW ML RESULT (MongoDB)
//	      -> NORMALIZATION -> STRUCTURED DECLARATIONS (PostgreSQL)
//	      -> enqueue DETERMINISTIC compliance evaluation
//
// Go never runs ML here — it calls the ML service, stores raw output in
// Mongo, normalizes into PostgreSQL, then hands off to the compliance queue.
package scans

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/declarations"
	"github.com/raunak0400/metrology-backend/internal/ml"
	"github.com/raunak0400/metrology-backend/internal/taskqueue"
	"github.com/raunak0400/metrology-backend/pkg/storage"
)

// Pipeline holds the dependencies needed to process one scan task. It is
// constructed once by the worker process and reused across tasks.
type Pipeline struct {
	q        *db.Queries
	storage  storage.ObjectStorage
	mlClient *ml.Client
	mlRepo   *ml.Repository
	declSvc  *declarations.Service
	asynq    *asynq.Client
}

func NewPipeline(
	q *db.Queries,
	store storage.ObjectStorage,
	mlClient *ml.Client,
	mlRepo *ml.Repository,
	declSvc *declarations.Service,
	asynqClient *asynq.Client,
) *Pipeline {
	return &Pipeline{q: q, storage: store, mlClient: mlClient, mlRepo: mlRepo, declSvc: declSvc, asynq: asynqClient}
}

func (p *Pipeline) Process(ctx context.Context, payload taskqueue.ScanPayload) error {
	inspectionID, err := uuid.Parse(payload.InspectionID)
	if err != nil {
		return fmt.Errorf("invalid inspectionId: %w", err)
	}
	imageID, err := uuid.Parse(payload.ImageID)
	if err != nil {
		return fmt.Errorf("invalid imageId: %w", err)
	}
	mlJobID, err := uuid.Parse(payload.MLJobID)
	if err != nil {
		return fmt.Errorf("invalid mlJobId: %w", err)
	}

	log := slog.With("inspectionId", inspectionID, "imageId", imageID, "mlJobId", mlJobID)

	image, err := p.q.GetInspectionImage(ctx, imageID)
	if err != nil {
		return fmt.Errorf("inspection image %s not found: %w", imageID, err)
	}

	_ = p.q.SetMLJobRunning(ctx, mlJobID)
	_ = p.q.UpdateInspectionImageStatus(ctx, db.UpdateInspectionImageStatusParams{ID: imageID, Status: db.ImageProcessingStatusPROCESSING})

	if err := p.run(ctx, inspectionID, imageID, mlJobID, image, log); err != nil {
		log.Error("scan pipeline failed", "err", err)
		_ = p.q.UpdateInspectionImageStatus(ctx, db.UpdateInspectionImageStatusParams{ID: imageID, Status: db.ImageProcessingStatusFAILED})
		errMsg := err.Error()
		_ = p.q.SetMLJobFailed(ctx, db.SetMLJobFailedParams{ID: mlJobID, Error: &errMsg})
		return err // let Asynq apply retry/backoff
	}

	log.Info("scan pipeline completed; compliance evaluation queued")
	return nil
}

func (p *Pipeline) run(ctx context.Context, inspectionID, imageID, mlJobID uuid.UUID, image db.InspectionImage, log *slog.Logger) error {
	// The ML service downloads the image via a short-lived signed URL.
	imageURL, err := p.storage.SignedDownloadURL(ctx, image.StorageKey, 15*time.Minute)
	if err != nil {
		return fmt.Errorf("sign download url: %w", err)
	}
	ref := ml.ImageRef{InspectionID: inspectionID.String(), ImageID: imageID.String(), StorageKey: image.StorageKey, ImageURL: imageURL}

	// 1) Call the Python ML service (preprocess + detect + OCR + extract).
	analysis, err := p.mlClient.Analyze(ctx, ref)
	if err != nil {
		return fmt.Errorf("ml analyze: %w", err)
	}

	// 2) Persist RAW ML output to MongoDB (evidence trail, not source of truth).
	ocrDocID, err := p.mlRepo.SaveOcrDocument(ctx, analysis.OCR)
	if err != nil {
		return fmt.Errorf("save ocr document: %w", err)
	}
	analyzeDocID, err := p.mlRepo.SaveMLResult(ctx, "ANALYZE", inspectionID.String(), imageID.String(), analysis.ModelVersion, analysis)
	if err != nil {
		return fmt.Errorf("save ml result: %w", err)
	}

	// 3) Normalize fields into authoritative PostgreSQL declarations.
	if err := p.declSvc.PersistExtracted(ctx, inspectionID, imageID, analysis.Fields, analyzeDocID); err != nil {
		return fmt.Errorf("persist declarations: %w", err)
	}

	// 3b) Persist the scale reference (if the ML service found one) so the
	// measurement package can convert declaration bbox heights to real-world
	// mm. Without this, MIN_TEXT_HEIGHT_MM checks always route to manual
	// review — never assumed, only computed from an actual calibration.
	if analysis.ScaleReferenceMM != nil {
		scaleJSON, err := json.Marshal(analysis.ScaleReferenceMM)
		if err != nil {
			return fmt.Errorf("marshal scale reference: %w", err)
		}
		if err := p.q.SetInspectionImageScaleReference(ctx, db.SetInspectionImageScaleReferenceParams{ID: imageID, ScaleReference: scaleJSON}); err != nil {
			return fmt.Errorf("persist scale reference: %w", err)
		}
	}

	// 4) Update authoritative rows with references to the raw Mongo documents.
	if err := p.q.MarkInspectionImageProcessed(ctx, db.MarkInspectionImageProcessedParams{ID: imageID, MongoOcrDocID: &ocrDocID}); err != nil {
		return fmt.Errorf("mark image processed: %w", err)
	}
	if err := p.q.SetMLJobSucceeded(ctx, db.SetMLJobSucceededParams{ID: mlJobID, ResultMongoID: &analyzeDocID}); err != nil {
		return fmt.Errorf("mark ml job succeeded: %w", err)
	}

	// 5) Hand off to the DETERMINISTIC compliance engine (separate queue).
	compliancePayload, err := json.Marshal(taskqueue.CompliancePayload{InspectionID: inspectionID.String()})
	if err != nil {
		return fmt.Errorf("marshal compliance payload: %w", err)
	}
	if _, err := p.asynq.Enqueue(asynq.NewTask(taskqueue.TaskComplianceCheck, compliancePayload), asynq.Queue(taskqueue.QueueCompliance)); err != nil {
		return fmt.Errorf("enqueue compliance check: %w", err)
	}

	return nil
}
