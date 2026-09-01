// Package ml is the ONLY place the Go backend talks to the internal Python ML
// service (OpenCV / PaddleOCR / custom models). Swapping the ML implementation
// must not require changes outside this package. The backend never contains ML
// logic — it only orchestrates and normalizes these structured results.
package ml

type BBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// ImageRef references an image the ML service should process.
type ImageRef struct {
	InspectionID string `json:"inspectionId"`
	ImageID      string `json:"imageId"`
	StorageKey   string `json:"storageKey"`
	// ImageURL is a short-lived signed URL the ML service can download from.
	ImageURL string `json:"imageUrl,omitempty"`
}

type OcrBlock struct {
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence"`
	BBox       BBox    `json:"bbox"`
}

type OcrResult struct {
	InspectionID  string                 `json:"inspectionId"`
	ImageID       string                 `json:"imageId"`
	ModelVersion  string                 `json:"modelVersion"`
	Blocks        []OcrBlock             `json:"blocks"`
	Preprocessing map[string]interface{} `json:"preprocessing,omitempty"`
	CreatedAt     string                 `json:"createdAt"`
}

type DetectedRegion struct {
	Type       string  `json:"type"` // e.g. PRINCIPAL_DISPLAY_PANEL, LABEL
	Confidence float64 `json:"confidence"`
	BBox       BBox    `json:"bbox"`
}

type DetectLabelResult struct {
	InspectionID string           `json:"inspectionId"`
	ImageID      string           `json:"imageId"`
	ModelVersion string           `json:"modelVersion"`
	Regions      []DetectedRegion `json:"regions"`
	CreatedAt    string           `json:"createdAt"`
}

// ExtractedField is a candidate declaration field emitted by NER/classification models.
type ExtractedField struct {
	Field      string  `json:"field"` // maps to a declaration type, e.g. MRP, NET_QUANTITY
	RawValue   string  `json:"rawValue"`
	Confidence float64 `json:"confidence"`
	BBox       *BBox   `json:"bbox,omitempty"`
}

type ExtractFieldsResult struct {
	InspectionID string           `json:"inspectionId"`
	ImageID      string           `json:"imageId"`
	ModelVersion string           `json:"modelVersion"`
	Fields       []ExtractedField `json:"fields"`
	CreatedAt    string           `json:"createdAt"`
}

type AnalyzeResult struct {
	InspectionID string           `json:"inspectionId"`
	ImageID      string           `json:"imageId"`
	ModelVersion string           `json:"modelVersion"`
	OCR          OcrResult        `json:"ocr"`
	Regions      []DetectedRegion `json:"regions"`
	Fields       []ExtractedField `json:"fields"`
	CreatedAt    string           `json:"createdAt"`
}

type HealthResult struct {
	Status       string `json:"status"`
	ModelVersion string `json:"modelVersion"`
}
