package reports

type CreateReportRequest struct {
	Type string `json:"type" binding:"omitempty,oneof=PDF DOCX"`
}
