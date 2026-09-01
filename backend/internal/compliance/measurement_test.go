package compliance

import "testing"

func TestEstimateTextHeightMm_NoScaleReference(t *testing.T) {
	heightMm, hasValidScale := EstimateTextHeightMm(&BBox{X: 0, Y: 0, Width: 100, Height: 30}, nil)
	if hasValidScale {
		t.Fatal("expected hasValidScale=false")
	}
	if heightMm != nil {
		t.Fatalf("expected nil height, got %v", *heightMm)
	}
}

func TestEstimateTextHeightMm_DirectPxPerMm(t *testing.T) {
	px := 10.0
	heightMm, hasValidScale := EstimateTextHeightMm(&BBox{X: 0, Y: 0, Width: 100, Height: 30}, &ScaleReference{PxPerMm: &px})
	if !hasValidScale {
		t.Fatal("expected hasValidScale=true")
	}
	if heightMm == nil || *heightMm != 3 {
		t.Fatalf("expected 3mm, got %v", heightMm)
	}
}

func TestResolvePxPerMm_FromReferenceObject(t *testing.T) {
	mm, px := 10.0, 100.0
	v := ResolvePxPerMm(&ScaleReference{ReferenceObjectMm: &mm, ReferenceObjectPx: &px})
	if v == nil || *v != 10 {
		t.Fatalf("expected 10, got %v", v)
	}
	if ResolvePxPerMm(&ScaleReference{}) != nil {
		t.Fatal("expected nil for empty scale reference")
	}
}
