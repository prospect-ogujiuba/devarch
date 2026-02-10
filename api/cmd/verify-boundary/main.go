package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

type TestResult struct {
	Name           string `json:"name"`
	Pass           bool   `json:"pass"`
	TargetBytes    int64  `json:"target_bytes"`
	ResponseStatus int    `json:"response_status"`
	ResponseError  string `json:"response_error,omitempty"`
	DurationMs     int64  `json:"duration_ms"`
	Details        string `json:"details,omitempty"`
}

type TestSummary struct {
	Passed int `json:"passed"`
	Failed int `json:"failed"`
	Total  int `json:"total"`
}

type TestOutput struct {
	Tests   []TestResult `json:"tests"`
	Summary TestSummary  `json:"summary"`
}

var (
	apiURL    = flag.String("api-url", "http://localhost:8550", "API base URL")
	apiKey    = flag.String("api-key", os.Getenv("DEVARCH_API_KEY"), "API key for X-API-Key header")
	stackName = flag.String("stack-name", "boundary-test", "Stack name for test imports")
	jsonOut   = flag.Bool("json", false, "Machine-readable JSON output")
)

func main() {
	flag.Parse()

	// Ensure API is reachable
	if err := checkAPIHealth(); err != nil {
		fmt.Fprintf(os.Stderr, "API health check failed: %v\n", err)
		os.Exit(1)
	}

	results := []TestResult{}

	// Setup: Create test stack
	if err := setupTestStack(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup test stack: %v\n", err)
		os.Exit(1)
	}

	// Test 1: 200MB payload (should NOT be rejected)
	test1 := runTest("200MB import accepted", 200<<20, false)
	results = append(results, test1)

	// Re-create stack between tests
	if err := setupTestStack(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to re-create test stack: %v\n", err)
		os.Exit(1)
	}

	// Test 2: 300MB payload (should be rejected with 413)
	test2 := runTest("300MB import rejected", 300<<20, true)
	results = append(results, test2)

	// Cleanup
	cleanupTestStack()

	// Calculate summary
	summary := TestSummary{Total: len(results)}
	for _, r := range results {
		if r.Pass {
			summary.Passed++
		} else {
			summary.Failed++
		}
	}

	// Output results
	if *jsonOut {
		output := TestOutput{Tests: results, Summary: summary}
		json.NewEncoder(os.Stdout).Encode(output)
	} else {
		for i, r := range results {
			fmt.Printf("\n=== Test %d: %s ===\n", i+1, r.Name)
			if r.Pass {
				fmt.Printf("PASS  Status: %d", r.ResponseStatus)
				if r.ResponseError != "" {
					fmt.Printf(", error: %q", r.ResponseError)
				}
				fmt.Printf(" (target: %dMB, duration: %.1fs)\n", r.TargetBytes>>20, float64(r.DurationMs)/1000)
				if r.Details != "" {
					fmt.Printf("      %s\n", r.Details)
				}
			} else {
				fmt.Printf("FAIL  Status: %d", r.ResponseStatus)
				if r.ResponseError != "" {
					fmt.Printf(", error: %q", r.ResponseError)
				}
				fmt.Printf(" (target: %dMB, duration: %.1fs)\n", r.TargetBytes>>20, float64(r.DurationMs)/1000)
				if r.Details != "" {
					fmt.Printf("      %s\n", r.Details)
				}
			}
		}
		fmt.Printf("\nSummary: %d/%d passed\n", summary.Passed, summary.Total)
	}

	if summary.Failed > 0 {
		os.Exit(1)
	}
}

