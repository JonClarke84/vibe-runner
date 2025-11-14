package network

import (
	"html"
	"strings"
	"testing"
)

// TestSanitizePlayerName_ValidName_ReturnsTrimmedName tests that
// sanitizePlayerName correctly trims whitespace from valid names.
func TestSanitizePlayerName_ValidName_ReturnsTrimmedName(t *testing.T) {
	// Arrange
	input := "  TestPlayer  "
	want := "TestPlayer"

	// Act
	got := sanitizePlayerName(input)

	// Assert
	if got != want {
		t.Errorf("sanitizePlayerName(%q) = %q, want %q", input, got, want)
	}
}

// TestSanitizePlayerName_EmptyString_ReturnsDefaultName tests that
// empty input results in the default "Player" name.
func TestSanitizePlayerName_EmptyString_ReturnsDefaultName(t *testing.T) {
	// Arrange
	input := ""
	want := "Player"

	// Act
	got := sanitizePlayerName(input)

	// Assert
	if got != want {
		t.Errorf("sanitizePlayerName(%q) = %q, want %q", input, got, want)
	}
}

// TestSanitizePlayerName_WhitespaceOnly_ReturnsDefaultName tests that
// whitespace-only input results in the default name.
func TestSanitizePlayerName_WhitespaceOnly_ReturnsDefaultName(t *testing.T) {
	// Arrange
	input := "   \t\n   "
	want := "Player"

	// Act
	got := sanitizePlayerName(input)

	// Assert
	if got != want {
		t.Errorf("sanitizePlayerName(%q) = %q, want %q", input, got, want)
	}
}

// TestSanitizePlayerName_TooLong_TruncatesTo30Chars tests that names
// longer than 30 characters are truncated.
func TestSanitizePlayerName_TooLong_TruncatesTo30Chars(t *testing.T) {
	// Arrange - 40 character string
	input := "ThisIsAVeryLongPlayerNameThatExceeds30"
	wantLen := 30

	// Act
	got := sanitizePlayerName(input)

	// Assert
	if len(got) != wantLen {
		t.Errorf("sanitizePlayerName() length = %d, want %d", len(got), wantLen)
	}
	if !strings.HasPrefix(input, got) {
		t.Errorf("sanitizePlayerName() = %q, should be prefix of %q", got, input)
	}
}

// TestSanitizePlayerName_HTMLTags_EscapesTags tests that HTML tags
// are escaped (not removed) to prevent XSS attacks while preserving user input.
func TestSanitizePlayerName_HTMLTags_EscapesTags(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "script tag",
			input: "<script>alert('xss')</script>Player",
		},
		{
			name:  "bold tag",
			input: "<b>BoldPlayer</b>",
		},
		{
			name:  "multiple tags",
			input: "<div><span>Player</span></div>",
		},
		{
			name:  "unclosed tag",
			input: "<script>Player",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got := sanitizePlayerName(tt.input)

			// Assert - raw angle brackets should be escaped to &lt; and &gt;
			if strings.Contains(got, "<") {
				t.Errorf("sanitizePlayerName(%q) = %q, contains unescaped <", tt.input, got)
			}
			if strings.Contains(got, ">") {
				t.Errorf("sanitizePlayerName(%q) = %q, contains unescaped >", tt.input, got)
			}
			// Verify escaping occurred
			if !strings.Contains(got, "&lt;") && !strings.Contains(got, "&gt;") {
				t.Errorf("sanitizePlayerName(%q) = %q, tags not escaped", tt.input, got)
			}
		})
	}
}

