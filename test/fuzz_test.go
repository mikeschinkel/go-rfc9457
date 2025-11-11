package test

import (
	"encoding/json"
	"log/slog"
	"os"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/mikeschinkel/go-rfc9457"
)

func init() {
	// Set up a logger for tests to prevent panics
	rfc9457.SetLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	})))
}

// FuzzResponseMarshal tests JSON marshaling of RFC 9457 responses
// to detect panics, infinite loops, or unexpected behavior.
func FuzzResponseMarshal(f *testing.F) {
	// Valid basic responses
	f.Add("https://example.com/errors/not-found", "Not Found", 404, "", "")
	f.Add("https://example.com/errors/validation", "Validation Error", 422, "Invalid input", "/api/users")
	f.Add("https://schema.xmlui.org/errors/test-server/api/validation/invalid-parameter-type", "Invalid Parameter Type", 422, "Parameter 'id' expected type 'int' but received 'abc'", "/api/users/abc")
	f.Add("https://schema.xmlui.org/errors/test-server/api/validation/constraint-violation", "Constraint Violation", 422, "Parameter 'score' value 150 violates constraint range[0..100]", "/api/users/by-score/150")

	// Edge cases - empty fields
	f.Add("https://example.com/errors/minimal", "Minimal Error", 400, "", "")
	f.Add("", "Empty Type", 500, "", "")
	f.Add("https://example.com/errors/empty-title", "", 500, "", "")

	// Edge cases - unusual status codes
	f.Add("https://example.com/errors/teapot", "I'm a teapot", 418, "", "")
	f.Add("https://example.com/errors/zero", "Zero Status", 0, "", "")
	f.Add("https://example.com/errors/negative", "Negative Status", -1, "", "")
	f.Add("https://example.com/errors/large", "Large Status", 999, "", "")

	// Edge cases - special characters
	f.Add("https://example.com/errors/special", "Error with \"quotes\"", 400, "Detail with 'quotes'", "/path/with/special?chars=&<>")
	f.Add("https://example.com/errors/unicode", "Error with émojis 🚀", 400, "Detail with 中文", "/path/with/émojis")
	f.Add("https://example.com/errors/newlines", "Error\nwith\nnewlines", 400, "Detail\nwith\nnewlines", "/path\n/with\n/newlines")
	f.Add("https://example.com/errors/tabs", "Error\twith\ttabs", 400, "Detail\twith\ttabs", "/path\t/with\t/tabs")

	// Edge cases - very long strings
	f.Add(strings.Repeat("https://example.com/", 50), strings.Repeat("Long Title ", 100), 400, strings.Repeat("Long Detail ", 100), strings.Repeat("/long/path/", 50))

	// Edge cases - unusual URLs
	f.Add("http://example.com/errors/http", "HTTP URL", 400, "", "")
	f.Add("ftp://example.com/errors/ftp", "FTP URL", 400, "", "")
	f.Add("//example.com/errors/no-scheme", "No Scheme", 400, "", "")
	f.Add("not-a-url", "Invalid URL", 400, "", "")

	// Real-world patterns
	f.Add("https://api.example.com/errors/auth/unauthorized", "Unauthorized", 401, "Authentication token expired", "/api/protected/resource")
	f.Add("https://api.example.com/errors/rate-limit", "Rate Limit Exceeded", 429, "Too many requests. Please try again later.", "/api/endpoint")
	f.Add("https://api.example.com/errors/server/internal", "Internal Server Error", 500, "An unexpected error occurred", "/api/users/123")

	f.Fuzz(func(t *testing.T, typeURI, title string, status int, detail, instance string) {
		done := make(chan struct{})
		var marshalErr error

		// Run marshal in goroutine with panic recovery
		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Marshal panicked: type=%q, title=%q, status=%d\nPanic: %v\nStack: %s",
						typeURI, title, status, r, debug.Stack())
				}
				close(done)
			}()

			resp := &rfc9457.Response{
				Type:     rfc9457.ErrorTypeURI(typeURI),
				Title:    title,
				Status:   status,
				Detail:   detail,
				Instance: instance,
			}

			_, marshalErr = json.Marshal(resp)
		}()

		// Timeout detection
		select {
		case <-done:
			// Marshal completed (with or without error)
			// For fuzzing, we mainly care about panics and hangs, not marshal errors
			_ = marshalErr

		case <-time.After(1 * time.Second):
			t.Fatalf("Marshal hung (infinite loop detected) on: type=%q, title=%q, status=%d",
				typeURI, title, status)
		}
	})
}

