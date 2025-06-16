package utils

import (
	"testing"
	"time"
)

func TestParseISO8601Duration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
		wantErr  bool
	}{
		// Valid duration formats
		{
			name:     "hours_only",
			input:    "PT2H",
			expected: 2 * time.Hour,
			wantErr:  false,
		},
		{
			name:     "minutes_only",
			input:    "PT30M",
			expected: 30 * time.Minute,
			wantErr:  false,
		},
		{
			name:     "seconds_only",
			input:    "PT45S",
			expected: 45 * time.Second,
			wantErr:  false,
		},
		{
			name:     "hours_and_minutes",
			input:    "PT1H30M",
			expected: 1*time.Hour + 30*time.Minute,
			wantErr:  false,
		},
		{
			name:     "hours_and_seconds",
			input:    "PT2H45S",
			expected: 2*time.Hour + 45*time.Second,
			wantErr:  false,
		},
		{
			name:     "minutes_and_seconds",
			input:    "PT15M30S",
			expected: 15*time.Minute + 30*time.Second,
			wantErr:  false,
		},
		{
			name:     "hours_minutes_seconds",
			input:    "PT1H30M45S",
			expected: 1*time.Hour + 30*time.Minute + 45*time.Second,
			wantErr:  false,
		},
		{
			name:     "fractional_seconds",
			input:    "PT30.5S",
			expected: time.Duration(30.5 * float64(time.Second)),
			wantErr:  false,
		},
		{
			name:     "zero_duration",
			input:    "PT0S",
			expected: 0,
			wantErr:  false,
		},
		{
			name:     "large_values",
			input:    "PT24H59M59S",
			expected: 24*time.Hour + 59*time.Minute + 59*time.Second,
			wantErr:  false,
		},
		{
			name:     "complex_fractional",
			input:    "PT1H15M30.75S",
			expected: 1*time.Hour + 15*time.Minute + time.Duration(30.75*float64(time.Second)),
			wantErr:  false,
		},

		// Invalid formats - missing PT prefix
		{
			name:    "missing_pt_prefix",
			input:   "1H30M",
			wantErr: true,
		},
		{
			name:    "empty_string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "only_pt",
			input:   "PT",
			wantErr: false, // Should return 0 duration
		},

		// Invalid formats - malformed numbers
		{
			name:    "invalid_hour",
			input:   "PTaH",
			wantErr: true,
		},
		{
			name:    "invalid_minute",
			input:   "PTbM",
			wantErr: true,
		},
		{
			name:    "invalid_second",
			input:   "PTcS",
			wantErr: true,
		},
		{
			name:     "negative_values",
			input:    "PT-1H",
			expected: -1 * time.Hour,
			wantErr:  false, // Function allows negative values
		},

		// Edge cases
		{
			name:     "very_small_fractional",
			input:    "PT0.001S",
			expected: time.Millisecond,
			wantErr:  false,
		},
		{
			name:    "mixed_case_letters",
			input:   "pt1h30m45s", // Should fail - case sensitive
			wantErr: true,
		},
		{
			name:     "extra_characters",
			input:    "PT1H30M45SX",
			expected: 1*time.Hour + 30*time.Minute + 45*time.Second,
			wantErr:  false, // Should parse successfully, ignore extra
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseISO8601Duration(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseISO8601Duration(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseISO8601Duration(%q) unexpected error: %v", tt.input, err)
				return
			}

			if result != tt.expected {
				t.Errorf("ParseISO8601Duration(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseInt64(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int
	}{
		// float64 conversions
		{
			name:     "float64_positive",
			input:    float64(42.7),
			expected: 42,
		},
		{
			name:     "float64_zero",
			input:    float64(0),
			expected: 0,
		},
		{
			name:     "float64_negative",
			input:    float64(-15.3),
			expected: -15,
		},
		{
			name:     "float64_large",
			input:    float64(999999.9),
			expected: 999999,
		},

		// int conversions
		{
			name:     "int_positive",
			input:    42,
			expected: 42,
		},
		{
			name:     "int_zero",
			input:    0,
			expected: 0,
		},
		{
			name:     "int_negative",
			input:    -15,
			expected: -15,
		},
		{
			name:     "int_large",
			input:    999999,
			expected: 999999,
		},

		// int64 conversions
		{
			name:     "int64_positive",
			input:    int64(42),
			expected: 42,
		},
		{
			name:     "int64_zero",
			input:    int64(0),
			expected: 0,
		},
		{
			name:     "int64_negative",
			input:    int64(-15),
			expected: -15,
		},
		{
			name:     "int64_large",
			input:    int64(999999),
			expected: 999999,
		},

		// string conversions (valid)
		{
			name:     "string_positive",
			input:    "42",
			expected: 42,
		},
		{
			name:     "string_zero",
			input:    "0",
			expected: 0,
		},
		{
			name:     "string_negative",
			input:    "-15",
			expected: -15,
		},
		{
			name:     "string_large",
			input:    "999999",
			expected: 999999,
		},

		// string conversions (invalid) - should return 0
		{
			name:     "string_invalid",
			input:    "abc",
			expected: 0,
		},
		{
			name:     "string_float",
			input:    "42.5",
			expected: 0, // strconv.Atoi doesn't handle floats
		},
		{
			name:     "string_empty",
			input:    "",
			expected: 0,
		},
		{
			name:     "string_mixed",
			input:    "42abc",
			expected: 0,
		},

		// Unsupported types - should return 0
		{
			name:     "bool_true",
			input:    true,
			expected: 0,
		},
		{
			name:     "bool_false",
			input:    false,
			expected: 0,
		},
		{
			name:     "nil",
			input:    nil,
			expected: 0,
		},
		{
			name:     "slice",
			input:    []int{1, 2, 3},
			expected: 0,
		},
		{
			name:     "map",
			input:    map[string]int{"key": 42},
			expected: 0,
		},
		{
			name:     "struct",
			input:    struct{ Value int }{Value: 42},
			expected: 0,
		},

		// Other numeric types
		{
			name:     "int32",
			input:    int32(42),
			expected: 0, // Not handled explicitly
		},
		{
			name:     "int16",
			input:    int16(42),
			expected: 0, // Not handled explicitly
		},
		{
			name:     "int8",
			input:    int8(42),
			expected: 0, // Not handled explicitly
		},
		{
			name:     "uint",
			input:    uint(42),
			expected: 0, // Not handled explicitly
		},
		{
			name:     "float32",
			input:    float32(42.5),
			expected: 0, // Not handled explicitly
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseInt64(tt.input)
			if result != tt.expected {
				t.Errorf("ParseInt64(%v) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

// Edge case tests for boundary conditions
func TestParseISO8601Duration_EdgeCases(t *testing.T) {
	// Test very large duration
	result, err := ParseISO8601Duration("PT999H999M999S")
	if err != nil {
		t.Errorf("ParseISO8601Duration(large values) unexpected error: %v", err)
	}
	expected := 999*time.Hour + 999*time.Minute + 999*time.Second
	if result != expected {
		t.Errorf("ParseISO8601Duration(large values) = %v, expected %v", result, expected)
	}

	// Test precision with fractional seconds
	result, err = ParseISO8601Duration("PT1.123456789S")
	if err != nil {
		t.Errorf("ParseISO8601Duration(fractional precision) unexpected error: %v", err)
	}
	// Due to floating point precision, we'll check if it's approximately correct
	expected = time.Duration(1.123456789 * float64(time.Second))
	if result < expected-time.Microsecond || result > expected+time.Microsecond {
		t.Errorf("ParseISO8601Duration(fractional precision) = %v, expected approximately %v", result, expected)
	}
}

// Benchmark tests to ensure functions are performant
func BenchmarkParseISO8601Duration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ParseISO8601Duration("PT1H30M45S")
	}
}

func BenchmarkParseInt64_Float64(b *testing.B) {
	val := float64(42.7)
	for i := 0; i < b.N; i++ {
		ParseInt64(val)
	}
}

func BenchmarkParseInt64_String(b *testing.B) {
	val := "42"
	for i := 0; i < b.N; i++ {
		ParseInt64(val)
	}
}

func BenchmarkParseInt64_Int(b *testing.B) {
	val := 42
	for i := 0; i < b.N; i++ {
		ParseInt64(val)
	}
}
