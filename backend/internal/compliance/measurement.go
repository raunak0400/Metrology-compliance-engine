// Measurement logic: deterministic geometry only, deliberately SEPARATE from
// OCR and from the legal rule engine (which now lives in the Python rule
// engine service). It estimates real-world sizes (e.g. declaration text
// height) from bounding boxes, but ONLY when a valid scale reference exists.
// Without one it reports hasValidScale=false: pixel-to-mm conversion is never
// assumed to be accurate. The result feeds the rule engine's
// font_height_mm/text_measurements inputs; the rule engine — not this
// package — decides whether a height is compliant.
package compliance

import "encoding/json"

// ScaleReference mirrors InspectionImage.scale_reference (JSONB, nullable).
type ScaleReference struct {
	PxPerMm           *float64 `json:"pxPerMm,omitempty"`
	ReferenceObjectMm *float64 `json:"referenceObjectMm,omitempty"`
	ReferenceObjectPx *float64 `json:"referenceObjectPx,omitempty"`
	Source            *string  `json:"source,omitempty"`
}

// BBox mirrors the bbox JSONB shape used across declarations/evidence.
type BBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// ResolvePxPerMm resolves a pixels-per-mm ratio from a scale reference, or nil
// if the reference is missing/invalid.
func ResolvePxPerMm(scale *ScaleReference) *float64 {
	if scale == nil {
		return nil
	}
	if scale.PxPerMm != nil && *scale.PxPerMm > 0 {
		return scale.PxPerMm
	}
	if scale.ReferenceObjectMm != nil && scale.ReferenceObjectPx != nil &&
		*scale.ReferenceObjectMm > 0 && *scale.ReferenceObjectPx > 0 {
		v := *scale.ReferenceObjectPx / *scale.ReferenceObjectMm
		return &v
	}
	return nil
}

// EstimateTextHeightMm estimates text height in mm from a bbox; returns
// hasValidScale=false and a nil height when no valid scale reference exists.
func EstimateTextHeightMm(bbox *BBox, scale *ScaleReference) (heightMm *float64, hasValidScale bool) {
	pxPerMm := ResolvePxPerMm(scale)
	if bbox == nil || pxPerMm == nil {
		return nil, false
	}
	mm := roundTo(bbox.Height / *pxPerMm, 3)
	return &mm, true
}

func roundTo(v float64, decimals int) float64 {
	mult := 1.0
	for i := 0; i < decimals; i++ {
		mult *= 10
	}
	return float64(int64(v*mult+0.5)) / mult
}

// DeclarationScaleRow is the minimal shape needed to compute a measurement
// for one declaration (see queries/declarations.sql: ListDeclarationsWithScaleByInspection).
type DeclarationScaleRow struct {
	DeclarationID  string
	Type           string
	BBoxJSON       []byte
	ScaleReference []byte
}

// DeclarationHeightsMM estimates real-world text height (mm) per declaration
// ID from each declaration's bbox and its source image's scale reference. A
// missing map entry, or a nil value, means no valid scale reference existed
// for that declaration — the caller (building a rule-engine request) omits
// font_height_mm rather than assuming pixels == mm.
func DeclarationHeightsMM(rows []DeclarationScaleRow) map[string]*float64 {
	out := make(map[string]*float64, len(rows))
	for _, r := range rows {
		var bbox *BBox
		if len(r.BBoxJSON) > 0 {
			var b BBox
			if err := json.Unmarshal(r.BBoxJSON, &b); err == nil {
				bbox = &b
			}
		}
		var scale *ScaleReference
		if len(r.ScaleReference) > 0 {
			var s ScaleReference
			if err := json.Unmarshal(r.ScaleReference, &s); err == nil {
				scale = &s
			}
		}
		heightMm, _ := EstimateTextHeightMm(bbox, scale)
		out[r.DeclarationID] = heightMm
	}
	return out
}
