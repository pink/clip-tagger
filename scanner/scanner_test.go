// scanner/scanner_test.go
package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanner_ScanDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test video files
	files := []string{"vid1.mp4", "vid2.mov", "vid3.avi", "readme.txt"}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	scanner := NewScanner(tmpDir)
	result, err := scanner.Scan(SortByName)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	// Should find 3 video files, not the txt
	if len(result.Files) != 3 {
		t.Errorf("expected 3 files, got %d", len(result.Files))
	}

	// Check sorted by name
	expected := []string{"vid1.mp4", "vid2.mov", "vid3.avi"}
	for i, f := range result.Files {
		if filepath.Base(f.Path) != expected[i] {
			t.Errorf("index %d: expected %s, got %s", i, expected[i], filepath.Base(f.Path))
		}
	}
}

func TestIsVideoFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"video.mp4", true},
		{"video.MP4", true},
		{"video.mov", true},
		{"video.avi", true},
		{"video.mkv", true},
		{"video.webm", true},
		{"document.pdf", false},
		{"image.jpg", false},
		{"noext", false},
	}

	for _, tt := range tests {
		result := isVideoFile(tt.path)
		if result != tt.expected {
			t.Errorf("isVideoFile(%s) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}
