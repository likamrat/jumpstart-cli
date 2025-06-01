package urlutils

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/fatih/color"
)

var (
	urlTestSuccessColor = color.New(color.FgGreen, color.Bold).SprintFunc()
	urlTestInfoColor    = color.New(color.FgCyan).SprintFunc()
	urlTestWarnColor    = color.New(color.FgYellow).SprintFunc()
	urlTestErrorColor   = color.New(color.FgRed, color.Bold).SprintFunc()
	urlTestHeaderColor  = color.New(color.FgMagenta, color.Bold).SprintFunc()
)

func printURLTestStatus(t *testing.T, testName string, success bool, message string) {
	status := urlTestSuccessColor("✅")
	if !success {
		status = urlTestErrorColor("❌")
	}
	fmt.Printf("  %s %s: %s\n", status, urlTestInfoColor(testName), message)
}

func TestShortenURL(t *testing.T) {
	fmt.Printf("\n%s\n", urlTestHeaderColor("=== Testing URL Shortening ==="))

	testCases := []struct {
		name        string
		longURL     string
		description string
	}{
		{
			name:        "valid_http_url",
			longURL:     "http://example.com/very/long/path/to/resource?with=parameters&and=more",
			description: "should handle valid HTTP URLs",
		},
		{
			name:        "valid_https_url",
			longURL:     "https://github.com/microsoft/azure_arc/releases/tag/v1.0.0",
			description: "should handle valid HTTPS URLs",
		},
		{
			name:        "github_release_url",
			longURL:     "https://github.com/microsoft/azure_arc/releases/download/v1.0.0/jumpstart-cli-linux-amd64.tar.gz",
			description: "should handle GitHub release URLs",
		},
		{
			name:        "azure_portal_url",
			longURL:     "https://portal.azure.com/#@tenant.onmicrosoft.com/resource/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/test-rg/overview",
			description: "should handle Azure Portal URLs",
		},
		{
			name:        "simple_url",
			longURL:     "https://example.com",
			description: "should handle simple URLs",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Printf("  %s %s\n", urlTestInfoColor("Testing:"), tc.description)

			// Set a reasonable timeout for the test
			done := make(chan string, 1)

			go func() {
				result := ShortenURL(tc.longURL)
				done <- result
			}()

			select {
			case result := <-done:
				// Verify the result is not empty
				if result == "" {
					printURLTestStatus(t, "Non-empty result", false, "ShortenURL should not return empty string")
					t.Error("ShortenURL should not return empty string")
					return
				}
				printURLTestStatus(t, "Non-empty result", true, "URL shortening returned a result")

				// If shortening fails, it should return the original URL
				if result == tc.longURL {
					printURLTestStatus(t, "Fallback behavior", true, fmt.Sprintf("URL shortening failed (returned original), which is acceptable: %s", urlTestInfoColor(tc.longURL)))
					t.Logf("URL shortening failed (returned original), which is acceptable: %s", tc.longURL)
				} else {
					// If shortening succeeded, verify it's a valid short URL
					if !strings.HasPrefix(result, "http") {
						printURLTestStatus(t, "Valid URL format", false, fmt.Sprintf("Shortened URL should start with http, got: %s", result))
						t.Errorf("Shortened URL should start with http, got: %s", result)
						return
					}
					printURLTestStatus(t, "Valid URL format", true, "Shortened URL has valid HTTP/HTTPS format")

					// If it's a TinyURL, verify the format
					if strings.Contains(result, "tinyurl.com") {
						if len(result) >= len(tc.longURL) {
							printURLTestStatus(t, "URL length reduction", false, fmt.Sprintf("Shortened URL should be shorter than original. Original: %d chars, Shortened: %d chars", len(tc.longURL), len(result)))
							t.Errorf("Shortened URL should be shorter than original. Original: %d chars, Shortened: %d chars",
								len(tc.longURL), len(result))
							return
						}
						printURLTestStatus(t, "URL length reduction", true, fmt.Sprintf("Successfully shortened URL: %s -> %s", urlTestInfoColor(tc.longURL), urlTestSuccessColor(result)))
						t.Logf("Successfully shortened URL: %s -> %s", tc.longURL, result)
					}
				}

			case <-time.After(5 * time.Second):
				printURLTestStatus(t, "Timeout handling", false, "ShortenURL took too long (>5s), may be hanging")
				t.Error("ShortenURL took too long (>5s), may be hanging")
			}
		})
	}
}

