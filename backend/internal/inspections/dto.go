package inspections

type CreateInspectionRequest struct {
	ProductID  string                 `json:"productId" binding:"omitempty,uuid4"`
	CategoryID string                 `json:"categoryId" binding:"omitempty,uuid4"`
	Source     string                 `json:"source" binding:"omitempty,oneof=FIELD MARKETPLACE LAB OTHER"`
	Notes      string                 `json:"notes" binding:"omitempty,max=2000"`
	Metadata   map[string]interface{} `json:"metadata"`
	Location   string                 `json:"location" binding:"omitempty,max=500"`
	Latitude   *float64               `json:"latitude" binding:"omitempty,gte=-90,lte=90"`
	Longitude  *float64               `json:"longitude" binding:"omitempty,gte=-180,lte=180"`
}

type UpdateInspectionRequest struct {
	Status     string                 `json:"status" binding:"omitempty,oneof=DRAFT QUEUED PROCESSING AWAITING_REVIEW COMPLETED FAILED"`
	Outcome    string                 `json:"outcome" binding:"omitempty,oneof=PENDING COMPLIANT NON_COMPLIANT MANUAL_REVIEW NOT_APPLICABLE"`
	ReviewerID string                 `json:"reviewerId" binding:"omitempty,uuid4"`
	Notes      string                 `json:"notes" binding:"omitempty,max=2000"`
	Metadata   map[string]interface{} `json:"metadata"`
	Location   string                 `json:"location" binding:"omitempty,max=500"`
	Latitude   *float64               `json:"latitude" binding:"omitempty,gte=-90,lte=90"`
	Longitude  *float64               `json:"longitude" binding:"omitempty,gte=-180,lte=180"`
}

type ListInspectionsQuery struct {
	Page        int    `form:"page"`
	Limit       int    `form:"limit"`
	Status      string `form:"status" binding:"omitempty,oneof=DRAFT QUEUED PROCESSING AWAITING_REVIEW COMPLETED FAILED"`
	Outcome     string `form:"outcome" binding:"omitempty,oneof=PENDING COMPLIANT NON_COMPLIANT MANUAL_REVIEW NOT_APPLICABLE"`
	InspectorID string `form:"inspectorId" binding:"omitempty,uuid4"`
	ProductID   string `form:"productId" binding:"omitempty,uuid4"`
}
