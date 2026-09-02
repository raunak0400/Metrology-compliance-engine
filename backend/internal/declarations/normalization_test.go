package declarations

import (
	"testing"

	"github.com/raunak0400/metrology-backend/internal/ml"
)

func field(name, rawValue string) ml.ExtractedField {
	return ml.ExtractedField{Field: name, RawValue: rawValue, Confidence: 0.9}
}

func TestNormalize_ParsesMRPIntoNumericValue(t *testing.T) {
	n := Normalize(field("MRP", "M.R.P. Rs. 199.00"))
	if n.Type != "MRP" {
		t.Fatalf("expected type MRP, got %s", n.Type)
	}
	if n.NumericValue == nil || *n.NumericValue != 199 {
		t.Fatalf("expected numeric value 199, got %v", n.NumericValue)
	}
	if n.NormalizedValue == nil || *n.NormalizedValue != "199.00" {
		t.Fatalf("expected normalized value 199.00, got %v", n.NormalizedValue)
	}
}

func TestNormalize_ParsesNetQuantityValueAndUnit(t *testing.T) {
	n := Normalize(field("NET_QUANTITY", "Net Qty: 500 g"))
	if n.Type != "NET_QUANTITY" {
		t.Fatalf("expected type NET_QUANTITY, got %s", n.Type)
	}
	if n.NumericValue == nil || *n.NumericValue != 500 {
		t.Fatalf("expected numeric value 500, got %v", n.NumericValue)
	}
	if n.Unit == nil || *n.Unit != "g" {
		t.Fatalf("expected unit g, got %v", n.Unit)
	}
}

func TestNormalize_MapsUnknownFieldsToOther(t *testing.T) {
	n := Normalize(field("SOMETHING_ELSE", "hello world"))
	if n.Type != "OTHER" {
		t.Fatalf("expected type OTHER, got %s", n.Type)
	}
	if n.NormalizedValue == nil || *n.NormalizedValue != "hello world" {
		t.Fatalf("expected raw string preserved, got %v", n.NormalizedValue)
	}
}

func TestNormalize_BestBeforeAndUseByDateAreDistinctFromExpiryDate(t *testing.T) {
	for _, typ := range []string{"BEST_BEFORE", "USE_BY_DATE", "EXPIRY_DATE"} {
		n := Normalize(field(typ, "12/2027"))
		if n.Type != typ {
			t.Fatalf("expected type %s to round-trip, got %s", typ, n.Type)
		}
		if n.NormalizedValue == nil || *n.NormalizedValue != "12/2027" {
			t.Fatalf("expected normalized date value preserved for %s, got %v", typ, n.NormalizedValue)
		}
	}
}
