// ui/review.go
package ui

import (
	"clip-tagger/renamer"
	"clip-tagger/state"
	"fmt"
	"path/filepath"
	"regexp"
)

// RenameItem represents a single file rename operation for display
type RenameItem struct {
	OriginalName string
	NewName      string
	IsSkipped    bool
	ChangeType   string // "new", "updated", "moved", or ""
}

// ReviewData contains the data needed to render the review screen
type ReviewData struct {
	ClassifiedCount int
	SkippedCount    int
	RenameItems     []RenameItem
	SelectedIndex   int
	ScrollOffset    int
	ViewportHeight  int // Number of items to show in viewport
}

// ReviewUpdateResult contains the result of a review update
type ReviewUpdateResult struct {
	Screen Screen // -1 for quit, -2 for no screen change, >= 0 for screen transition
}

// NewReviewData creates review data from state and file list
func NewReviewData(appState *state.State, files []string) *ReviewData {
	data := &ReviewData{
		ClassifiedCount: len(appState.Classifications),
		SkippedCount:    len(appState.Skipped),
		RenameItems:     []RenameItem{},
		SelectedIndex:   0,
		ScrollOffset:    0,
		ViewportHeight:  10, // Default viewport height
	}

	// Build rename items for classified files
	for _, classification := range appState.Classifications {
		group := appState.FindGroupByID(classification.GroupID)
		if group == nil {
			// Skip if group not found (shouldn't happen)
			continue
		}

		originalPath := filepath.Join(appState.Directory, classification.File)
		targetPath := renamer.GenerateTargetPath(
			appState.Directory,
			originalPath,
			group.Order,
			classification.TakeNumber,
			group.Name,
		)

		changeType := detectChangeType(originalPath, targetPath)

		data.RenameItems = append(data.RenameItems, RenameItem{
			OriginalName: classification.File,
			NewName:      filepath.Base(targetPath),
			IsSkipped:    false,
			ChangeType:   changeType,
		})
	}

	// Add skipped files
	for _, skipped := range appState.Skipped {
		data.RenameItems = append(data.RenameItems, RenameItem{
			OriginalName: skipped,
			NewName:      skipped, // No change for skipped files
			IsSkipped:    true,
			ChangeType:   "",
		})
	}

	return data
}

// detectChangeType determines what kind of change this rename represents
func detectChangeType(originalPath, newPath string) string {
	originalName := filepath.Base(originalPath)
	newName := filepath.Base(newPath)

	// If names are identical, no change
	if originalName == newName {
		return ""
	}

	// Pattern to match [XX_YY] format
	pattern := regexp.MustCompile(`^\[(\d+)_(\d+)\]`)

	// Check if original has the pattern
	originalMatches := pattern.FindStringSubmatch(originalName)
	newMatches := pattern.FindStringSubmatch(newName)

	// If original doesn't have pattern, it's a new file being classified
	if len(originalMatches) == 0 {
		return "new"
	}

	// If both have patterns, compare positions
	if len(originalMatches) == 3 && len(newMatches) == 3 {
		oldGroup := originalMatches[1]
		oldTake := originalMatches[2]
		newGroup := newMatches[1]
		newTake := newMatches[2]

		// If group changed, it's moved
		if oldGroup != newGroup {
			return "moved"
		}

		// If take changed within same group, it's updated
		if oldTake != newTake {
			return "updated"
		}
	}

	return ""
}

// ReviewView renders the review screen
func ReviewView(data *ReviewData) string {
	var output string

	// Header
	output += "=== Review Changes ===\n\n"

	// Summary
	classifiedText := "files"
	if data.ClassifiedCount == 1 {
		classifiedText = "file"
	}
	output += fmt.Sprintf("%d %s classified, %d skipped\n\n", data.ClassifiedCount, classifiedText, data.SkippedCount)

	// If no items to show
	if len(data.RenameItems) == 0 {
		output += "No changes to review.\n\n"
		output += "Press Esc to go back, q to quit\n"
		return output
	}

	// Calculate visible window
	startIdx := data.ScrollOffset
	endIdx := data.ScrollOffset + data.ViewportHeight
	if endIdx > len(data.RenameItems) {
		endIdx = len(data.RenameItems)
	}

	// Show scroll indicator if needed
	if data.ScrollOffset > 0 {
		output += "  ... (more items above)\n"
	}

	// Display rename items in viewport
	for i := startIdx; i < endIdx; i++ {
		item := data.RenameItems[i]

		// Selection indicator
		indicator := "  "
		if i == data.SelectedIndex {
			indicator = "> "
		}

		output += indicator

		// Show rename or skip
		if item.IsSkipped {
			output += fmt.Sprintf("%s [SKIPPED]\n", item.OriginalName)
		} else {
			output += fmt.Sprintf("%s -> %s", item.OriginalName, item.NewName)

			// Add change tag if applicable
			if item.ChangeType != "" {
				output += fmt.Sprintf(" [%s]", item.ChangeType)
			}

			output += "\n"
		}
	}

	// Show scroll indicator if needed
	if endIdx < len(data.RenameItems) {
		output += "  ... (more items below)\n"
	}

	// Instructions
	output += "\n"
	output += "Navigation:\n"
	output += "  Up/Down - Navigate list\n"
	output += "  Enter - Proceed to rename files\n"
	output += "  Esc - Return to classification (make more edits)\n"
	output += "  q - Quit\n"

	return output
}

// ReviewUpdate handles input for the review screen
func ReviewUpdate(data *ReviewData, msg string) ReviewUpdateResult {
	switch msg {
	case "up":
		// Move selection up
		if data.SelectedIndex > 0 {
			data.SelectedIndex--

			// Adjust scroll offset if selection moves above viewport
			if data.SelectedIndex < data.ScrollOffset {
				data.ScrollOffset = data.SelectedIndex
			}
		}
		return ReviewUpdateResult{Screen: -2}

	case "down":
		// Move selection down
		if data.SelectedIndex < len(data.RenameItems)-1 {
			data.SelectedIndex++

			// Adjust scroll offset if selection moves below viewport
			if data.SelectedIndex >= data.ScrollOffset+data.ViewportHeight {
				data.ScrollOffset = data.SelectedIndex - data.ViewportHeight + 1
			}
		}
		return ReviewUpdateResult{Screen: -2}

	case "enter":
		// Proceed to rename confirmation/execution
		return ReviewUpdateResult{Screen: ScreenComplete}

	case "esc":
		// Go back to classification screen
		return ReviewUpdateResult{Screen: ScreenClassification}

	case "q", "ctrl+c":
		// Quit
		return ReviewUpdateResult{Screen: -1}

	default:
		// No action for unrecognized keys
		return ReviewUpdateResult{Screen: -2}
	}
}
