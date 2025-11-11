package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mikeschinkel/go-rfc9457"
)

// TestFuzzCorpus runs all discovered corpus files from fuzzing to ensure
// they continue to pass without panics or infinite loops. This test runs
// on every CI/CD build and is much faster than full fuzzing.
func TestFuzzCorpus(t *testing.T) {
	testFuzzCorpusForFuzz(t, "FuzzResponseMarshal", testResponseMarshalCorpus)
	testFuzzCorpusForFuzz(t, "FuzzResponseUnmarshal", testResponseUnmarshalCorpus)
	testFuzzCorpusForFuzz(t, "FuzzResponseRoundtrip", testResponseRoundtripCorpus)
}

func testFuzzCorpusForFuzz(t *testing.T, fuzzName string, testFunc func(*testing.T, string)) {
	t.Run(fuzzName, func(t *testing.T) {
		corpusDir := filepath.Join("testdata", "fuzz", fuzzName)

		entries, err := os.ReadDir(corpusDir)
		if err != nil {
			if os.IsNotExist(err) {
				//t.Skipf("No corpus directory found for %s (run fuzzing first to generate corpus)", fuzzName)
				return
			}
			t.Fatalf("Failed to read corpus directory: %v", err)
		}

		if len(entries) == 0 {
			t.Skip("Corpus directory is empty")
			return
		}

		var infiniteLoops, panics, errors, successes int

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			// Read corpus file
			corpusPath := filepath.Join(corpusDir, entry.Name())
			data, err := os.ReadFile(corpusPath)
			if err != nil {
				t.Errorf("Failed to read corpus file %s: %v", entry.Name(), err)
				continue
			}

			// Extract input from corpus file
			input := extractCorpusInput(string(data))
			if input == "" {
				continue
			}

			// Test with timeout
			done := make(chan struct{})
			var testErr error
			var panicked bool

			go func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = true
						t.Errorf("PANIC in corpus file %s: %v", entry.Name(), r)
					}
					close(done)
				}()

				// Use a subtest to capture test errors
				testFunc(t, input)
			}()

			select {
			case <-done:
				if panicked {
					panics++
				} else if testErr != nil {
					errors++
				} else {
					successes++
				}

			case <-time.After(2 * time.Second):
				infiniteLoops++
				t.Errorf("INFINITE LOOP detected in corpus file %s", entry.Name())
			}
		}

		// Fail test if we found critical issues
		if panics > 0 {
			t.Errorf("Found %d corpus files causing panics", panics)
		}
		if infiniteLoops > 0 {
			t.Errorf("Found %d corpus files causing infinite loops", infiniteLoops)
		}
	})
}

// testResponseMarshalCorpus tests a single marshal corpus input
func testResponseMarshalCorpus(t *testing.T, input string) {
	// For FuzzResponseMarshal, we expect 5 string fields
	parts := parseCorpusParts(input, 5)
	if len(parts) != 5 {
		return // Skip malformed corpus
	}

	typeURI, title, statusStr, detail, instance := parts[0], parts[1], parts[2], parts[3], parts[4]

	// Parse status as integer
	status, err := strconv.Atoi(statusStr)
	if err != nil {
		status = 0 // Use 0 for unparseable status
	}

	resp := &rfc9457.Response{
		Type:     rfc9457.ErrorTypeURI(typeURI),
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: instance,
	}

	_, _ = json.Marshal(resp)
	// We don't fail on marshal errors - just checking for panics/hangs
}

// testResponseUnmarshalCorpus tests a single unmarshal corpus input
func testResponseUnmarshalCorpus(t *testing.T, input string) {
	var resp rfc9457.Response
	_ = json.Unmarshal([]byte(input), &resp)
	// We don't fail on unmarshal errors - just checking for panics/hangs
}

// testResponseRoundtripCorpus tests a single roundtrip corpus input
func testResponseRoundtripCorpus(t *testing.T, input string) {
	// For FuzzResponseRoundtrip, we expect 5 string fields
	parts := parseCorpusParts(input, 5)
	if len(parts) != 5 {
		return // Skip malformed corpus
	}

	typeURI, title, statusStr, detail, instance := parts[0], parts[1], parts[2], parts[3], parts[4]

	// Parse status as integer
	status, err := strconv.Atoi(statusStr)
	if err != nil {
		status = 0
	}

	original := &rfc9457.Response{
		Type:     rfc9457.ErrorTypeURI(typeURI),
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: instance,
	}

	data, err := json.Marshal(original)
	if err != nil {
		return // Skip if marshal fails
	}

	var decoded rfc9457.Response
	_ = json.Unmarshal(data, &decoded)
	// We don't fail on roundtrip mismatches - just checking for panics/hangs
}

// extractCorpusInput extracts the first input from a Go fuzz corpus file.
// Corpus file format: "go test fuzz v1\nstring(\"value\")" or "go test fuzz v1\nstring(\"str1\")\nstring(\"str2\")..."
func extractCorpusInput(corpusData string) string {
	lines := strings.Split(corpusData, "\n")
	if len(lines) < 2 {
		return ""
	}

	// For multi-parameter fuzzing, we combine all parameters into a single string
	// separated by a delimiter that's unlikely to appear in real data
	var inputs []string

	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		if strings.HasPrefix(line, "string(") {
			// Extract the quoted string
			start := strings.Index(line, `"`)
			if start == -1 {
				continue
			}

			// Find the end of the string literal
			end := strings.LastIndex(line, `"`) + 1
			if end > start {
				candidate := line[start:end]
				if unescaped, err := strconv.Unquote(candidate); err == nil {
					inputs = append(inputs, unescaped)
				}
			}
		} else if strings.HasPrefix(line, "int(") {
			// Extract integer value
			start := strings.Index(line, "(") + 1
			end := strings.Index(line, ")")
			if start > 0 && end > start {
				intStr := line[start:end]
				inputs = append(inputs, intStr)
			}
		}
	}

	if len(inputs) == 0 {
		return ""
	}

	// Join all inputs with a special delimiter
	return strings.Join(inputs, "\x00")
}

// parseCorpusParts splits corpus input back into individual parameters
func parseCorpusParts(input string, expectedCount int) []string {
	parts := strings.Split(input, "\x00")
	if len(parts) == expectedCount {
		return parts
	}

	// If we don't have the expected count, pad with empty strings or truncate
	if len(parts) < expectedCount {
		for len(parts) < expectedCount {
			parts = append(parts, "")
		}
	} else {
		parts = parts[:expectedCount]
	}

	return parts
}
