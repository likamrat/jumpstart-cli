// urlutils.go - Shared URL utility functions
package utils

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ShortenURL creates a short URL using TinyURL service
// Returns the original URL if shortening fails
func ShortenURL(longURL string) string {
	// Use TinyURL API to create a short URL
	apiURL := "http://tinyurl.com/api-create.php?url=" + url.QueryEscape(longURL)

	// Set a short timeout for the HTTP request
	client := &http.Client{Timeout: 3 * time.Second}

	resp, err := client.Get(apiURL)
	if err != nil {
		// If shortening fails, return original URL
		return longURL
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// If shortening fails, return original URL
		return longURL
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// If shortening fails, return original URL
		return longURL
	}

	shortURL := strings.TrimSpace(string(body))
	// Validate that we got a proper TinyURL response
	if strings.HasPrefix(shortURL, "http") && len(shortURL) < len(longURL) && strings.Contains(shortURL, "tinyurl.com") {
		return shortURL
	}

	// If shortening didn't work properly, return original URL
	return longURL
}
