package utils

import (
	"fmt"
	"strings"
	"testing"
)

func TestPrintASCIITable(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing ASCII Table Printing ==="))
	
	t.Run("basic_table", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing basic table"))
		headers := []string{"Name", "Status", "Location"}
		rows := [][]string{
			{"Resource1", "✅ Running", "East US"},
			{"Resource2", "❌ Stopped", "West US"},
			{"Resource3", "⚠️ Warning", "Central US"},
		}

		// Capture the output by redirecting stdout temporarily
		// Since PrintASCIITable prints directly, we'll test that it doesn't panic
		defer func() {
			if r := recover(); r != nil {
				printTestStatus(t, "Basic Table", false, fmt.Sprintf("PrintASCIITable panicked: %v", r))
				return
			}
			printTestStatus(t, "Basic Table", true, "Table printed without panic")
		}()

		PrintASCIITable(headers, rows)
	})

	t.Run("empty_table", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing empty table"))
		headers := []string{"Column1", "Column2"}
		rows := [][]string{}

		defer func() {
			if r := recover(); r != nil {
				printTestStatus(t, "Empty Table", false, fmt.Sprintf("PrintASCIITable with empty rows panicked: %v", r))
				return
			}
			printTestStatus(t, "Empty Table", true, "Empty table handled correctly")
		}()

		PrintASCIITable(headers, rows)
	})

	t.Run("single_row", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing single row table"))
		headers := []string{"Name"}
		rows := [][]string{{"Single Item"}}

		defer func() {
			if r := recover(); r != nil {
				printTestStatus(t, "Single Row Table", false, fmt.Sprintf("PrintASCIITable with single row panicked: %v", r))
				return
			}
			printTestStatus(t, "Single Row Table", true, "Single row table handled correctly")
		}()

		PrintASCIITable(headers, rows)
	})
	t.Run("unicode_content", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing unicode content"))
		headers := []string{"Service", "Status", "Emoji"}
		rows := [][]string{
			{"Service 1", "Active", "🟢"},
			{"Service 2", "Inactive", "🔴"},
			{"Long Service Name", "Pending", "🟡"},
		}

		defer func() {
			if r := recover(); r != nil {
				printTestStatus(t, "Unicode Content Table", false, fmt.Sprintf("PrintASCIITable with unicode content panicked: %v", r))
				return
			}
			printTestStatus(t, "Unicode Content Table", true, "Unicode content handled correctly")
		}()

		PrintASCIITable(headers, rows)
	})

	t.Run("varying_column_widths", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing varying column widths"))
		headers := []string{"Short", "Very Long Header Name", "Mid"}
		rows := [][]string{
			{"A", "Short content", "Medium length"},
			{"Very long content here", "B", "C"},
			{"X", "Y", "Very very long content that exceeds header"},
		}

		defer func() {
			if r := recover(); r != nil {
				printTestStatus(t, "Varying Column Widths", false, fmt.Sprintf("PrintASCIITable with varying widths panicked: %v", r))
				return
			}
			printTestStatus(t, "Varying Column Widths", true, "Varying column widths handled correctly")
		}()

		PrintASCIITable(headers, rows)
	})
}

func TestStripANSI(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing ANSI Stripping ==="))
	
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no_ansi",
			input:    "Plain text",
			expected: "Plain text",
		},
		{
			name:     "simple_color",
			input:    "\x1b[31mRed text\x1b[0m",
			expected: "Red text",
		},
		{
			name:     "multiple_colors",
			input:    "\x1b[31mRed\x1b[0m and \x1b[32mGreen\x1b[0m",
			expected: "Red and Green",
		},
		{
			name:     "complex_ansi",
			input:    "\x1b[1;31;40mBold red on black\x1b[0m",
			expected: "Bold red on black",
		},
		{
			name:     "empty_string",
			input:    "",
			expected: "",
		},
		{
			name:     "only_ansi",
			input:    "\x1b[31m\x1b[0m",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := stripANSI(tc.input)
			success := result == tc.expected
			printTestStatus(t, fmt.Sprintf("ANSI Strip: %s", tc.name), success,
				fmt.Sprintf("Expected '%s', got '%s'", tc.expected, result))
		})
	}
}

func TestVisualWidth(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Visual Width Calculation ==="))
	
	testCases := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "empty_string",
			input:    "",
			expected: 0,
		},
		{
			name:     "ascii_text",
			input:    "Hello",
			expected: 5,
		},
		{
			name:     "emoji_single",
			input:    "🔴",
			expected: 2, // Emojis are typically 2 character widths
		},
		{
			name:     "mixed_ascii_emoji",
			input:    "Status: ✅",
			expected: 10, // "Status: " (8) + "✅" (2)
		},
		{
			name:     "multiple_emojis",
			input:    "🟢🟡🔴",
			expected: 6, // 3 emojis × 2 width each
		},
		{
			name:     "ascii_numbers",
			input:    "12345",
			expected: 5,
		},
		{
			name:     "spaces",
			input:    "   ",
			expected: 3,
		},
		{
			name:     "mixed_with_spaces",
			input:    "Test 🔴 End",
			expected: 11, // "Test " (5) + "🔴" (2) + " End" (4)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := visualWidth(tc.input)
			success := result == tc.expected
			printTestStatus(t, fmt.Sprintf("Visual Width: %s", tc.name), success,
				fmt.Sprintf("Expected %d, got %d for '%s'", tc.expected, result, tc.input))
		})
	}
}

