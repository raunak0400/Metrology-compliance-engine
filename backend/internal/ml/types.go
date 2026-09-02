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

// ScaleReferenceMM is the pixel<->real-world-millimetre calibration the ML
// service may detect (e.g. from a reference card/ruler in frame). Mirrors
// internal/compliance.ScaleReference exactly, since it round-trips through
// InspectionImage.scale_reference (JSONB) unchanged — see scans/pipeline.go.
// A nil/zero value means no valid scale reference was found; the measurement
// package (internal/compliance/measurement.go) never assumes pixels == mm.
type ScaleReferenceMM struct {
	PxPerMm           *float64 `json:"pxPerMm,omitempty"`
	ReferenceObjectMm *float64 `json:"referenceObjectMm,omitempty"`
	ReferenceObjectPx *float64 `json:"referenceObjectPx,omitempty"`
	Source            *string  `json:"source,omitempty"`
}

// AnalyzeResult is the consolidated ML analysis JSON contract: the single
// call the scan pipeline makes per image (POST /ml/analyze). OCR/Regions/
// Fields are consumed directly (OCR -> MongoDB evidence trail, Fields ->
// normalized declarations via internal/declarations). The remaining fields
// are the broader per-package analysis the ML team may populate over time;
// they are stored verbatim in MongoDB (see internal/ml.Repository.SaveMLResult)
// as the flexible analysis document and are NOT forced into PostgreSQL —
// only the legally relevant, normalized subset the pipeline explicitly
// extracts (Fields, ScaleReferenceMM) reaches PostgreSQL. Never contains
// image_base64 — the image itself lives in S3/MinIO, referenced by
// ImageRef.StorageKey/ImageURL.
type AnalyzeResult struct {
	InspectionID string           `json:"inspectionId"`
	ImageID      string           `json:"imageId"`
	ModelVersion string           `json:"modelVersion"`
	OCR          OcrResult        `json:"ocr"`
	Regions      []DetectedRegion `json:"regions"`
	Fields       []ExtractedField `json:"fields"`
	CreatedAt    string           `json:"createdAt"`

	// Consolidated per-package analysis (see package doc). Flexible/nested
	// parts use map[string]interface{} deliberately — the ML team owns their
	// shape, and this backend must not need a code change every time a new
	// nested key is added; the whole document is preserved as-is in MongoDB.
	ProductID           *string                `json:"productId,omitempty"`
	Category            *string                `json:"category,omitempty"`
	SaleChannel         *string                `json:"saleChannel,omitempty"`
	NetWeightKg         *float64               `json:"netWeightKg,omitempty"`
	IsImported          *bool                  `json:"isImported,omitempty"`
	PackageShape        *string                `json:"packageShape,omitempty"`
	PackageSurfaceType  *string                `json:"packageSurfaceType,omitempty"`
	ScaleReferenceMM    *ScaleReferenceMM      `json:"scaleReferenceMm,omitempty"`
	PanelMetadata       map[string]interface{} `json:"panelMetadata,omitempty"`
	PanelAnalysis       map[string]interface{} `json:"panelAnalysis,omitempty"`
	TextOverlapAnalysis map[string]interface{} `json:"textOverlapAnalysis,omitempty"`
	ExtractedTags       []string               `json:"extractedTags,omitempty"`
	TextMeasurements    map[string]interface{} `json:"textMeasurements,omitempty"`
	ExtraFields         map[string]interface{} `json:"extraFields,omitempty"`
}

type HealthResult struct {
	Status       string `json:"status"`
	ModelVersion string `json:"modelVersion"`
}
