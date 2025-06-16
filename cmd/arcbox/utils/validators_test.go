package utils

import (
	"strings"
	"testing"
)

func TestValidateLocations(t *testing.T) {
	tests := []struct {
		name        string
		locations   []string
		wantErr     bool
		errContains []string // Expected content in error message
	}{
		// Valid cases
		{
			name:      "valid_single_display_name",
			locations: []string{"East US"},
			wantErr:   false,
		},
		{
			name:      "valid_single_normalized_name",
			locations: []string{"eastus"},
			wantErr:   false,
		},
		{
			name:      "valid_multiple_regions",
			locations: []string{"East US", "West US 2", "North Europe"},
			wantErr:   false,
		},
		{
			name:      "valid_mixed_formats",
			locations: []string{"eastus", "West US 2", "northeurope"},
			wantErr:   false,
		},
		{
			name:      "valid_with_whitespace",
			locations: []string{" East US ", " westus2 "},
			wantErr:   false,
		},
		{
			name:      "valid_case_insensitive",
			locations: []string{"EAST US", "west us 2", "North Europe"},
			wantErr:   false,
		},

		// Invalid cases
		{
			name:        "empty_locations_array",
			locations:   []string{},
			wantErr:     true,
			errContains: []string{"no locations provided"},
		},
		{
			name:        "single_invalid_location",
			locations:   []string{"Invalid Region"},
			wantErr:     true,
			errContains: []string{"invalid location(s): Invalid Region"},
		},
		{
			name:        "multiple_invalid_locations",
			locations:   []string{"Invalid Region 1", "Invalid Region 2"},
			wantErr:     true,
			errContains: []string{"invalid location(s): Invalid Region 1, Invalid Region 2"},
		},
		{
			name:        "mixed_valid_invalid",
			locations:   []string{"East US", "Invalid Region", "West US 2"},
			wantErr:     true,
			errContains: []string{"invalid location(s): Invalid Region"},
		},
		{
			name:        "typo_in_region_name",
			locations:   []string{"East USA"}, // Common typo
			wantErr:     true,
			errContains: []string{"invalid location(s): East USA", "Did you mean"},
		},
		{
			name:        "partial_region_name",
			locations:   []string{"East"},
			wantErr:     true,
			errContains: []string{"invalid location(s): East"},
		},

		// Edge cases
		{
			name:      "empty_string_in_locations",
			locations: []string{"East US", "", "West US 2"},
			wantErr:   false, // Empty strings should be ignored
		},
		{
			name:      "whitespace_only_location",
			locations: []string{"East US", "   ", "West US 2"},
			wantErr:   false, // Whitespace-only should be ignored after trim
		},
		{
			name:        "numeric_location",
			locations:   []string{"12345"},
			wantErr:     true,
			errContains: []string{"invalid location(s): 12345"},
		},
		{
			name:        "special_characters",
			locations:   []string{"East@US"},
			wantErr:     true,
			errContains: []string{"invalid location(s): East@US"},
		},

		// Test suggestion mechanism
		{
			name:      "close_match_eastus2",
			locations: []string{"eastus2"}, // Maps to "East US 2"
			wantErr:   false,
		},
		{
			name:        "case_variation_typo",
			locations:   []string{"WESTUS3"}, // Non-existent region
			wantErr:     true,
			errContains: []string{"invalid location(s): WESTUS3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLocations(tt.locations)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateLocations(%v) expected error, got nil", tt.locations)
					return
				}

				// Check if error contains expected content
				errMsg := err.Error()
				for _, expected := range tt.errContains {
					if !strings.Contains(errMsg, expected) {
						t.Errorf("ValidateLocations(%v) error = %q, expected to contain %q", tt.locations, errMsg, expected)
					}
				}
			} else {
				if err != nil {
					t.Errorf("ValidateLocations(%v) unexpected error: %v", tt.locations, err)
				}
			}
		})
	}
}

func TestValidateLocations_SupportedRegionsListing(t *testing.T) {
	// Test that error messages include supported regions list
	err := ValidateLocations([]string{"InvalidRegion"})
	if err == nil {
		t.Fatal("Expected error for invalid region")
	}

	errMsg := err.Error()

	// Should contain supported regions section
	if !strings.Contains(errMsg, "Supported ArcBox regions:") {
		t.Error("Error message should contain supported regions list")
	}

	// Should contain some known regions
	expectedRegions := []string{"East US", "West US 2", "North Europe"}
	for _, region := range expectedRegions {
		if !strings.Contains(errMsg, region) {
			t.Errorf("Error message should contain supported region: %s", region)
		}
	}
}