func TestIsWideCharacter(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Wide Character Detection ==="))
	
	testCases := []struct {
		name     string
		char     rune
		expected bool
	}{
		{
			name:     "ascii_letter",
			char:     'A',
			expected: false,
		},
		{
			name:     "ascii_digit",
			char:     '5',
			expected: false,
		},
		{
			name:     "check_mark_emoji",
			char:     '✅',
			expected: true,
		},
		{
			name:     "red_circle_emoji",
			char:     '🔴',
			expected: true,
		},
		{
			name:     "space",
			char:     ' ',
			expected: false,
		},
		{
			name:     "chinese_character",
			char:     '中',
			expected: true,
		},
		{
			name:     "japanese_hiragana",
			char:     'あ',
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isWideCharacter(tc.char)
			success := result == tc.expected
			printTestStatus(t, fmt.Sprintf("Wide Char: %s", tc.name), success,
				fmt.Sprintf("Character '%c' should be wide=%t, got %t", tc.char, tc.expected, result))
		})
	}
}

func TestTableConsistency(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Table Function Consistency ==="))
	
	// Test that stripANSI and visualWidth work together correctly
	t.Run("ansi_and_visual_width", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing ANSI stripping with visual width"))
		ansiText := "\x1b[31mRed text\x1b[0m"
		stripped := stripANSI(ansiText)
		width := visualWidth(stripped)
		
		expectedWidth := len("Red text")
		success := width == expectedWidth
		printTestStatus(t, "ANSI + Visual Width", success,
			fmt.Sprintf("Visual width should be %d, got %d", expectedWidth, width))
	})

	t.Run("emoji_consistency", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing emoji visual width"))
		emojiText := "Status: ✅"
		width := visualWidth(emojiText)
		
		// "Status: " = 8 characters, "✅" = 2 character widths
		expectedWidth := 10
		success := width == expectedWidth
		printTestStatus(t, "Emoji Width Consistency", success,
			fmt.Sprintf("Visual width should be %d, got %d", expectedWidth, width))
	})
}

func TestTableHelperFunctions(t *testing.T) {
	fmt.Printf("\n%s\n", testHeaderColor("=== Testing Table Helper Functions ==="))
	
	t.Run("ansi_regexp_validity", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing ANSI regexp validity"))
		// Test that the ANSI regexp is valid and doesn't panic
		testString := "\x1b[31mTest\x1b[0m"
		result := ansiRegexp.ReplaceAllString(testString, "")
		
		success := result == "Test"
		printTestStatus(t, "ANSI Regexp", success, 
			fmt.Sprintf("Expected 'Test', got '%s'", result))
	})

	t.Run("edge_cases", func(t *testing.T) {
		fmt.Printf("    %s %s\n", testInfoColor("→"), testInfoColor("Testing edge cases"))
		// Test edge cases for visual width calculation
		edgeCases := []string{
			"\x00\x01\x02", // Control characters
			"\u200B\u200C", // Zero-width characters
			"Normal text",   // Regular text
		}

		for _, testCase := range edgeCases {
			// Should not panic
			defer func() {
				if r := recover(); r != nil {
					printTestStatus(t, fmt.Sprintf("Edge case: %q", testCase), false,
						fmt.Sprintf("visualWidth panicked: %v", r))
					return
				}
			}()
			
			_ = visualWidth(testCase)
		}
		printTestStatus(t, "Edge Cases", true, "All edge cases handled without panic")
	})
}

// Benchmark tests for performance-critical functions
func BenchmarkVisualWidth(b *testing.B) {
	testStrings := []string{
		"Simple ASCII text",
		"Text with emoji 🔴 mixed in",
		"\x1b[31mANSI colored text\x1b[0m",
		"Mixed: ASCII + 中文 + 🎉 + \x1b[32mcolor\x1b[0m",
	}

	for _, s := range testStrings {
		b.Run("string_"+strings.ReplaceAll(s, " ", "_"), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				visualWidth(s)
			}
		})
	}
}

func BenchmarkStripANSI(b *testing.B) {
	testString := "\x1b[1;31;40mComplex ANSI\x1b[0m with \x1b[32mmultiple\x1b[0m \x1b[33mcolors\x1b[0m"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stripANSI(testString)
	}
}
