// validators.go - Validation utilities for ArcBox command
package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"jumpstartcli/internal/artifacts/regions"
	"jumpstartcli/internal/utils"
)

// ValidateLocations checks if all provided locations are valid Azure regions supported by ArcBox
func ValidateLocations(locations []string) error {
	if len(locations) == 0 {
		return fmt.Errorf("no locations provided")
	}

	// Load supported ArcBox regions and create lookup maps
	var supportedRegions []string
	if err := json.Unmarshal(regions.ArcboxSupportedRegionsData, &supportedRegions); err != nil {
		return fmt.Errorf("failed to load supported regions: %v", err)
	}

	// Create maps for both normalized names and display names
	normalizedToDisplay := make(map[string]string)
	displayToNormalized := make(map[string]string)

	for _, region := range supportedRegions {
		normalized := utils.NormalizeRegion(region)
		normalizedToDisplay[normalized] = region
		displayToNormalized[region] = normalized
	}

	var invalidLocations []string
	var suggestions []string

	for _, location := range locations {
		location = strings.TrimSpace(location)
		if location == "" {
			continue
		}

		normalized := utils.NormalizeRegion(location)

		// Check if location is valid (either as normalized name or display name)
		isValid := false

		// First check if it's a supported ArcBox region (normalized)
		if _, exists := normalizedToDisplay[normalized]; exists {
			isValid = true
		}

		// Then check if it's a display name that maps to a supported region
		if !isValid {
			if normalizedFromDisplay, exists := displayToNormalized[location]; exists {
				if _, supported := normalizedToDisplay[normalizedFromDisplay]; supported {
					isValid = true
				}
			}
		}

		if !isValid {
			invalidLocations = append(invalidLocations, location)

			// Try to suggest a similar region
			suggestion := utils.SuggestSimilarCommand(normalized, getSupportedRegionsList(normalizedToDisplay), 3)
			if suggestion != "" {
				if displayName, exists := normalizedToDisplay[suggestion]; exists {
					suggestions = append(suggestions, fmt.Sprintf("'%s' → '%s' (%s)", location, suggestion, displayName))
				}
			}
		}
	}

	if len(invalidLocations) > 0 {
		errorMsg := fmt.Sprintf("invalid location(s): %s", strings.Join(invalidLocations, ", "))

		if len(suggestions) > 0 {
			errorMsg += "\n\nDid you mean:\n  " + strings.Join(suggestions, "\n  ")
		}

		errorMsg += fmt.Sprintf("\n\nSupported ArcBox regions:\n  %s", strings.Join(getSupportedRegionsDisplayList(normalizedToDisplay), "\n  "))

		return fmt.Errorf("%s", errorMsg)
	}

	return nil
}

// getSupportedRegionsList returns a list of supported region names (normalized)
func getSupportedRegionsList(normalizedToDisplay map[string]string) []string {
	var regions []string
	for normalized := range normalizedToDisplay {
		regions = append(regions, normalized)
	}
	return regions
}

// getSupportedRegionsDisplayList returns a list of supported region display names
func getSupportedRegionsDisplayList(normalizedToDisplay map[string]string) []string {
	var regions []string
	for _, display := range normalizedToDisplay {
		regions = append(regions, display)
	}
	return regions
}
