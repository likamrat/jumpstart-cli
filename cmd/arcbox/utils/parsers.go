package utils

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// This file contains parsing functions for ArcBox deployment parameters,
// configuration files, and Azure CLI output processing

// ParseISO8601Duration parses Azure's ISO 8601 duration format (e.g., "PT1H30M45S")
func ParseISO8601Duration(duration string) (time.Duration, error) {
	// Remove "PT" prefix
	if !strings.HasPrefix(duration, "PT") {
		return 0, fmt.Errorf("invalid ISO 8601 duration format")
	}
	duration = duration[2:]

	var totalDuration time.Duration

	// Parse hours
	if idx := strings.Index(duration, "H"); idx != -1 {
		hours, err := strconv.Atoi(duration[:idx])
		if err != nil {
			return 0, err
		}
		totalDuration += time.Duration(hours) * time.Hour
		duration = duration[idx+1:]
	}

	// Parse minutes
	if idx := strings.Index(duration, "M"); idx != -1 {
		minutes, err := strconv.Atoi(duration[:idx])
		if err != nil {
			return 0, err
		}
		totalDuration += time.Duration(minutes) * time.Minute
		duration = duration[idx+1:]
	}

	// Parse seconds
	if idx := strings.Index(duration, "S"); idx != -1 {
		seconds, err := strconv.ParseFloat(duration[:idx], 64)
		if err != nil {
			return 0, err
		}
		totalDuration += time.Duration(seconds * float64(time.Second))
	}

	return totalDuration, nil
}

// ParseInt64 safely converts interface{} to int64
func ParseInt64(val interface{}) int {
	switch v := val.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case string:
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return 0
}
