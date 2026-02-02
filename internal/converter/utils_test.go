package converter

import "testing"

func TestParseSize(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		wantErr  bool
	}{
		{"200kb", 200 * 1024, false},
		{"1mb", 1024 * 1024, false},
		{"1.5mb", 1572864, false},
		{"500b", 500, false},
		{"100", 100, false},
		{" 200 KB ", 200 * 1024, false},
		{"abc", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		got, err := ParseSize(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseSize(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if got != tt.expected {
			t.Errorf("ParseSize(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}
