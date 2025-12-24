package main

import "testing"

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Insert Performance", "insert_performance"},
		{"Get all values by key", "get_all_values_by_key"},
		{"Delete All Values", "delete_all_values"},
		{"Simple", "simple"},
		{"WITH-DASHES", "withdashes"},
		{"With Spaces", "with_spaces"},
		{"Multiple   Spaces", "multiple___spaces"},
		{"CamelCase", "camelcase"},
		{"with123numbers", "with123numbers"},
		{"123start", "123start"},
	}

	for _, tt := range tests {
		result := sanitizeFilename(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestSanitizeFilename_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"only spaces", "   ", "___"},
		{"only special chars", "!@#$%^&*()", ""},
		{"underscores preserved", "test_file_name", "test_file_name"},
		{"mixed case with specials", "Test-File!Name", "testfilename"},
		{"unicode characters", "Test™File®", "testfile"},
		{"numbers only", "12345", "12345"},
		{"underscores only", "___", "___"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
