package utils

import (
	"testing"
)

func TestNormalizeFlavorCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// ITPro variations
		{
			name:     "lowercase_itpro",
			input:    "itpro",
			expected: "ITPro",
		},
		{
			name:     "uppercase_itpro",
			input:    "ITPRO",
			expected: "ITPro",
		},
		{
			name:     "mixedcase_itpro",
			input:    "ItPrO",
			expected: "ITPro",
		},
		{
			name:     "hyphenated_itpro",
			input:    "it-pro",
			expected: "ITPro",
		},
		{
			name:     "hyphenated_mixed_itpro",
			input:    "IT-Pro",
			expected: "ITPro",
		},

		// DevOps variations
		{
			name:     "lowercase_devops",
			input:    "devops",
			expected: "DevOps",
		},
		{
			name:     "uppercase_devops",
			input:    "DEVOPS",
			expected: "DevOps",
		},
		{
			name:     "mixedcase_devops",
			input:    "DevOps",
			expected: "DevOps",
		},
		{
			name:     "hyphenated_devops",
			input:    "dev-ops",
			expected: "DevOps",
		},
		{
			name:     "hyphenated_mixed_devops",
			input:    "DEV-OPS",
			expected: "DevOps",
		},

		// DataOps variations
		{
			name:     "lowercase_dataops",
			input:    "dataops",
			expected: "DataOps",
		},
		{
			name:     "uppercase_dataops",
			input:    "DATAOPS",
			expected: "DataOps",
		},
		{
			name:     "mixedcase_dataops",
			input:    "DataOps",
			expected: "DataOps",
		},
		{
			name:     "hyphenated_dataops",
			input:    "data-ops",
			expected: "DataOps",
		},
		{
			name:     "hyphenated_mixed_dataops",
			input:    "DATA-OPS",
			expected: "DataOps",
		},

		// Special case for 'all'
		{
			name:     "all_lowercase",
			input:    "all",
			expected: "all",
		},
		{
			name:     "all_uppercase",
			input:    "ALL",
			expected: "all",
		},
		{
			name:     "all_mixed",
			input:    "All",
			expected: "all",
		},

		// Edge cases
		{
			name:     "empty_string",
			input:    "",
			expected: "",
		},
		{
			name:     "unknown_flavor",
			input:    "unknown",
			expected: "unknown",
		},
		{
			name:     "numeric_input",
			input:    "123",
			expected: "123",
		},
		{
			name:     "special_chars",
			input:    "flavor@123",
			expected: "flavor@123",
		},
		{
			name:     "whitespace",
			input:    " itpro ",
			expected: " itpro ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeFlavorCase(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeFlavorCase(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeSqlServerEditionCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Developer variations
		{
			name:     "lowercase_developer",
			input:    "developer",
			expected: "Developer",
		},
		{
			name:     "uppercase_developer",
			input:    "DEVELOPER",
			expected: "Developer",
		},
		{
			name:     "mixedcase_developer",
			input:    "Developer",
			expected: "Developer",
		},
		{
			name:     "mixedcase_dev",
			input:    "DeVeLoPeR",
			expected: "Developer",
		},

		// Standard variations
		{
			name:     "lowercase_standard",
			input:    "standard",
			expected: "Standard",
		},
		{
			name:     "uppercase_standard",
			input:    "STANDARD",
			expected: "Standard",
		},
		{
			name:     "mixedcase_standard",
			input:    "Standard",
			expected: "Standard",
		},
		{
			name:     "mixedcase_std",
			input:    "StAnDaRd",
			expected: "Standard",
		},

		// Enterprise variations
		{
			name:     "lowercase_enterprise",
			input:    "enterprise",
			expected: "Enterprise",
		},
		{
			name:     "uppercase_enterprise",
			input:    "ENTERPRISE",
			expected: "Enterprise",
		},
		{
			name:     "mixedcase_enterprise",
			input:    "Enterprise",
			expected: "Enterprise",
		},
		{
			name:     "mixedcase_ent",
			input:    "EnTeRpRiSe",
			expected: "Enterprise",
		},

		// Edge cases
		{
			name:     "empty_string",
			input:    "",
			expected: "",
		},
		{
			name:     "unknown_edition",
			input:    "unknown",
			expected: "unknown",
		},
		{
			name:     "numeric_input",
			input:    "123",
			expected: "123",
		},
		{
			name:     "special_chars",
			input:    "edition@123",
			expected: "edition@123",
		},
		{
			name:     "whitespace",
			input:    " developer ",
			expected: " developer ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeSqlServerEditionCase(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeSqlServerEditionCase(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeBastionSkuCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Basic variations
		{
			name:     "lowercase_basic",
			input:    "basic",
			expected: "Basic",
		},
		{
			name:     "uppercase_basic",
			input:    "BASIC",
			expected: "Basic",
		},
		{
			name:     "mixedcase_basic",
			input:    "Basic",
			expected: "Basic",
		},
		{
			name:     "mixedcase_bas",
			input:    "BaSiC",
			expected: "Basic",
		},

		// Standard variations
		{
			name:     "lowercase_standard",
			input:    "standard",
			expected: "Standard",
		},
		{
			name:     "uppercase_standard",
			input:    "STANDARD",
			expected: "Standard",
		},
		{
			name:     "mixedcase_standard",
			input:    "Standard",
			expected: "Standard",
		},
		{
			name:     "mixedcase_std",
			input:    "StAnDaRd",
			expected: "Standard",
		},

		// Developer variations
		{
			name:     "lowercase_developer",
			input:    "developer",
			expected: "Developer",
		},
		{
			name:     "uppercase_developer",
			input:    "DEVELOPER",
			expected: "Developer",
		},
		{
			name:     "mixedcase_developer",
			input:    "Developer",
			expected: "Developer",
		},
		{
			name:     "mixedcase_dev",
			input:    "DeVeLoPeR",
			expected: "Developer",
		},

		// Edge cases
		{
			name:     "empty_string",
			input:    "",
			expected: "",
		},
		{
			name:     "unknown_sku",
			input:    "premium",
			expected: "premium",
		},
		{
			name:     "numeric_input",
			input:    "123",
			expected: "123",
		},
		{
			name:     "special_chars",
			input:    "sku@123",
			expected: "sku@123",
		},
		{
			name:     "whitespace",
			input:    " basic ",
			expected: " basic ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeBastionSkuCase(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeBastionSkuCase(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Benchmark tests to ensure functions are performant
func BenchmarkNormalizeFlavorCase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NormalizeFlavorCase("itpro")
	}
}

func BenchmarkNormalizeSqlServerEditionCase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NormalizeSqlServerEditionCase("developer")
	}
}

func BenchmarkNormalizeBastionSkuCase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NormalizeBastionSkuCase("basic")
	}
}
