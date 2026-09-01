// Package dbutil holds tiny conversion helpers shared by every module's
// service layer when translating between HTTP DTOs and sqlc-generated
// params/rows (optional strings <-> *string, JSON maps <-> []byte, UUID
// strings <-> uuid.NullUUID).
package dbutil

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// StrPtr returns nil for an empty string, otherwise a pointer to it — matches
// sqlc's *string representation for nullable text columns.
func StrPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// StrVal safely dereferences a possibly-nil *string.
func StrVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// JSONBytes marshals a map/slice into []byte for a JSONB column, or nil when
// the input is empty/nil.
func JSONBytes(v interface{}) []byte {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

// NullUUID parses a UUID string into uuid.NullUUID; an empty or invalid
// string yields an invalid (SQL NULL) value rather than an error, matching
// how optional foreign keys are treated throughout the API.
func NullUUID(s string) uuid.NullUUID {
	if s == "" {
		return uuid.NullUUID{}
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{UUID: id, Valid: true}
}

// GenerateCode builds a random human-friendly reference code, e.g. INSP-2026-A1B2C3D4.
func GenerateCode(prefix string) string {
	year := time.Now().Year()
	rand := strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", ""))[:8]
	return prefix + "-" + strconv.Itoa(year) + "-" + rand
}
