// Report management. Generation is ASYNC: the API creates a Report row and
// enqueues a task; the report worker collects the data, renders the file,
// uploads it to object storage and marks the report READY. Downloads are
// served via a short-lived signed URL.
package reports

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/raunak0400/metrology-backend/internal/apperror"
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

type RequestResult struct {
	ReportID string `json:"reportId"`
	JobID    string `json:"jobId"`
	Status   string `json:"status"`
}

func (s *Service) Request(ctx context.Context, inspectionID uuid.UUID, reportType string, requestedByID string) (RequestResult, error) {
	if _, err := s.inspSvc.Ensure(ctx, inspectionID); err != nil {
		return RequestResult{}, err
	}
	if reportType == "" {
		reportType = "PDF"
	}

	report, err := s.q.CreateReport(ctx, db.CreateReportParams{
		InspectionID: inspectionID, Type: db.ReportType(reportType), GeneratedByID: dbutil.NullUUID(requestedByID),
	})
	if err != nil {
		return RequestResult{}, err
	}

	payload, err := json.Marshal(taskqueue.ReportPayload{
		InspectionID: inspectionID.String(), ReportID: report.ID.String(), Type: reportType, RequestedByUserID: requestedByID,
	})
	if err != nil {
		return RequestResult{}, err
	}
	info, err := s.asynq.Enqueue(asynq.NewTask(taskqueue.TaskReportGenerate, payload), asynq.Queue(taskqueue.QueueReport))
	if err != nil {
		return RequestResult{}, err
	}

	return RequestResult{ReportID: report.ID.String(), JobID: info.ID, Status: "PENDING"}, nil
}

func (s *Service) ListByInspection(ctx context.Context, inspectionID uuid.UUID) ([]db.Report, error) {
	if _, err := s.inspSvc.Ensure(ctx, inspectionID); err != nil {
		return nil, err
	}
	return s.q.ListReportsByInspection(ctx, inspectionID)
}

type ReportView struct {
	db.Report
	DownloadURL *string `json:"downloadUrl"`
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (ReportView, error) {
	report, err := s.q.GetReport(ctx, id)
	if err != nil {
		return ReportView{}, apperror.NotFound("Report")
	}
	view := ReportView{Report: report}
	if report.Status == db.ReportStatusREADY && report.StorageKey != nil {
		if url, err := s.storage.SignedDownloadURL(ctx, *report.StorageKey, 15*time.Minute); err == nil {
			view.DownloadURL = &url
		}
	}
	return view, nil
}

// ---- Worker-side helpers ----------------------------------------------------

func (s *Service) MarkGenerating(ctx context.Context, reportID uuid.UUID) error {
	return s.q.MarkReportGenerating(ctx, reportID)
}

func (s *Service) MarkReady(ctx context.Context, reportID uuid.UUID, storageKey, fileName string, sizeBytes int32, checksum string) error {
	return s.q.MarkReportReady(ctx, db.MarkReportReadyParams{
		ID: reportID, StorageKey: &storageKey, FileName: &fileName, SizeBytes: &sizeBytes, Checksum: &checksum,
	})
}

func (s *Service) MarkFailed(ctx context.Context, reportID uuid.UUID, errMsg string) error {
	return s.q.MarkReportFailed(ctx, db.MarkReportFailedParams{ID: reportID, Error: &errMsg})
}

// CollectData gathers everything the report renderer needs from PostgreSQL.
func (s *Service) CollectData(ctx context.Context, inspectionID uuid.UUID) (ReportData, error) {
	insp, err := s.q.GetInspectionForReport(ctx, inspectionID)
	if err != nil {
		return ReportData{}, apperror.NotFound("Inspection")
	}

	declarations, err := s.q.ListDeclarationsByInspection(ctx, inspectionID)
	if err != nil {
		return ReportData{}, err
	}
	checks, err := s.q.ListComplianceChecksByInspection(ctx, inspectionID)
	if err != nil {
		return ReportData{}, err
	}
	violations, err := s.q.ListViolationsByInspection(ctx, inspectionID)
	if err != nil {
		return ReportData{}, err
	}

	data := ReportData{
		Code: insp.Code, Outcome: string(insp.Outcome), Status: string(insp.Status),
		ProductName: dbutil.StrVal(insp.ProductName),
	}
	for _, d := range declarations {
		conf := ""
		if d.Confidence != nil {
			conf = formatFloat(*d.Confidence)
		}
		data.Declarations = append(data.Declarations, ReportDeclaration{
			Type: string(d.Type), NormalizedValue: dbutil.StrVal(d.NormalizedValue), Confidence: conf,
		})
	}
	for _, v := range violations {
		data.Violations = append(data.Violations, ReportViolation{
			Title: v.Title, Severity: string(v.Severity), LegalReference: dbutil.StrVal(v.LegalReference), Reason: dbutil.StrVal(v.Reason),
		})
	}
	for _, c := range checks {
		data.Checks = append(data.Checks, ReportCheck{RuleCode: c.RuleCode, Status: string(c.Status), Reason: dbutil.StrVal(c.Reason)})
	}
	return data, nil
}

func formatFloat(f float64) string {
	b, _ := json.Marshal(f)
	return string(b)
}
