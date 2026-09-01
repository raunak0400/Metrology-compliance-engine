// Package workers hosts the Asynq worker server: it registers one handler per
// task type and dispatches to the relevant domain service. This is the Go
// equivalent of the original BullMQ workers (scan/ocr/compliance/report) —
// each handler here does exactly what its Node counterpart did, just backed
// by Asynq instead of BullMQ.
package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/raunak0400/metrology-backend/internal/compliance"
	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/ml"
	"github.com/raunak0400/metrology-backend/internal/reports"
	"github.com/raunak0400/metrology-backend/internal/scans"
	"github.com/raunak0400/metrology-backend/internal/taskqueue"
	"github.com/raunak0400/metrology-backend/pkg/storage"
)

// Deps bundles everything the worker handlers need.
type Deps struct {
	Queries       *db.Queries
	Storage       storage.ObjectStorage
	ScanPipeline  *scans.Pipeline
	ComplianceSvc *compliance.Service
	ReportSvc     *reports.Service
	MLClient      *ml.Client
	MLRepo        *ml.Repository
}

// NewMux builds the Asynq handler mux with one handler per task type.
func NewMux(d Deps) *asynq.ServeMux {
	mux := asynq.NewServeMux()

	mux.HandleFunc(taskqueue.TaskScanProcess, func(ctx context.Context, t *asynq.Task) error {
		var payload taskqueue.ScanPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return fmt.Errorf("unmarshal scan payload: %w", err)
		}
		slog.Info("processing scan task", "inspectionId", payload.InspectionID, "imageId", payload.ImageID)
		return d.ScanPipeline.Process(ctx, payload)
	})

	mux.HandleFunc(taskqueue.TaskOCRProcess, func(ctx context.Context, t *asynq.Task) error {
		var payload taskqueue.OCRPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return fmt.Errorf("unmarshal ocr payload: %w", err)
		}
		return d.processOCROnly(ctx, payload)
	})

	mux.HandleFunc(taskqueue.TaskComplianceCheck, func(ctx context.Context, t *asynq.Task) error {
		var payload taskqueue.CompliancePayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return fmt.Errorf("unmarshal compliance payload: %w", err)
		}
		inspectionID, err := uuid.Parse(payload.InspectionID)
		if err != nil {
			return fmt.Errorf("invalid inspectionId: %w", err)
		}
		slog.Info("running compliance evaluation", "inspectionId", inspectionID)
		summary, err := d.ComplianceSvc.Run(ctx, inspectionID)
		if err != nil {
			return err
		}
		slog.Info("compliance evaluation complete", "inspectionId", inspectionID, "outcome", summary.Outcome,
			"passed", summary.Passed, "failed", summary.Failed, "manualReview", summary.ManualReview)
		return nil
	})

	mux.HandleFunc(taskqueue.TaskReportGenerate, func(ctx context.Context, t *asynq.Task) error {
		var payload taskqueue.ReportPayload
		if err := json.Unmarshal(t.Payload(), &payload); err != nil {
			return fmt.Errorf("unmarshal report payload: %w", err)
		}
		return d.processReport(ctx, payload)
	})

	return mux
}

// processOCROnly is the standalone OCR-only step (e.g. re-OCR with a
// different model version). It stores raw OCR in MongoDB and updates the
// corresponding MLJob — it does NOT make compliance decisions.
func (d Deps) processOCROnly(ctx context.Context, payload taskqueue.OCRPayload) error {
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

	image, err := d.Queries.GetInspectionImage(ctx, imageID)
	if err != nil {
		return fmt.Errorf("inspection image %s not found: %w", imageID, err)
	}
	_ = d.Queries.SetMLJobRunning(ctx, mlJobID)

	imageURL, err := d.Storage.SignedDownloadURL(ctx, image.StorageKey, 15*time.Minute)
	if err != nil {
		return fmt.Errorf("sign download url: %w", err)
	}
	ref := ml.ImageRef{InspectionID: inspectionID.String(), ImageID: imageID.String(), StorageKey: image.StorageKey, ImageURL: imageURL}

	ocr, err := d.MLClient.OCR(ctx, ref)
	if err != nil {
		return fmt.Errorf("ml ocr: %w", err)
	}
	ocrDocID, err := d.MLRepo.SaveOcrDocument(ctx, *ocr)
	if err != nil {
		return fmt.Errorf("save ocr document: %w", err)
	}

	if err := d.Queries.MarkInspectionImageProcessed(ctx, db.MarkInspectionImageProcessedParams{ID: imageID, MongoOcrDocID: &ocrDocID}); err != nil {
		return err
	}
	return d.Queries.SetMLJobSucceeded(ctx, db.SetMLJobSucceededParams{ID: mlJobID, ResultMongoID: &ocrDocID})
}

// processReport collects the inspection data, renders the report, uploads it
// to object storage, and marks the Report row READY (or FAILED).
func (d Deps) processReport(ctx context.Context, payload taskqueue.ReportPayload) error {
	reportID, err := uuid.Parse(payload.ReportID)
	if err != nil {
		return fmt.Errorf("invalid reportId: %w", err)
	}
	inspectionID, err := uuid.Parse(payload.InspectionID)
	if err != nil {
		return fmt.Errorf("invalid inspectionId: %w", err)
	}

	slog.Info("generating report", "reportId", reportID, "inspectionId", inspectionID)

	if err := d.ReportSvc.MarkGenerating(ctx, reportID); err != nil {
		return err
	}

	data, err := d.ReportSvc.CollectData(ctx, inspectionID)
	if err != nil {
		_ = d.ReportSvc.MarkFailed(ctx, reportID, err.Error())
		return err
	}

	rendered := reports.Render(data, payload.Type)

	key := storage.BuildObjectKey(fmt.Sprintf("inspections/%s/reports", inspectionID), fmt.Sprintf("report-%s.%s", reportID, rendered.Extension))
	if err := d.Storage.Put(ctx, key, rendered.Buffer, rendered.ContentType); err != nil {
		_ = d.ReportSvc.MarkFailed(ctx, reportID, err.Error())
		return err
	}

	checksum := storage.SHA256Hex(rendered.Buffer)
	fileName := fmt.Sprintf("report-%s.%s", reportID, rendered.Extension)
	if err := d.ReportSvc.MarkReady(ctx, reportID, key, fileName, int32(len(rendered.Buffer)), checksum); err != nil {
		return err
	}

	slog.Info("report ready", "reportId", reportID, "key", key)
	return nil
}
