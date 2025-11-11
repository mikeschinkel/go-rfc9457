package test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/mikeschinkel/go-rfc9457"
)

// assertRFC9457ErrorEqual compares two Response structs field by field
func assertRFC9457ErrorEqual(t *testing.T, got, want *rfc9457.Response) {
	t.Helper()

	if got.Type != want.Type {
		t.Errorf("Type: got %v, want %v", got.Type, want.Type)
	}
	if got.Title != want.Title {
		t.Errorf("Title: got %v, want %v", got.Title, want.Title)
	}
	if got.Status != want.Status {
		t.Errorf("Status: got %v, want %v", got.Status, want.Status)
	}
	if got.Detail != want.Detail {
		t.Errorf("Detail: got %v, want %v", got.Detail, want.Detail)
	}
	if got.Instance != want.Instance {
		t.Errorf("Instance: got %v, want %v", got.Instance, want.Instance)
	}
}

func TestRFC9457Error_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		err  *rfc9457.Response
		want string
	}{
		{
			name: "basic_error_all_fields",
			err: &rfc9457.Response{
				Type:     rfc9457.InvalidParameterErrorType,
				Title:    "Invalid Parameter Type",
				Status:   422,
				Detail:   "Parameter 'id' expected type 'int' but received 'abc'",
				Instance: "/api/users/abc",
			},
			want: `{"type":"https://schema.xmlui.org/errors/test-server/api/validation/invalid-parameter-type","title":"Invalid Parameter Type","status":422,"detail":"Parameter 'id' expected type 'int' but received 'abc'","instance":"/api/users/abc"}`,
		},
		{
			name: "error_with_constraint",
			err: &rfc9457.Response{
				Type:     rfc9457.ConstraintViolationErrorType,
				Title:    "Constraint Violation",
				Status:   422,
				Detail:   "Parameter 'score' value 150 violates constraint range[0..100]",
				Instance: "/api/users/by-score/150",
			},
			want: `{"type":"https://schema.xmlui.org/errors/test-server/api/validation/constraint-violation","title":"Constraint Violation","status":422,"detail":"Parameter 'score' value 150 violates constraint range[0..100]","instance":"/api/users/by-score/150"}`,
		},
		{
			name: "minimal_error",
			err: &rfc9457.Response{
				Type:   rfc9457.InvalidParameterErrorType,
				Title:  "Invalid Parameter Type",
				Status: 422,
			},
			want: `{"type":"https://schema.xmlui.org/errors/test-server/api/validation/invalid-parameter-type","title":"Invalid Parameter Type","status":422}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.err)
			if err != nil {
				t.Fatalf("Marshal error: %v", err)
			}

			if string(got) != tt.want {
				t.Errorf("Marshal mismatch:\ngot:  %s\nwant: %s", string(got), tt.want)
			}
		})
	}
}

func TestRFC9457Error_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    *rfc9457.Response
		wantErr bool
	}{
		{
			name: "basic_error_roundtrip",
			json: `{"type":"https://schema.xmlui.org/errors/test-server/api/validation/invalid-parameter-type","title":"Invalid Parameter Type","status":422,"detail":"Parameter 'id' expected type 'int' but received 'abc'","instance":"/api/users/abc"}`,
			want: &rfc9457.Response{
				Type:     rfc9457.InvalidParameterErrorType,
				Title:    "Invalid Parameter Type",
				Status:   422,
				Detail:   "Parameter 'id' expected type 'int' but received 'abc'",
				Instance: "/api/users/abc",
			},
		},
		{
			name: "minimal_error",
			json: `{"type":"https://schema.xmlui.org/errors/test-server/api/validation/invalid-parameter-type","title":"Minimal Error","status":400}`,
			want: &rfc9457.Response{
				Type:   rfc9457.InvalidParameterErrorType,
				Title:  "Minimal Error",
				Status: 400,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got rfc9457.Response
			err := json.Unmarshal([]byte(tt.json), &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal error: %v, wantErr: %v", err, tt.wantErr)
			}

			assertRFC9457ErrorEqual(t, &got, tt.want)
		})
	}
}

func TestRFC9457Error_HTTPStatusCode(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{"status_422", 422},
		{"status_400", 400},
		{"status_404", 404},
		{"status_500", 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &rfc9457.Response{Status: tt.status}
			if got := err.HTTPStatusCode(); got != tt.status {
				t.Errorf("HTTPStatusCode: got %d, want %d", got, tt.status)
			}
		})
	}
}

func TestRFC9457Error_Write(t *testing.T) {
	err := &rfc9457.Response{
		Type:     rfc9457.InvalidParameterErrorType,
		Title:    "Invalid Parameter Type",
		Status:   422,
		Detail:   "Parameter 'id' expected type 'int' but received 'abc'",
		Instance: "/api/users/abc",
	}

	// Create test response writer
	recorder := httptest.NewRecorder()

	// Write error
	writeErr := err.Write(recorder)
	if writeErr != nil {
		t.Fatalf("Write error: %v", writeErr)
	}

	// Verify status code
	if recorder.Code != 422 {
		t.Errorf("Status code: got %d, want 422", recorder.Code)
	}

	// Verify Content-Type header
	contentType := recorder.Header().Get("Content-Type")
	if contentType != "application/problem+json" {
		t.Errorf("Content-Type: got %q, want %q", contentType, "application/problem+json")
	}

	// Verify JSON body
	var got rfc9457.Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("Unmarshal response: %v", err)
	}

	if got.Type != rfc9457.InvalidParameterErrorType {
		t.Errorf("Response Type: got %v, want %v", got.Type, rfc9457.InvalidParameterErrorType)
	}
	if got.Status != 422 {
		t.Errorf("Response Status: got %d, want 422", got.Status)
	}
	if got.Detail != "Parameter 'id' expected type 'int' but received 'abc'" {
		t.Errorf("Response Detail: got %q, want %q", got.Detail, "Parameter 'id' expected type 'int' but received 'abc'")
	}
}

func TestNewRFC9457Error(t *testing.T) {
	args := rfc9457.ResponseArgs{
		Type:     rfc9457.InvalidParameterErrorType,
		Title:    "Invalid Parameter Type",
		Status:   422,
		Detail:   "Test detail",
		Instance: "/test",
	}

	got := rfc9457.NewResponse(args)

	// Build expected result from args
	want := &rfc9457.Response{
		Type:     args.Type,
		Title:    args.Title,
		Status:   args.Status,
		Detail:   args.Detail,
		Instance: args.Instance,
	}

	assertRFC9457ErrorEqual(t, got, want)
}