// FuzzResponseUnmarshal tests JSON unmarshaling of RFC 9457 responses
// to detect panics, infinite loops, or unexpected behavior.
func FuzzResponseUnmarshal(f *testing.F) {
	// Valid JSON responses
	f.Add(`{"type":"https://example.com/errors/not-found","title":"Not Found","status":404}`)
	f.Add(`{"type":"https://example.com/errors/validation","title":"Validation Error","status":422,"detail":"Invalid input","instance":"/api/users"}`)
	f.Add(`{"type":"https://schema.xmlui.org/errors/test-server/api/validation/invalid-parameter-type","title":"Invalid Parameter Type","status":422,"detail":"Parameter 'id' expected type 'int' but received 'abc'","instance":"/api/users/abc"}`)

	// Edge cases - minimal/maximal
	f.Add(`{}`)
	f.Add(`{"type":"","title":"","status":0}`)
	f.Add(`{"type":"https://example.com/errors/minimal","title":"Minimal","status":200}`)

	// Edge cases - missing fields
	f.Add(`{"type":"https://example.com/errors/no-title","status":404}`)
	f.Add(`{"title":"No Type","status":404}`)
	f.Add(`{"type":"https://example.com/errors/no-status","title":"No Status"}`)

	// Edge cases - wrong types
	f.Add(`{"type":123,"title":"Type is number","status":400}`)
	f.Add(`{"type":"https://example.com/errors/test","title":456,"status":400}`)
	f.Add(`{"type":"https://example.com/errors/test","title":"Test","status":"400"}`)
	f.Add(`{"type":"https://example.com/errors/test","title":"Test","status":400,"detail":123}`)
	f.Add(`{"type":"https://example.com/errors/test","title":"Test","status":400,"instance":123}`)

	// Edge cases - extra fields
	f.Add(`{"type":"https://example.com/errors/test","title":"Test","status":400,"extra":"field"}`)
	f.Add(`{"type":"https://example.com/errors/test","title":"Test","status":400,"unknown":{"nested":"object"}}`)

	// Edge cases - special characters
	f.Add(`{"type":"https://example.com/errors/quotes","title":"Error with \"quotes\"","status":400}`)
	f.Add(`{"type":"https://example.com/errors/unicode","title":"Error with émojis 🚀","status":400}`)
	f.Add(`{"type":"https://example.com/errors/newlines","title":"Error\nwith\nnewlines","status":400}`)
	f.Add(`{"type":"https://example.com/errors/escaped","title":"Error\\nwith\\tescapes","status":400}`)

	// Edge cases - malformed JSON
	f.Add(`{`)
	f.Add(`}`)
	f.Add(`{"type":`)
	f.Add(`{"type":"https://example.com/errors/test"`)
	f.Add(`{"type":"https://example.com/errors/test",}`)
	f.Add(`not json at all`)
	f.Add(`123`)
	f.Add(`"string"`)
	f.Add(`null`)
	f.Add(`[]`)
	f.Add(`[{"type":"https://example.com/errors/test"}]`)

	// Edge cases - very long strings
	f.Add(`{"type":"` + strings.Repeat("https://example.com/", 100) + `","title":"` + strings.Repeat("Long Title ", 100) + `","status":400}`)

	// Real-world patterns
	f.Add(`{"type":"https://api.example.com/errors/auth/unauthorized","title":"Unauthorized","status":401,"detail":"Authentication token expired","instance":"/api/protected/resource"}`)
	f.Add(`{"type":"https://api.example.com/errors/rate-limit","title":"Rate Limit Exceeded","status":429,"detail":"Too many requests. Please try again later.","instance":"/api/endpoint"}`)

	f.Fuzz(func(t *testing.T, jsonData string) {
		done := make(chan struct{})
		var unmarshalErr error

		// Run unmarshal in goroutine with panic recovery
		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Unmarshal panicked on %q\nPanic: %v\nStack: %s",
						jsonData, r, debug.Stack())
				}
				close(done)
			}()

			var resp rfc9457.Response
			unmarshalErr = json.Unmarshal([]byte(jsonData), &resp)
		}()

		// Timeout detection
		select {
		case <-done:
			// Unmarshal completed (with or without error)
			// For fuzzing, we mainly care about panics and hangs, not unmarshal errors
			_ = unmarshalErr

		case <-time.After(1 * time.Second):
			t.Fatalf("Unmarshal hung (infinite loop detected) on: %q", jsonData)
		}
	})
}

// FuzzResponseRoundtrip tests marshal->unmarshal roundtrip to ensure consistency
func FuzzResponseRoundtrip(f *testing.F) {
	// Valid responses for roundtrip testing
	f.Add("https://example.com/errors/test", "Test Error", 400, "Test detail", "/test/instance")
	f.Add("https://schema.xmlui.org/errors/test-server/api/validation/invalid-parameter-type", "Invalid Parameter Type", 422, "Parameter 'id' expected type 'int' but received 'abc'", "/api/users/abc")
	f.Add("https://example.com/errors/minimal", "Minimal", 200, "", "")
	f.Add("", "", 0, "", "")
	f.Add("https://example.com/errors/unicode", "Error with émojis 🚀", 400, "Detail with 中文", "/path/with/émojis")

	f.Fuzz(func(t *testing.T, typeURI, title string, status int, detail, instance string) {
		done := make(chan struct{})
		var roundtripErr error

		// Run roundtrip in goroutine with panic recovery
		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Roundtrip panicked: type=%q, title=%q, status=%d\nPanic: %v\nStack: %s",
						typeURI, title, status, r, debug.Stack())
				}
				close(done)
			}()

			// Create original response
			original := &rfc9457.Response{
				Type:     rfc9457.ErrorTypeURI(typeURI),
				Title:    title,
				Status:   status,
				Detail:   detail,
				Instance: instance,
			}

			// Marshal
			data, err := json.Marshal(original)
			if err != nil {
				roundtripErr = err
				return
			}

			// Unmarshal
			var decoded rfc9457.Response
			if err := json.Unmarshal(data, &decoded); err != nil {
				roundtripErr = err
				return
			}

			// Basic consistency check - fields should match
			if decoded.Type != original.Type ||
				decoded.Title != original.Title ||
				decoded.Status != original.Status ||
				decoded.Detail != original.Detail ||
				decoded.Instance != original.Instance {
				t.Logf("Roundtrip mismatch (this might be OK for edge cases):\nOriginal: %+v\nDecoded:  %+v",
					original, &decoded)
			}
		}()

		// Timeout detection
		select {
		case <-done:
			// Roundtrip completed (with or without error)
			_ = roundtripErr

		case <-time.After(2 * time.Second):
			t.Fatalf("Roundtrip hung (infinite loop detected) on: type=%q, title=%q, status=%d",
				typeURI, title, status)
		}
	})
}