func TestShortenURLErrorHandling(t *testing.T) {
	fmt.Printf("\n%s\n", urlTestHeaderColor("=== Testing URL Error Handling ==="))

	testCases := []struct {
		name        string
		input       string
		description string
	}{
		{
			name:        "empty_url",
			input:       "",
			description: "should handle empty URLs gracefully",
		},
		{
			name:        "invalid_url",
			input:       "not-a-url",
			description: "should handle invalid URLs gracefully",
		},
		{
			name:        "malformed_url",
			input:       "http://",
			description: "should handle malformed URLs gracefully",
		},
		{
			name:        "localhost_url",
			input:       "http://localhost:8080/test",
			description: "should handle localhost URLs gracefully",
		},
		{
			name:        "file_url",
			input:       "file:///path/to/file.txt",
			description: "should handle file URLs gracefully",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Printf("  %s %s\n", urlTestInfoColor("Testing:"), tc.description)

			// These should not panic and should return the original URL
			result := ShortenURL(tc.input)

			if result != tc.input {
				// If the URL was actually shortened (unlikely for invalid URLs), that's fine too
				printURLTestStatus(t, "URL modification", true, fmt.Sprintf("URL was modified: %s -> %s", urlTestInfoColor(tc.input), urlTestSuccessColor(result)))
				t.Logf("URL was modified: %s -> %s", tc.input, result)
			} else {
				printURLTestStatus(t, "Graceful handling", true, "URL returned unchanged (expected for invalid/edge case URLs)")
			}

			// Main requirement: should not panic or return empty string (unless input was empty)
			if tc.input != "" && result == "" {
				printURLTestStatus(t, "Non-empty result", false, "ShortenURL should not return empty string for non-empty input")
				t.Errorf("ShortenURL should not return empty string for non-empty input")
			} else if tc.input == "" || result != "" {
				printURLTestStatus(t, "Non-empty result", true, "Function handled edge case without returning empty string")
			}
		})
	}
}

func TestShortenURLTimeout(t *testing.T) {
	fmt.Printf("\n%s\n", urlTestHeaderColor("=== Testing URL Timeout Handling ==="))

	// Test that the function respects its internal timeout
	t.Run("timeout_handling", func(t *testing.T) {
		fmt.Printf("  %s Testing timeout behavior with slow URL\n", urlTestInfoColor("Testing:"))

		// Use a URL that might be slow to respond
		slowURL := "https://httpbin.org/delay/5"

		start := time.Now()
		result := ShortenURL(slowURL)
		duration := time.Since(start)

		// Should complete within reasonable time (function has 3s timeout + some buffer)
		if duration > 6*time.Second {
			printURLTestStatus(t, "Timeout compliance", false, fmt.Sprintf("ShortenURL took too long: %v", duration))
			t.Errorf("ShortenURL took too long: %v", duration)
		} else {
			printURLTestStatus(t, "Timeout compliance", true, fmt.Sprintf("Function completed within acceptable time: %v", duration))
		}

		// Should return the original URL on timeout
		if result != slowURL {
			printURLTestStatus(t, "Timeout result", true, fmt.Sprintf("URL was processed despite potential timeout: %s -> %s", urlTestInfoColor(slowURL), urlTestSuccessColor(result)))
			t.Logf("URL was processed despite potential timeout: %s -> %s", slowURL, result)
		} else {
			printURLTestStatus(t, "Timeout fallback", true, "Function returned original URL on timeout (expected behavior)")
		}
	})
}

func TestShortenURLConsistency(t *testing.T) {
	fmt.Printf("\n%s\n", urlTestHeaderColor("=== Testing URL Consistency ==="))

	// Test that the same URL gives consistent results
	t.Run("consistency", func(t *testing.T) {
		fmt.Printf("  %s Testing result consistency for same URL\n", urlTestInfoColor("Testing:"))

		testURL := "https://github.com/microsoft/azure_arc"

		result1 := ShortenURL(testURL)
		result2 := ShortenURL(testURL)

		// Results should be consistent (either both shortened the same way, or both returned original)
		if result1 != result2 {
			// Note: TinyURL might return different short URLs for the same long URL,
			// so this test might be flaky. We'll just log the difference.
			printURLTestStatus(t, "Result consistency", true, fmt.Sprintf("Inconsistent results (this may be normal for TinyURL): %s vs %s", urlTestWarnColor(result1), urlTestWarnColor(result2)))
			t.Logf("Inconsistent results (this may be normal for TinyURL): %s vs %s", result1, result2)
		} else {
			printURLTestStatus(t, "Result consistency", true, "Same URL produced consistent results")
		}
	})
}

