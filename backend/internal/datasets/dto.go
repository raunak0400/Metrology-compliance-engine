package datasets

type CreateDatasetRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=120"`
	Description string `json:"description" binding:"omitempty,max=1000"`
	Purpose     string `json:"purpose" binding:"omitempty,max=120"`
}

type AddItemRequest struct {
	StorageKey    string                 `json:"storageKey" binding:"required,min=1,max=500"`
	SourceImageID string                 `json:"sourceImageId" binding:"omitempty,uuid4"`
	MimeType      string                 `json:"mimeType" binding:"omitempty,max=80"`
	Split         string                 `json:"split" binding:"omitempty,oneof=UNASSIGNED TRAIN VALIDATION TEST"`
	Metadata      map[string]interface{} `json:"metadata"`
}

type AddAnnotationRequest struct {
	FieldLabel          string                 `json:"fieldLabel" binding:"omitempty,max=80"`
	ClassificationLabel string                 `json:"classificationLabel" binding:"omitempty,max=80"`
	BBox                map[string]interface{} `json:"bbox"`
	Value               string                 `json:"value" binding:"omitempty,max=1000"`
	Confidence          *float64               `json:"confidence" binding:"omitempty,gte=0,lte=1"`
	Status              string                 `json:"status" binding:"omitempty,oneof=PENDING IN_PROGRESS ANNOTATED REVIEWED REJECTED"`
	Metadata            map[string]interface{} `json:"metadata"`
}

type ListQuery struct {
	Page             int    `form:"page"`
	Limit            int    `form:"limit"`
	Split            string `form:"split" binding:"omitempty,oneof=UNASSIGNED TRAIN VALIDATION TEST"`
	AnnotationStatus string `form:"annotationStatus" binding:"omitempty,oneof=PENDING IN_PROGRESS ANNOTATED REVIEWED REJECTED"`
}
