// ui/group_selection.go
package ui

import (
	"clip-tagger/state"
	"fmt"
	"strings"
)

// GroupSelectionData contains the data needed to render the group selection screen
type GroupSelectionData struct {
	CurrentFile    string
	AllGroups      []state.Group
	FilteredGroups []state.Group
	FilterText     string
	SelectedIndex  int
	ScrollOffset   int // Track scroll position
	ViewportHeight int // Number of items to show (default: 10)
}

// GroupSelectionUpdateResult contains the result of a group selection update
type GroupSelectionUpdateResult struct {
	Screen            Screen // -1 for quit, -2 for no screen change, >= 0 for screen transition
	SelectedGroupID   string
	SelectedGroupName string
}

// NewGroupSelectionData creates group selection data from state and current file
func NewGroupSelectionData(appState *state.State, currentFile string) *GroupSelectionData {
	return &GroupSelectionData{
		CurrentFile:    currentFile,
		AllGroups:      appState.Groups,
		FilteredGroups: appState.Groups,
		FilterText:     "",
		SelectedIndex:  0,
		ScrollOffset:   0,
		ViewportHeight: 10,
	}
}

// filterGroups filters groups by case-insensitive substring matching
func filterGroups(groups []state.Group, filterText string) []state.Group {
	if filterText == "" {
		return groups
	}

	lowerFilter := strings.ToLower(filterText)
	filtered := make([]state.Group, 0)

	for _, group := range groups {
		if strings.Contains(strings.ToLower(group.Name), lowerFilter) {
			filtered = append(filtered, group)
		}
	}

	return filtered
}

// GroupSelectionView renders the group selection screen
func GroupSelectionView(data *GroupSelectionData) string {
	var output string

	// Header with current file context
	output += RenderHeader("=== Group Selection ===") + "\n\n"
	output += fmt.Sprintf("%s %s\n\n", RenderMuted("Classifying:"), RenderSubheader(data.CurrentFile))

	// Filter input
	output += fmt.Sprintf("%s %s\n\n", RenderMuted("Filter:"), RenderHighlight(data.FilterText))

	// List of groups
	if len(data.FilteredGroups) == 0 {
		output += RenderWarning("No groups available.") + "\n"
		if data.FilterText != "" {
			output += RenderWarning(fmt.Sprintf("No groups match '%s'.", data.FilterText)) + "\n"
		}
	} else {
		output += RenderHighlight("Groups:") + "\n"

		// Calculate visible window
		startIdx := data.ScrollOffset
		endIdx := data.ScrollOffset + data.ViewportHeight
		if endIdx > len(data.FilteredGroups) {
			endIdx = len(data.FilteredGroups)
		}

		// Show scroll indicator if needed
		if data.ScrollOffset > 0 {
			output += RenderMuted("  ... (more items above)") + "\n"
		}

		// Display groups in viewport
		for i := startIdx; i < endIdx; i++ {
			group := data.FilteredGroups[i]
			// Show selection indicator
			if i == data.SelectedIndex {
				output += fmt.Sprintf("%s %s %s\n",
					RenderCursor(">"),
					RenderMuted(fmt.Sprintf("[%d]", group.Order)),
					RenderHighlight(group.Name))
			} else {
				output += fmt.Sprintf("  %s %s\n",
					RenderMuted(fmt.Sprintf("[%d]", group.Order)),
					group.Name)
			}
		}

		// Show scroll indicator if needed
		if endIdx < len(data.FilteredGroups) {
			output += RenderMuted("  ... (more items below)") + "\n"
		}
	}

	output += "\n"

	// Instructions
	output += RenderMuted("Instructions:") + "\n"
	output += RenderKeyHint("  Type to filter groups (case-insensitive)") + "\n"
	output += RenderKeyHint("  Use arrow keys to navigate") + "\n"
	output += RenderKeyHint("  Enter to select") + "\n"
	output += RenderKeyHint("  Backspace to delete filter character") + "\n"
	output += RenderKeyHint("  Esc to cancel") + "\n"
	output += RenderKeyHint("  Ctrl+C to quit") + "\n"

	return output
}

// GroupSelectionUpdate handles input for the group selection screen
func GroupSelectionUpdate(data *GroupSelectionData, msg string) GroupSelectionUpdateResult {
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
		return GroupSelectionUpdateResult{Screen: -2}

	case "down":
		// Move selection down
		if len(data.FilteredGroups) > 0 && data.SelectedIndex < len(data.FilteredGroups)-1 {
			data.SelectedIndex++

			// Adjust scroll offset if selection moves below viewport
			if data.SelectedIndex >= data.ScrollOffset+data.ViewportHeight {
				data.ScrollOffset = data.SelectedIndex - data.ViewportHeight + 1
			}
		}
		return GroupSelectionUpdateResult{Screen: -2}

	case "enter":
		// Select current group
		if len(data.FilteredGroups) > 0 && data.SelectedIndex < len(data.FilteredGroups) {
			selectedGroup := data.FilteredGroups[data.SelectedIndex]
			return GroupSelectionUpdateResult{
				Screen:            ScreenClassification,
				SelectedGroupID:   selectedGroup.ID,
				SelectedGroupName: selectedGroup.Name,
			}
		}
		return GroupSelectionUpdateResult{Screen: -2}

	case "esc":
		// Cancel and return to classification screen
		return GroupSelectionUpdateResult{
			Screen:            ScreenClassification,
			SelectedGroupID:   "",
			SelectedGroupName: "",
		}

	case "ctrl+c":
		// Quit
		return GroupSelectionUpdateResult{Screen: -1}

	case "backspace":
		// Remove last character from filter
		if len(data.FilterText) > 0 {
			data.FilterText = data.FilterText[:len(data.FilterText)-1]
			data.FilteredGroups = filterGroups(data.AllGroups, data.FilterText)
			// Reset selection AND scroll to top
			data.SelectedIndex = 0
			data.ScrollOffset = 0
		}
		return GroupSelectionUpdateResult{Screen: -2}

	default:
		// Check if it's a printable character for filtering
		if len(msg) == 1 || msg == " " {
			// Add to filter text
			data.FilterText += msg
			data.FilteredGroups = filterGroups(data.AllGroups, data.FilterText)
			// Reset selection AND scroll to top
			data.SelectedIndex = 0
			data.ScrollOffset = 0
			return GroupSelectionUpdateResult{Screen: -2}
		}

		// Unknown key, no action
		return GroupSelectionUpdateResult{Screen: -2}
	}
}
