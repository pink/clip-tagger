// renamer/generator_test.go
package renamer

import (
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