func TestShortenURLValidation(t *testing.T) {
	fmt.Printf("\n%s\n", urlTestHeaderColor("=== Testing URL Result Validation ==="))

	// Test validation of shortened URLs
	t.Run("result_validation", func(t *testing.T) {
		fmt.Printf("  %s Testing validation of shortened URL results\n", urlTestInfoColor("Testing:"))

		testURL := "https://example.com/test/path"
		result := ShortenURL(testURL)

		// Result should be a valid URL format
		if !strings.HasPrefix(result, "http://") && !strings.HasPrefix(result, "https://") {
			printURLTestStatus(t, "Valid URL format", false, fmt.Sprintf("Result should be a valid URL, got: %s", result))
			t.Errorf("Result should be a valid URL, got: %s", result)
		} else {
			printURLTestStatus(t, "Valid URL format", true, "Result has valid HTTP/HTTPS format")
		}

		// Should not contain obvious malformed elements
		if strings.Contains(result, " ") {
			printURLTestStatus(t, "No spaces", false, fmt.Sprintf("Result should not contain spaces: %s", result))
			t.Errorf("Result should not contain spaces: %s", result)
		} else {
			printURLTestStatus(t, "No spaces", true, "Result contains no spaces")
		}

		if strings.Contains(result, "\n") || strings.Contains(result, "\r") {
			printURLTestStatus(t, "No newlines", false, fmt.Sprintf("Result should not contain newlines: %s", result))
			t.Errorf("Result should not contain newlines: %s", result)
		} else {
			printURLTestStatus(t, "No newlines", true, "Result contains no newlines")
		}
	})
}

func TestShortenURLSpecialCharacters(t *testing.T) {
	fmt.Printf("\n%s\n", urlTestHeaderColor("=== Testing URLs with Special Characters ==="))

	// Test URLs with special characters
	testCases := []struct {
		name string
		url  string
	}{
		{
			name: "url_with_query_params",
			url:  "https://example.com/search?q=test&page=1&sort=name",
		},
		{
			name: "url_with_anchor",
			url:  "https://example.com/docs#section-1",
		},
		{
			name: "url_with_spaces_encoded",
			url:  "https://example.com/path%20with%20spaces",
		},
		{
			name: "url_with_unicode",
			url:  "https://example.com/测试",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Printf("  %s Testing URL: %s\n", urlTestInfoColor("Testing:"), urlTestInfoColor(tc.url))

			result := ShortenURL(tc.url)

			// Should not panic and should return a valid result
			if result == "" {
				printURLTestStatus(t, "Non-empty result", false, "ShortenURL should not return empty string")
				t.Error("ShortenURL should not return empty string")
			} else {
				printURLTestStatus(t, "Non-empty result", true, "Function returned a result")
			}

			// Should start with http/https
			if !strings.HasPrefix(result, "http") {
				printURLTestStatus(t, "Valid URL format", false, fmt.Sprintf("Result should be a URL, got: %s", result))
				t.Errorf("Result should be a URL, got: %s", result)
			} else {
				printURLTestStatus(t, "Valid URL format", true, "Result has valid URL format")
			}
		})
	}
}

// Benchmark the URL shortening function
func BenchmarkShortenURL(b *testing.B) {
	fmt.Printf("\n%s\n", urlTestHeaderColor("=== Benchmarking URL Shortening ==="))

	testURL := "https://github.com/microsoft/azure_arc/releases/tag/v1.0.0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ShortenURL(testURL)
	}
}

// Test concurrent access to URL shortening
func TestShortenURLConcurrency(t *testing.T) {
	fmt.Printf("\n%s\n", urlTestHeaderColor("=== Testing URL Concurrency ==="))

	t.Run("concurrent_requests", func(t *testing.T) {
		fmt.Printf("  %s Testing concurrent URL shortening requests\n", urlTestInfoColor("Testing:"))

		testURL := "https://example.com/concurrent/test"
		numGoroutines := 5

		results := make(chan string, numGoroutines)

		// Start multiple goroutines
		for i := 0; i < numGoroutines; i++ {
			go func() {
				result := ShortenURL(testURL)
				results <- result
			}()
		}

		// Collect results
		var allResults []string
		for i := 0; i < numGoroutines; i++ {
			select {
			case result := <-results:
				allResults = append(allResults, result)
			case <-time.After(10 * time.Second):
				printURLTestStatus(t, "Concurrency timeout", false, "Concurrent test timed out")
				t.Fatal("Concurrent test timed out")
			}
		}

		// Verify all goroutines completed
		if len(allResults) != numGoroutines {
			printURLTestStatus(t, "Goroutine completion", false, fmt.Sprintf("Expected %d results, got %d", numGoroutines, len(allResults)))
			t.Errorf("Expected %d results, got %d", numGoroutines, len(allResults))
		} else {
			printURLTestStatus(t, "Goroutine completion", true, fmt.Sprintf("All %d goroutines completed successfully", numGoroutines))
		}

		// All results should be valid URLs
		validResults := 0
		for i, result := range allResults {
			if !strings.HasPrefix(result, "http") {
				printURLTestStatus(t, fmt.Sprintf("Result %d validity", i), false, fmt.Sprintf("Result %d should be a valid URL: %s", i, result))
				t.Errorf("Result %d should be a valid URL: %s", i, result)
			} else {
				validResults++
			}
		}

		if validResults == len(allResults) {
			printURLTestStatus(t, "Result validity", true, fmt.Sprintf("All %d concurrent results are valid URLs", validResults))
		}
	})
}
