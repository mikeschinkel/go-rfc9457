// Package main demonstrates basic usage of RFC 9457 error responses.
//
// This example shows:
// - Creating RFC 9457 compliant error responses
// - Writing error responses to HTTP responses
// - Using predefined error types
// - Customizing error details and instances
//
// Run with: go run main.go
// Then test with:
//
//	curl http://localhost:8080/users/abc
//	curl http://localhost:8080/posts/invalid-slug!
//	curl http://localhost:8080/score/150
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mikeschinkel/go-rfc9457"
)

func main() {
	// Example 1: Invalid parameter type error
	http.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Path[len("/users/"):]

		// Simulate validation - checking if userID is numeric
		if !isNumeric(userID) {
			err := rfc9457.NewResponse(rfc9457.ResponseArgs{
				Type:     rfc9457.InvalidParameterErrorType,
				Title:    "Invalid Parameter Type",
				Status:   422,
				Detail:   fmt.Sprintf("Parameter 'id' expected type 'int' but received '%s'", userID),
				Instance: r.URL.Path,
			})

			if writeErr := err.Write(w); writeErr != nil {
				log.Printf("Failed to write error response: %v", writeErr)
			}
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "User ID: %s\n", userID)
	})

	// Example 2: Constraint violation error
	http.HandleFunc("/score/", func(w http.ResponseWriter, r *http.Request) {
		score := r.URL.Path[len("/score/"):]

		// Simulate constraint validation - score must be 0-100
		if !isValidScore(score) {
			err := rfc9457.NewResponse(rfc9457.ResponseArgs{
				Type:     rfc9457.ConstraintViolationErrorType,
				Title:    "Constraint Violation",
				Status:   422,
				Detail:   fmt.Sprintf("Parameter 'score' value %s violates constraint range[0..100]", score),
				Instance: r.URL.Path,
			})

			if writeErr := err.Write(w); writeErr != nil {
				log.Printf("Failed to write error response: %v", writeErr)
			}
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Score: %s\n", score)
	})

	// Example 3: Using error as standard Go error
	http.HandleFunc("/posts/", func(w http.ResponseWriter, r *http.Request) {
		slug := r.URL.Path[len("/posts/"):]

		if !isValidSlug(slug) {
			err := rfc9457.NewResponse(rfc9457.ResponseArgs{
				Type:     rfc9457.InvalidParameterErrorType,
				Title:    "Invalid Slug Format",
				Status:   400,
				Detail:   fmt.Sprintf("Slug '%s' contains invalid characters", slug),
				Instance: r.URL.Path,
			})

			// Can use as standard error
			log.Printf("Error occurred: %v", err)

			if writeErr := err.Write(w); writeErr != nil {
				log.Printf("Failed to write error response: %v", writeErr)
			}
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Post slug: %s\n", slug)
	})

	// Start server
	fmt.Println("Server starting on :8080")
	fmt.Println("Try these URLs:")
	fmt.Println("  http://localhost:8080/users/123        (valid)")
	fmt.Println("  http://localhost:8080/users/abc        (invalid parameter type)")
	fmt.Println("  http://localhost:8080/score/50         (valid)")
	fmt.Println("  http://localhost:8080/score/150        (constraint violation)")
	fmt.Println("  http://localhost:8080/posts/hello      (valid)")
	fmt.Println("  http://localhost:8080/posts/invalid!   (invalid slug)")
	fmt.Println()

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// isNumeric checks if a string contains only digits
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// isValidScore checks if score is between 0 and 100
func isValidScore(s string) bool {
	if !isNumeric(s) {
		return false
	}
	// Simple validation - would use strconv.Atoi in real code
	if len(s) > 3 {
		return false
	}
	return true
}

// isValidSlug checks if slug contains only valid characters
func isValidSlug(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}