func TestValidateLocations_EdgeCasesWithEmptyAndWhitespace(t *testing.T) {
	tests := []struct {
		name      string
		locations []string
		wantErr   bool
	}{
		{
			name:      "all_empty_strings",
			locations: []string{"", "", ""},
			wantErr:   false, // Empty strings are filtered out, but function doesn't fail
		},
		{
			name:      "all_whitespace",
			locations: []string{"   ", "\t", "\n"},
			wantErr:   false, // Whitespace strings are filtered out, but function doesn't fail
		},
		{
			name:      "mixed_empty_and_valid",
			locations: []string{"", "East US", ""},
			wantErr:   false, // Should pass as we have one valid location
		},
		{
			name:      "mixed_whitespace_and_valid",
			locations: []string{"   ", "East US", "\t"},
			wantErr:   false, // Should pass as we have one valid location
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLocations(tt.locations)

			if tt.wantErr && err == nil {
				t.Errorf("ValidateLocations(%v) expected error, got nil", tt.locations)
			}

			if !tt.wantErr && err != nil {
				t.Errorf("ValidateLocations(%v) unexpected error: %v", tt.locations, err)
			}
		})
	}
}

func TestGetSupportedRegionsList(t *testing.T) {
	// Test the helper function
	testMap := map[string]string{
		"eastus":      "East US",
		"westus2":     "West US 2",
		"northeurope": "North Europe",
	}

	result := getSupportedRegionsList(testMap)

	if len(result) != len(testMap) {
		t.Errorf("getSupportedRegionsList() returned %d items, expected %d", len(result), len(testMap))
	}

	// Check that all normalized names are present
	expectedNormalized := []string{"eastus", "westus2", "northeurope"}
	for _, expected := range expectedNormalized {
		found := false
		for _, actual := range result {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("getSupportedRegionsList() missing expected normalized name: %s", expected)
		}
	}
}

func TestGetSupportedRegionsDisplayList(t *testing.T) {
	// Test the helper function
	testMap := map[string]string{
		"eastus":      "East US",
		"westus2":     "West US 2",
		"northeurope": "North Europe",
	}

	result := getSupportedRegionsDisplayList(testMap)

	if len(result) != len(testMap) {
		t.Errorf("getSupportedRegionsDisplayList() returned %d items, expected %d", len(result), len(testMap))
	}

	// Check that all display names are present
	expectedDisplay := []string{"East US", "West US 2", "North Europe"}
	for _, expected := range expectedDisplay {
		found := false
		for _, actual := range result {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("getSupportedRegionsDisplayList() missing expected display name: %s", expected)
		}
	}
}

// Test performance with large region lists
func TestValidateLocations_Performance(t *testing.T) {
	// Create a large list of valid locations
	largeLocationList := make([]string, 100)
	for i := 0; i < 100; i++ {
		largeLocationList[i] = "East US" // All valid
	}

	err := ValidateLocations(largeLocationList)
	if err != nil {
		t.Errorf("ValidateLocations with large list failed: %v", err)
	}
}

func TestValidateLocations_RealWorldCases(t *testing.T) {
	tests := []struct {
		name      string
		locations []string
		wantErr   bool
		desc      string
	}{
		{
			name:      "common_user_input_1",
			locations: []string{"us-east-1"}, // AWS-style naming
			wantErr:   true,
			desc:      "User might confuse with AWS region naming",
		},
		{
			name:      "common_user_input_2",
			locations: []string{"eastus1"}, // Adding number
			wantErr:   true,
			desc:      "User might think eastus1 exists",
		},
		{
			name:      "common_user_input_3",
			locations: []string{"east"}, // Partial name
			wantErr:   true,
			desc:      "User might try abbreviated names",
		},
		{
			name:      "common_user_input_4",
			locations: []string{"East United States"}, // Verbose name
			wantErr:   true,
			desc:      "User might try full country names",
		},
		{
			name:      "acceptable_variation_1",
			locations: []string{"EAST US"}, // All caps
			wantErr:   false,
			desc:      "Should accept case variations",
		},
		{
			name:      "acceptable_variation_2",
			locations: []string{"east us"}, // All lowercase
			wantErr:   false,
			desc:      "Should accept case variations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLocations(tt.locations)

			if tt.wantErr && err == nil {
				t.Errorf("ValidateLocations(%v) expected error for case: %s", tt.locations, tt.desc)
			}

			if !tt.wantErr && err != nil {
				t.Errorf("ValidateLocations(%v) unexpected error for case: %s. Error: %v", tt.locations, tt.desc, err)
			}
		})
	}
}

// Benchmark test to ensure function is performant
func BenchmarkValidateLocations_SingleValid(b *testing.B) {
	locations := []string{"East US"}
	for i := 0; i < b.N; i++ {
		ValidateLocations(locations)
	}
}

func BenchmarkValidateLocations_MultipleValid(b *testing.B) {
	locations := []string{"East US", "West US 2", "North Europe", "Southeast Asia"}
	for i := 0; i < b.N; i++ {
		ValidateLocations(locations)
	}
}

func BenchmarkValidateLocations_Invalid(b *testing.B) {
	locations := []string{"Invalid Region"}
	for i := 0; i < b.N; i++ {
		ValidateLocations(locations)
	}
}
