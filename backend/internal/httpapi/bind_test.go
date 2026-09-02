package httpapi

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/raunak0400/metrology-backend/internal/apperror"
)

func init() { gin.SetMode(gin.TestMode) }

type sampleBody struct {
	Email string `json:"email" binding:"required,email"`
	Name  string `json:"name" binding:"required,min=2,max=10"`
}

func ginContextWithJSON(body string) *gin.Context {
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = req
	return c
}

func TestBindJSON_MissingRequiredFields_ReturnsValidationErrorWithDetails(t *testing.T) {
	c := ginContextWithJSON(`{}`)
	var dst sampleBody
	err := BindJSON(c, &dst)
	if err == nil {
		t.Fatal("expected a validation error for missing required fields")
	}

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected an *apperror.AppError, got %T", err)
	}
	if appErr.ErrCode != apperror.CodeValidation {
		t.Fatalf("expected CodeValidation, got %s", appErr.ErrCode)
	}
	if len(appErr.Details) != 2 {
		t.Fatalf("expected 2 field-level details (email, name), got %d: %+v", len(appErr.Details), appErr.Details)
	}
	for _, d := range appErr.Details {
		if d.Field != "email" && d.Field != "name" {
			t.Errorf("unexpected field in details: %s", d.Field)
		}
		if d.Code != "required" {
			t.Errorf("expected code=required, got %s", d.Code)
		}
	}
}

func TestBindJSON_InvalidEmail_ReturnsFieldSpecificDetail(t *testing.T) {
	c := ginContextWithJSON(`{"email":"not-an-email","name":"ok"}`)
	var dst sampleBody
	err := BindJSON(c, &dst)
	if err == nil {
		t.Fatal("expected a validation error for an invalid email")
	}

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected an *apperror.AppError, got %T", err)
	}
	if len(appErr.Details) != 1 || appErr.Details[0].Field != "email" || appErr.Details[0].Code != "email" {
		t.Fatalf("expected exactly one email-tag detail on field 'email', got %+v", appErr.Details)
	}
}

func TestBindJSON_ValidBody_Succeeds(t *testing.T) {
	c := ginContextWithJSON(`{"email":"user@example.test","name":"Rohan"}`)
	var dst sampleBody
	if err := BindJSON(c, &dst); err != nil {
		t.Fatalf("expected no error for a valid body, got %v", err)
	}
	if dst.Email != "user@example.test" || dst.Name != "Rohan" {
		t.Fatalf("expected fields to be populated, got %+v", dst)
	}
}

func TestParseUUIDParam_Invalid_ReturnsValidationError(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Params = gin.Params{{Key: "id", Value: "not-a-uuid"}}

	_, err := ParseUUIDParam(c, "id")
	if err == nil {
		t.Fatal("expected an error for a malformed UUID path param")
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.ErrCode != apperror.CodeValidation {
		t.Fatalf("expected a validation AppError, got %v", err)
	}
}

func TestParseUUIDParam_Valid_ReturnsParsedUUID(t *testing.T) {
	want := uuid.New()
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Params = gin.Params{{Key: "id", Value: want.String()}}

	got, err := ParseUUIDParam(c, "id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}
