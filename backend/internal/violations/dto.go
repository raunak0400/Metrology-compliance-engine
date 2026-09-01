package violations

type UpdateViolationRequest struct {
	Status      string `json:"status" binding:"omitempty,oneof=OPEN CONFIRMED DISMISSED RESOLVED"`
	Severity    string `json:"severity" binding:"omitempty,oneof=LOW MEDIUM HIGH CRITICAL"`
	ReviewNotes string `json:"reviewNotes" binding:"omitempty,max=2000"`
}
