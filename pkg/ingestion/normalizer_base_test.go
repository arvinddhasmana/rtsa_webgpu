// CLASSIFICATION: UNCLASSIFIED
package ingestion_test

import (
	"testing"

	"github.com/arvinddhasmana/RTSA_VS_Opus/pkg/ingestion"
)

func TestTrimString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already trimmed",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "leading and trailing spaces",
			input:    "  hello world  ",
			expected: "hello world",
		},
		{
			name:     "tabs and newlines",
			input:    "\t\nhello world\r\n",
			expected: "hello world",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only spaces",
			input:    "     ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ingestion.TrimString(tt.input)
			if result != tt.expected {
				t.Errorf("TrimString(%q) expected %q, got %q", tt.input, tt.expected, result)
			}
		})
	}
}
