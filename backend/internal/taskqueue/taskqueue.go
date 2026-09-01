// Package taskqueue defines the Asynq task types and payload contracts shared
// by producers (HTTP handlers enqueuing work, in internal/scans, internal/reports)
// and internal/workers, which registers the handlers that consume them. Kept
// separate from internal/workers so producer packages don't have to import
// the worker/handler wiring (which itself imports the producer packages'
// domain services) — avoids an import cycle. The API layer MUST NOT run
// expensive OCR/ML/report work synchronously — it enqueues a task here and
// returns a job id immediately.
package taskqueue

const (
	TaskScanProcess     = "scan:process"
	TaskOCRProcess      = "ocr:process"
	TaskComplianceCheck = "compliance:check"
	TaskReportGenerate  = "report:generate"
)

// ScanPayload drives the combined ML pipeline (preprocess + detect + OCR +
// extract) for one inspection image.
type ScanPayload struct {
	InspectionID string `json:"inspectionId"`
	ImageID      string `json:"imageId"`
	MLJobID      string `json:"mlJobId"`
}

// OCRPayload drives the standalone OCR-only step (e.g. re-OCR with a
// different model version).
type OCRPayload struct {
	InspectionID string `json:"inspectionId"`
	ImageID      string `json:"imageId"`
	MLJobID      string `json:"mlJobId"`
}

// CompliancePayload triggers a deterministic compliance evaluation run.
type CompliancePayload struct {
	InspectionID      string `json:"inspectionId"`
	TriggeredByUserID string `json:"triggeredByUserId,omitempty"`
}

// ReportPayload drives asynchronous report generation.
type ReportPayload struct {
	InspectionID      string `json:"inspectionId"`
	ReportID          string `json:"reportId"`
	Type              string `json:"type"` // PDF | DOCX
	RequestedByUserID string `json:"requestedByUserId,omitempty"`
}

// Queue names (Asynq priority queues), mirroring the original BullMQ queue set.
const (
	QueueScan       = "scan-processing"
	QueueOCR        = "ocr-processing"
	QueueCompliance = "compliance-check"
	QueueReport     = "report-generation"
)
