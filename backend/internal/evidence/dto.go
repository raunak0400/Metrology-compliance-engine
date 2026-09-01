package evidence

type AddEvidenceRequest struct {
	Type        string                 `json:"type" binding:"omitempty,oneof=ORIGINAL_IMAGE CROPPED_PANEL ANNOTATED_IMAGE DOCUMENT OTHER"`
	ViolationID string                 `json:"violationId" binding:"omitempty,uuid4"`
	ImageID     string                 `json:"imageId" binding:"omitempty,uuid4"`
	BBox        map[string]interface{} `json:"bbox"`
	Caption     string                 `json:"caption" binding:"omitempty,max=500"`
	Metadata    map[string]interface{} `json:"metadata"`
}
