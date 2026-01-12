// renamer/generator.go
package renamer

import (
	"fmt"
	"os"
	"path/filepath"
)

// Rename represents a file rename operation
type Rename struct {
	OriginalPath string
	TargetPath   string
	ChangeType   string // "new", "updated", "moved", or ""
}

// GenerateFilename creates a filename in format [XX_YY] name.ext
func GenerateFilename(groupOrder, takeNumber int, groupName, extension string) string {
	groupNum := formatNumber(groupOrder)
	takeNum := formatNumber(takeNumber)
	return fmt.Sprintf("[%s_%s] %s%s", groupNum, takeNum, groupName, extension)
}

// formatNumber formats a number with leading zero (01, 02, ..., 10, 11, ...)
func formatNumber(n int) string {
	if n < 10 {
		return fmt.Sprintf("0%d", n)
	}
	return fmt.Sprintf("%d", n)
}

// GenerateTargetPath generates the full target path for a file
func GenerateTargetPath(directory, originalPath string, groupOrder, takeNumber int, groupName string) string {
	ext := filepath.Ext(originalPath)
	newName := GenerateFilename(groupOrder, takeNumber, groupName, ext)
	return filepath.Join(directory, newName)
}

// DetectConflicts checks if any target paths already exist
func DetectConflicts(renames []Rename) []Rename {
	var conflicts []Rename
	for _, r := range renames {
		// Skip if target is same as source (no actual rename)
		if r.OriginalPath == r.TargetPath {
			continue
		}

		if _, err := os.Stat(r.TargetPath); err == nil {
			conflicts = append(conflicts, r)
		}
	}
	return conflicts
}
