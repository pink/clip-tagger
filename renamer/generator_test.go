// renamer/generator_test.go
package renamer

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestGenerateFilename(t *testing.T) {
	tests := []struct {
		groupOrder int
		takeNum    int
		groupName  string
		origExt    string
		expected   string
	}{
		{1, 1, "intro", ".mp4", "[01_01] intro.mp4"},
		{2, 3, "magic trick", ".mov", "[02_03] magic trick.mov"},
		{15, 2, "outro", ".avi", "[15_02] outro.avi"},
	}

	for _, tt := range tests {
		result := GenerateFilename(tt.groupOrder, tt.takeNum, tt.groupName, tt.origExt)
		if result != tt.expected {
			t.Errorf("GenerateFilename(%d, %d, %s, %s) = %s, want %s",
				tt.groupOrder, tt.takeNum, tt.groupName, tt.origExt,
				result, tt.expected)
		}
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		num      int
		expected string
	}{
		{0, "00"},
		{1, "01"},
		{9, "09"},
		{10, "10"},
		{99, "99"},
		{100, "100"},
	}

	for _, tt := range tests {
		result := formatNumber(tt.num)
		if result != tt.expected {
			t.Errorf("formatNumber(%d) = %s, want %s", tt.num, result, tt.expected)
		}
	}
}

func TestGenerateTargetPath(t *testing.T) {
	tests := []struct {
		directory    string
		originalPath string
		groupOrder   int
		takeNumber   int
		groupName    string
		expected     string
	}{
		{"/test/dir", "video.mp4", 1, 1, "intro", "[01_01] intro.mp4"},
		{"/videos", "/old/path/clip.mov", 2, 3, "magic trick", "[02_03] magic trick.mov"},
		{".", "test.avi", 15, 2, "outro", "[15_02] outro.avi"},
		{"/output", "/source/file.MKV", 10, 5, "scene", "[10_05] scene.MKV"},
		{"relative/path", "input.mp4", 3, 7, "test scene", "[03_07] test scene.mp4"},
	}

	for _, tt := range tests {
		result := GenerateTargetPath(tt.directory, tt.originalPath, tt.groupOrder, tt.takeNumber, tt.groupName)
		// Use filepath.Join for expected path to handle platform differences
		expected := filepath.Join(tt.directory, fmt.Sprintf("[%02d_%02d] %s%s",
			tt.groupOrder, tt.takeNumber, tt.groupName, filepath.Ext(tt.originalPath)))
		if result != expected {
			t.Errorf("GenerateTargetPath(%s, %s, %d, %d, %s) = %s, want %s",
				tt.directory, tt.originalPath, tt.groupOrder, tt.takeNumber, tt.groupName,
				result, expected)
		}
	}
}
