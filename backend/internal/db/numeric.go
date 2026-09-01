package db

import "github.com/jackc/pgx/v5/pgtype"

// NumericToFloat64 converts a possibly-NULL pgtype.Numeric column into a Go
// *float64 for use in API responses and the compliance engine, which only
// needs standard float precision (mirrors the original Node service's
// `Number(decimal)` conversion).
func NumericToFloat64(n pgtype.Numeric) *float64 {
	if !n.Valid {
		return nil
	}
	f, err := n.Float64Value()
	if err != nil || !f.Valid {
		return nil
	}
	return &f.Float64
}

// Float64ToNumeric converts an optional float64 into a pgtype.Numeric query
// parameter (nil -> SQL NULL).
func Float64ToNumeric(v *float64) pgtype.Numeric {
	if v == nil {
		return pgtype.Numeric{Valid: false}
	}
	var n pgtype.Numeric
	if err := n.Scan(*v); err != nil {
		return pgtype.Numeric{Valid: false}
	}
	return n
}
