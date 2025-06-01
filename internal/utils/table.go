package utils

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// PrintASCIITable prints a table with dynamic ASCII borders based on headers and rows.
// Optionally, right-aligns the 4th column (SKU status/emoji) for ArcBox and similar tables.
// headers: slice of column names
// rows: slice of rows, each row is a slice of strings
func PrintASCIITable(headers []string, rows [][]string) {
	// Calculate max width for each column (strip ANSI and calculate visual width)
	colWidths := make([]int, len(headers))
	for i, h := range headers {
		colWidths[i] = visualWidth(stripANSI(h))
	}
	for _, row := range rows {
		for i, cell := range row {
			w := visualWidth(stripANSI(cell))
			if w > colWidths[i] {
				colWidths[i] = w
			}
		}
	}
	// Print top border
	border := "+"
	for _, w := range colWidths {
		border += strings.Repeat("-", w+2) + "+"
	}
	fmt.Println(border)

	// Print header
	fmt.Print("|")
	for i, h := range headers {
		fmt.Printf(" %-*s |", colWidths[i], h)
	}
	fmt.Println()

	// Print header separator
	fmt.Println(border)

	// Print rows
	for _, row := range rows {
		fmt.Print("|")
		for i, cell := range row {
			cellStripped := stripANSI(cell)
			visualPad := colWidths[i] - visualWidth(cellStripped)
			fmt.Printf(" %-s%s |", cell, strings.Repeat(" ", visualPad))
		}
		fmt.Println()
	}

	// Print bottom border
	fmt.Println(border)
}

// Helper to strip ANSI color codes
var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiRegexp.ReplaceAllString(s, "")
}

// visualWidth calculates the visual/display width of a string, accounting for:
// - Wide Unicode characters (like emojis) that take 2 character widths
// - Zero-width characters
// - Normal ASCII characters that take 1 character width
func visualWidth(s string) int {
	if s == "" {
		return 0
	}

	width := 0
	for _, r := range s {
		switch {
		case r < 32:
			// Control characters (invisible)
			continue
		case r < 127:
			// ASCII printable characters
			width++
		case unicode.Is(unicode.Mn, r):
			// Non-spacing marks (combining characters) - zero width
			continue
		case unicode.Is(unicode.Me, r):
			// Enclosing marks - zero width
			continue
		case unicode.Is(unicode.Cf, r):
			// Format characters - zero width
			continue
		case isWideCharacter(r):
			// Wide characters (emojis, CJK characters, etc.) - 2 character widths
			width += 2
		default:
			// Regular Unicode characters - 1 character width
			width++
		}
	}
	return width
}

// isWideCharacter determines if a rune should be treated as wide (2 character widths)
func isWideCharacter(r rune) bool {
	// Check for common emoji ranges
	if (r >= 0x1F600 && r <= 0x1F64F) || // Emoticons
		(r >= 0x1F300 && r <= 0x1F5FF) || // Misc Symbols and Pictographs
		(r >= 0x1F680 && r <= 0x1F6FF) || // Transport and Map
		(r >= 0x1F700 && r <= 0x1F77F) || // Alchemical Symbols
		(r >= 0x1F780 && r <= 0x1F7FF) || // Geometric Shapes Extended
		(r >= 0x1F800 && r <= 0x1F8FF) || // Supplemental Arrows-C
		(r >= 0x1F900 && r <= 0x1F9FF) || // Supplemental Symbols and Pictographs
		(r >= 0x1FA00 && r <= 0x1FA6F) || // Chess Symbols
		(r >= 0x1FA70 && r <= 0x1FAFF) || // Symbols and Pictographs Extended-A
		(r >= 0x2600 && r <= 0x26FF) || // Miscellaneous Symbols
		(r >= 0x2700 && r <= 0x27BF) || // Dingbats
		(r >= 0xFE00 && r <= 0xFE0F) || // Variation Selectors
		(r >= 0x1F000 && r <= 0x1F02F) || // Mahjong Tiles
		(r >= 0x1F0A0 && r <= 0x1F0FF) { // Playing Cards
		return true
	}

	// Check for CJK characters (Chinese, Japanese, Korean)
	if (r >= 0x1100 && r <= 0x115F) || // Hangul Jamo
		(r >= 0x2E80 && r <= 0x2EFF) || // CJK Radicals Supplement
		(r >= 0x2F00 && r <= 0x2FDF) || // Kangxi Radicals
		(r >= 0x3000 && r <= 0x303F) || // CJK Symbols and Punctuation
		(r >= 0x3040 && r <= 0x309F) || // Hiragana
		(r >= 0x30A0 && r <= 0x30FF) || // Katakana
		(r >= 0x3100 && r <= 0x312F) || // Bopomofo
		(r >= 0x3130 && r <= 0x318F) || // Hangul Compatibility Jamo
		(r >= 0x31A0 && r <= 0x31BF) || // Bopomofo Extended
		(r >= 0x31F0 && r <= 0x31FF) || // Katakana Phonetic Extensions
		(r >= 0x3200 && r <= 0x32FF) || // Enclosed CJK Letters and Months
		(r >= 0x3300 && r <= 0x33FF) || // CJK Compatibility
		(r >= 0x3400 && r <= 0x4DBF) || // CJK Unified Ideographs Extension A
		(r >= 0x4E00 && r <= 0x9FFF) || // CJK Unified Ideographs
		(r >= 0xA960 && r <= 0xA97F) || // Hangul Jamo Extended-A
		(r >= 0xAC00 && r <= 0xD7AF) || // Hangul Syllables
		(r >= 0xD7B0 && r <= 0xD7FF) || // Hangul Jamo Extended-B
		(r >= 0xF900 && r <= 0xFAFF) || // CJK Compatibility Ideographs
		(r >= 0xFE10 && r <= 0xFE1F) || // Vertical Forms
		(r >= 0xFE30 && r <= 0xFE4F) || // CJK Compatibility Forms
		(r >= 0xFF00 && r <= 0xFFEF) { // Halfwidth and Fullwidth Forms
		return true
	}

	// Additional check for high surrogate pairs and other wide characters
	if r >= 0x20000 && r <= 0x2FFFD { // CJK Unified Ideographs Extension B, C, D, E
		return true
	}

	return false
}
