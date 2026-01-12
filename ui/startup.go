// ui/startup.go
package ui

import (
	"clip-tagger/state"
	"fmt"
)

// StartupData contains the data needed to render the startup screen
type StartupData struct {
	IsResume          bool
	ClassifiedCount   int
	RemainingCount    int
	TotalFiles        int
	NewFilesCount     int
	MissingFilesCount int
	SortBy            state.SortBy
}

// NewStartupData creates startup data from state and scanned files
func NewStartupData(appState *state.State, scannedFiles []string, mergeResult *state.MergeResult) *StartupData {
	data := &StartupData{
		TotalFiles: len(scannedFiles),
		SortBy:     appState.SortBy,
	}

	// Determine if this is a resume session
	data.IsResume = len(appState.Classifications) > 0

	// Calculate classified count from state
	data.ClassifiedCount = len(appState.Classifications)

	// If we have a merge result, use it for new/missing file counts
	if mergeResult != nil {
		data.NewFilesCount = len(mergeResult.NewFiles)
		data.MissingFilesCount = len(mergeResult.MissingFiles)
		data.RemainingCount = len(mergeResult.NewFiles)
	} else {
		// No merge result - calculate remaining count
		if data.IsResume {
			// Count unclassified files
			classified := make(map[string]bool)
			for _, c := range appState.Classifications {
				classified[c.File] = true
			}
			remainingCount := 0
			for _, f := range scannedFiles {
				if !classified[f] {
					remainingCount++
				}
			}
			data.RemainingCount = remainingCount
		} else {
			// New session - all files are remaining
			data.RemainingCount = len(scannedFiles)
		}
		data.NewFilesCount = 0
		data.MissingFilesCount = 0
	}

	return data
}

// StartupView renders the startup screen
func StartupView(data *StartupData) string {
	var output string

	// Header
	output += "=== clip-tagger ===\n\n"

	// Session status
	if data.IsResume {
		output += fmt.Sprintf("Found existing session (%d classified, %d remaining)\n", data.ClassifiedCount, data.RemainingCount)

		// New files information
		if data.NewFilesCount > 0 {
			output += fmt.Sprintf("  %d new files detected\n", data.NewFilesCount)
		}

		// Missing files warning
		if data.MissingFilesCount > 0 {
			output += fmt.Sprintf("  WARNING: %d missing file", data.MissingFilesCount)
			if data.MissingFilesCount > 1 {
				output += "s"
			}
			output += " (previously classified but not found)\n"
		}
	} else {
		output += "New session\n"
	}

	// File count
	output += fmt.Sprintf("\nTotal: %d files\n", data.TotalFiles)

	// Sorting information
	output += fmt.Sprintf("Sorted by: %s\n", data.SortBy)

	// Instructions
	output += "\n"
	if data.IsResume {
		output += "Press Enter to continue classification\n"
	} else {
		output += "Press Enter to start classification\n"
	}
	output += "Press 'q' or Ctrl+C to quit\n"

	return output
}

// StartupUpdate handles input for the startup screen
// Returns the screen to transition to, or -1 for quit, or -2 for no action
func StartupUpdate(data *StartupData, msg string) Screen {
	switch msg {
	case "enter":
		return ScreenClassification
	case "q", "ctrl+c":
		return -1 // Signal to quit
	default:
		return -2 // No action
	}
}
