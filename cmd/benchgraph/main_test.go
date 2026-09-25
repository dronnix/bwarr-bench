package main

import (
	"math"
	"testing"
	"time"
)

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

func TestDurationToMillis(t *testing.T) {
	tests := []struct {
		d        time.Duration
		expected float64
	}{
		{0, 0},
		{time.Millisecond, 1},
		{812039 * time.Nanosecond, 0.812039}, // sub-millisecond must not become 0
		{1500 * time.Microsecond, 1.5},
		{2 * time.Second, 2000},
	}

	for _, tt := range tests {
		got := durationToMillis(tt.d)
		if math.Abs(got-tt.expected) > 1e-9 {
			t.Errorf("durationToMillis(%v) = %v, want %v", tt.d, got, tt.expected)
		}
	}
}