func checkAPIHealth() error {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(*apiURL + "/api/v1/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}
	return nil
}

func setupTestStack() error {
	// Delete if exists
	deleteTestStack()

	// Create fresh stack
	client := &http.Client{Timeout: 10 * time.Second}
	payload := map[string]string{
		"name":         *stackName,
		"network_name": "boundary-test-net",
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", *apiURL+"/api/v1/stacks", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if *apiKey != "" {
		req.Header.Set("X-API-Key", *apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create stack returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func deleteTestStack() error {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("DELETE", *apiURL+"/api/v1/stacks/"+*stackName, nil)
	if err != nil {
		return err
	}
	if *apiKey != "" {
		req.Header.Set("X-API-Key", *apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Ignore 404 - stack might not exist
	return nil
}

func cleanupTestStack() {
	deleteTestStack()
}

func runTest(name string, targetBytes int64, expectRejection bool) TestResult {
	start := time.Now()
	result := TestResult{
		Name:        name,
		TargetBytes: targetBytes,
	}

	// Generate streaming multipart payload
	reader, contentType, err := generateMultipartPayload(targetBytes, *stackName)
	if err != nil {
		result.Pass = false
		result.Details = fmt.Sprintf("Failed to generate payload: %v", err)
		result.DurationMs = time.Since(start).Milliseconds()
		return result
	}

	// Make request
	client := &http.Client{Timeout: 5 * time.Minute}
	req, err := http.NewRequest("POST", *apiURL+"/api/v1/stacks/"+*stackName+"/import", reader)
	if err != nil {
		result.Pass = false
		result.Details = fmt.Sprintf("Failed to create request: %v", err)
		result.DurationMs = time.Since(start).Milliseconds()
		return result
	}

	req.Header.Set("Content-Type", contentType)
	if *apiKey != "" {
		req.Header.Set("X-API-Key", *apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		// Check for connection reset - acceptable for oversized payloads
		if expectRejection && (strings.Contains(err.Error(), "connection reset") || strings.Contains(err.Error(), "EOF")) {
			result.Pass = true
			result.ResponseStatus = 0
			result.Details = "Connection reset by server (acceptable rejection)"
			result.DurationMs = time.Since(start).Milliseconds()
			return result
		}
		result.Pass = false
		result.Details = fmt.Sprintf("Request failed: %v", err)
		result.DurationMs = time.Since(start).Milliseconds()
		return result
	}
	defer resp.Body.Close()

	result.ResponseStatus = resp.StatusCode

	// Read response body
	bodyBytes, _ := io.ReadAll(resp.Body)

	if expectRejection {
		// Test 2: Expect 413
		if resp.StatusCode == http.StatusRequestEntityTooLarge {
			// Parse JSON response
			var errResp map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &errResp); err == nil {
				if errMsg, ok := errResp["error"].(string); ok {
					result.ResponseError = errMsg
				}
				// Validate required fields
				if maxBytes, ok := errResp["max_bytes"].(float64); ok && maxBytes == 268435456 {
					if receivedBytes, ok := errResp["received_bytes"].(float64); ok && receivedBytes > 0 {
						result.Pass = true
						result.Details = fmt.Sprintf("Correctly rejected with max_bytes=%d, received_bytes=%d", int64(maxBytes), int64(receivedBytes))
					} else {
						result.Pass = false
						result.Details = "Response missing valid received_bytes field"
					}
				} else {
					result.Pass = false
					result.Details = "Response missing valid max_bytes field"
				}
			} else {
				result.Pass = false
				result.Details = "Failed to parse JSON error response"
			}
		} else {
			result.Pass = false
			result.Details = fmt.Sprintf("Expected 413, got %d", resp.StatusCode)
		}
	} else {
		// Test 1: Expect NOT 413
		if resp.StatusCode != http.StatusRequestEntityTooLarge {
			result.Pass = true
			result.Details = "Payload accepted (not rejected by size limit)"
		} else {
			result.Pass = false
			result.Details = "Payload incorrectly rejected with 413"
			var errResp map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &errResp); err == nil {
				if errMsg, ok := errResp["error"].(string); ok {
					result.ResponseError = errMsg
				}
			}
		}
	}

	result.DurationMs = time.Since(start).Milliseconds()
	return result
}

// generateMultipartPayload creates a streaming multipart payload using io.Pipe
func generateMultipartPayload(targetBytes int64, stackName string) (io.Reader, string, error) {
	pr, pw := io.Pipe()

	writer := multipart.NewWriter(pw)
	contentType := writer.FormDataContentType()

	go func() {
		defer pw.Close()
		defer writer.Close()

		// Create form file field
		part, err := writer.CreateFormFile("file", "stack.yml")
		if err != nil {
			pw.CloseWithError(err)
			return
		}

		// YAML header
		header := fmt.Sprintf("version: 1\nstack:\n  name: %s\n  network_name: boundary-test-net\ninstances:\n", stackName)
		written := int64(len(header))
		if _, err := part.Write([]byte(header)); err != nil {
			pw.CloseWithError(err)
			return
		}

		// Generate synthetic instances until target size reached
		instanceNum := 0
		paddingChunk := strings.Repeat("A", 1024) // 1KB padding

		for written < targetBytes {
			instanceNum++
			instanceName := fmt.Sprintf("boundary-svc-%04d", instanceNum)

			// Instance header
			instanceHeader := fmt.Sprintf("  %s:\n    template: nginx\n    enabled: true\n    environment:\n", instanceName)
			if _, err := part.Write([]byte(instanceHeader)); err != nil {
				pw.CloseWithError(err)
				return
			}
			written += int64(len(instanceHeader))

			// Add padding env vars (~10 vars per instance for ~10KB)
			for i := 0; i < 10 && written < targetBytes; i++ {
				envLine := fmt.Sprintf("      PADDING_%d: \"%s\"\n", i, paddingChunk)
				if _, err := part.Write([]byte(envLine)); err != nil {
					pw.CloseWithError(err)
					return
				}
				written += int64(len(envLine))
			}
		}

		// CRITICAL: Close writer to finalize multipart boundary
		if err := writer.Close(); err != nil {
			pw.CloseWithError(err)
			return
		}
	}()

	return pr, contentType, nil
}
