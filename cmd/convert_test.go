package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindFilesSanitization(t *testing.T) {
	// Create a temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "gophersnap-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create some files
	filesToCreate := []string{
		"test1.jpg",
		filepath.Join("subdir", "test2.png"),
	}

	for _, f := range filesToCreate {
		fullPath := filepath.Join(tempDir, f)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatalf("Failed to create temp subdir: %v", err)
		}
		err = os.WriteFile(fullPath, []byte("dummy image content"), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	tests := []struct {
		name          string
		inputPath     string
		expectedFiles []string // base names
		expectError   bool
	}{
		{
			name:      "Path with dot dot",
			inputPath: tempDir + string(filepath.Separator) + "subdir" + string(filepath.Separator) + "..",
			expectedFiles: []string{
				"test1.jpg",
				"test2.png",
			},
			expectError: false,
		},
		{
			name:      "Path with dot",
			inputPath: tempDir + string(filepath.Separator) + ".",
			expectedFiles: []string{
				"test1.jpg",
				"test2.png",
			},
			expectError: false,
		},
		{
			name:      "Direct file path",
			inputPath: tempDir + string(filepath.Separator) + "subdir" + string(filepath.Separator) + ".." + string(filepath.Separator) + "test1.jpg",
			expectedFiles: []string{
				"test1.jpg",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := findFiles(tt.inputPath)
			if (err != nil) != tt.expectError {
				t.Fatalf("expected error: %v, got error: %v", tt.expectError, err)
			}

			if len(files) != len(tt.expectedFiles) {
				t.Fatalf("expected %d files, got %d", len(tt.expectedFiles), len(files))
			}

			// Verify basenames
			foundMap := make(map[string]bool)
			for _, f := range files {
				foundMap[filepath.Base(f)] = true
			}

			for _, ef := range tt.expectedFiles {
				if !foundMap[ef] {
					t.Errorf("expected to find file %s, but it was missing", ef)
				}
			}
		})
	}
}
