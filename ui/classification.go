// ui/classification.go
package ui

import (
	"clip-tagger/state"
	"fmt"
	"path/filepath"
)

// ClassificationAction represents the action taken on the classification screen
type ClassificationAction int

const (
	ClassificationActionNone ClassificationAction = iota
	ClassificationActionPreview
	ClassificationActionSameAsLast
	ClassificationActionSelectGroup
	ClassificationActionCreateGroup
	ClassificationActionSkip
)

// ClassificationData contains the data needed to render the classification screen
type ClassificationData struct {
	CurrentFile              string
	CurrentIndex             int // 1-based index for display
	TotalFiles               int
	FilePath                 string
	HasPreviousClassification bool
	PreviousGroupName        string
	PreviousGroupID          string
}

// ClassificationUpdateResult contains the result of a classification update
type ClassificationUpdateResult struct {
	Action ClassificationAction
	Screen Screen // -1 for quit, -2 for no screen change, >= 0 for screen transition
}

// NewClassificationData creates classification data from state and file list
func NewClassificationData(appState *state.State, files []string, fileIndex int) *ClassificationData {
	if fileIndex < 0 || fileIndex >= len(files) {
		// Return empty data if index is out of bounds
		return &ClassificationData{}
	}

	currentFile := files[fileIndex]
	data := &ClassificationData{
		CurrentFile:  currentFile,
		CurrentIndex: fileIndex + 1, // Convert to 1-based index
		TotalFiles:   len(files),
		FilePath:     filepath.Join(appState.Directory, currentFile),
	}

	// Check if there's a previous classification
	if fileIndex > 0 {
		// Look for the most recent classified file before current index
		for i := fileIndex - 1; i >= 0; i-- {
			prevFile := files[i]
			if classification, ok := appState.GetClassification(prevFile); ok {
				data.HasPreviousClassification = true
				data.PreviousGroupID = classification.GroupID
				if group := appState.FindGroupByID(classification.GroupID); group != nil {
					data.PreviousGroupName = group.Name
				}
				break
			}
		}
	}

	return data
}

// ClassificationView renders the classification screen
func ClassificationView(data *ClassificationData) string {
	var output string

	// Header with progress
	output += fmt.Sprintf("=== Classification: File %d of %d ===\n\n", data.CurrentIndex, data.TotalFiles)

	// Current file info
	output += fmt.Sprintf("File: %s\n", data.CurrentFile)
	output += fmt.Sprintf("Path: %s\n\n", data.FilePath)

	// Progress indicator
	progressBar := makeProgressBar(data.CurrentIndex, data.TotalFiles, 30)
	output += fmt.Sprintf("Progress: %s\n\n", progressBar)

	// Available actions
	output += "Actions:\n"
	output += "  'p' - Preview file\n"

	// "Same as last" only if previous classification exists
	if data.HasPreviousClassification {
		output += fmt.Sprintf("  '1' - Same as last (%s)\n", data.PreviousGroupName)
	}

	output += "  '2' - Select from existing groups\n"
	output += "  '3' - Create new group\n"
	output += "  's' - Skip this file\n"
	output += "  'q' - Quit\n"

	return output
}

// ClassificationUpdate handles input for the classification screen
func ClassificationUpdate(data *ClassificationData, msg string) ClassificationUpdateResult {
	switch msg {
	case "p":
		return ClassificationUpdateResult{
			Action: ClassificationActionPreview,
			Screen: -2, // No screen change, preview will be handled separately
		}
	case "1":
		// Only allow "same as last" if there's a previous classification
		if data.HasPreviousClassification {
			return ClassificationUpdateResult{
				Action: ClassificationActionSameAsLast,
				Screen: -2, // Action handled, stay on classification screen
			}
		}
		return ClassificationUpdateResult{
			Action: ClassificationActionNone,
			Screen: -2, // Invalid action, no change
		}
	case "2":
		return ClassificationUpdateResult{
			Action: ClassificationActionSelectGroup,
			Screen: ScreenGroupSelection,
		}
	case "3":
		return ClassificationUpdateResult{
			Action: ClassificationActionCreateGroup,
			Screen: ScreenGroupInsertion,
		}
	case "s":
		return ClassificationUpdateResult{
			Action: ClassificationActionSkip,
			Screen: -2, // Action handled, will move to next file
		}
	case "q", "ctrl+c":
		return ClassificationUpdateResult{
			Action: ClassificationActionNone,
			Screen: -1, // Quit signal
		}
	default:
		return ClassificationUpdateResult{
			Action: ClassificationActionNone,
			Screen: -2, // No action
		}
	}
}

// makeProgressBar creates a simple text progress bar
func makeProgressBar(current, total, width int) string {
	if total == 0 {
		return ""
	}

	filled := (current * width) / total
	if filled > width {
		filled = width
	}

	bar := "["
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "="
		} else {
			bar += " "
		}
	}
	bar += "]"

	percentage := (current * 100) / total
	bar += fmt.Sprintf(" %d%%", percentage)

	return bar
}