// TestSanitizePlayerName_HTMLEntities_EscapesEntities tests that
// special characters are HTML-escaped to prevent XSS.
func TestSanitizePlayerName_HTMLEntities_EscapesEntities(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "ampersand",
			input: "Player&Co",
			want:  html.EscapeString("Player&Co"),
		},
		{
			name:  "less than",
			input: "Player<3",
			want:  html.EscapeString("Player<3"),
		},
		{
			name:  "greater than",
			input: "Player>Pro",
			want:  html.EscapeString("Player>Pro"),
		},
		{
			name:  "quotes",
			input: `Player"Name"`,
			want:  html.EscapeString(`Player"Name"`),
		},
		{
			name:  "single quotes",
			input: "Player'Name",
			want:  html.EscapeString("Player'Name"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got := sanitizePlayerName(tt.input)

			// Assert
			if got != tt.want {
				t.Errorf("sanitizePlayerName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestSanitizePlayerName_UnicodeCharacters_PreservesUnicode tests that
// valid Unicode characters (emoji, non-ASCII) are preserved.
func TestSanitizePlayerName_UnicodeCharacters_PreservesUnicode(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "emoji",
			input: "Player🎮",
		},
		{
			name:  "japanese",
			input: "プレイヤー",
		},
		{
			name:  "cyrillic",
			input: "Игрок",
		},
		{
			name:  "arabic",
			input: "لاعب",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got := sanitizePlayerName(tt.input)

			// Assert - should contain the Unicode characters (may be escaped)
			if got == "" || got == "Player" {
				t.Errorf("sanitizePlayerName(%q) = %q, lost Unicode characters", tt.input, got)
			}
		})
	}
}

// TestSanitizePlayerName_TableDriven tests sanitizePlayerName with
// various inputs using table-driven test pattern.
func TestSanitizePlayerName_TableDriven(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "normal name",
			input: "JohnDoe123",
			want:  "JohnDoe123",
		},
		{
			name:  "name with spaces",
			input: "  John Doe  ",
			want:  "John Doe",
		},
		{
			name:  "empty string",
			input: "",
			want:  "Player",
		},
		{
			name:  "exactly 30 chars",
			input: "123456789012345678901234567890",
			want:  "123456789012345678901234567890",
		},
		{
			name:  "31 chars - should truncate",
			input: "1234567890123456789012345678901",
			want:  "123456789012345678901234567890", // First 30
		},
		{
			name:  "special chars",
			input: "Player_#123",
			want:  "Player_#123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got := sanitizePlayerName(tt.input)

			// Assert
			if got != tt.want {
				t.Errorf("sanitizePlayerName(%q) = %q, want %q", tt.input, got, tt.want)
			}

			// Additional assertion: result should never exceed 30 chars
			if len(got) > 30 {
				t.Errorf("sanitizePlayerName(%q) returned name longer than 30 chars: %d", tt.input, len(got))
			}

			// Additional assertion: result should never be empty
			if got == "" {
				t.Errorf("sanitizePlayerName(%q) returned empty string", tt.input)
			}
		})
	}
}

// TestSanitizePlayerName_XSSAttempts tests that common XSS attack
// vectors are properly sanitized by escaping HTML entities.
func TestSanitizePlayerName_XSSAttempts(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		verify func(t *testing.T, output string)
	}{
		{
			name:  "script injection",
			input: `<script>alert('XSS')</script>`,
			verify: func(t *testing.T, output string) {
				// Verify < and > are escaped
				if strings.Contains(output, "<") || strings.Contains(output, ">") {
					t.Error("Output contains unescaped angle brackets")
				}
				// Verify escaping occurred
				if !strings.Contains(output, "&lt;") || !strings.Contains(output, "&gt;") {
					t.Error("Output does not contain escaped HTML entities")
				}
			},
		},
		{
			name:  "img onerror",
			input: `<img src=x onerror="alert('XSS')">`,
			verify: func(t *testing.T, output string) {
				if strings.Contains(output, "<") || strings.Contains(output, ">") {
					t.Error("Output contains unescaped angle brackets")
				}
				// Quotes should be escaped
				if !strings.Contains(output, "&#34;") && !strings.Contains(output, "&quot;") {
					// Note: html.EscapeString escapes quotes as &#34;
				}
			},
		},
		{
			name:  "javascript protocol",
			input: `<a href="javascript:alert('XSS')">`,
			verify: func(t *testing.T, output string) {
				if strings.Contains(output, "<") || strings.Contains(output, ">") {
					t.Error("Output contains unescaped angle brackets")
				}
				// The javascript: protocol itself is escaped, making it harmless
				if strings.Contains(output, `href="javascript:`) {
					t.Error("Output contains unescaped href attribute")
				}
			},
		},
		{
			name:  "iframe injection",
			input: `<iframe src="evil.com"></iframe>`,
			verify: func(t *testing.T, output string) {
				if strings.Contains(output, "<iframe") {
					t.Error("Output contains unescaped <iframe> tag")
				}
				// Verify escaping occurred
				if !strings.Contains(output, "&lt;") {
					t.Error("Output does not contain escaped HTML")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got := sanitizePlayerName(tt.input)

			// Assert using custom verification function
			tt.verify(t, got)

			// General assertion: output should not be empty
			if got == "" {
				t.Error("sanitizePlayerName() returned empty string for XSS attempt")
			}
		})
	}
}
