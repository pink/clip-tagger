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
	output += fmt.Sprintf("=== Group Selection ===\n\n")
	output += fmt.Sprintf("Classifying: %s\n\n", data.CurrentFile)

	// Filter input
	output += fmt.Sprintf("Filter: %s\n\n", data.FilterText)

	// List of groups
	if len(data.FilteredGroups) == 0 {
		output += "No groups available.\n"
		if data.FilterText != "" {
			output += fmt.Sprintf("No groups match '%s'.\n", data.FilterText)
		}
	} else {
		output += "Groups:\n"
		for i, group := range data.FilteredGroups {
			// Show selection indicator
			indicator := "  "
			if i == data.SelectedIndex {
				indicator = "> "
			}
			output += fmt.Sprintf("%s[%d] %s\n", indicator, group.Order, group.Name)
		}
	}

	output += "\n"

	// Instructions
	output += "Instructions:\n"
	output += "  Type to filter groups (case-insensitive)\n"
	output += "  Use arrow keys to navigate\n"
	output += "  Enter to select\n"
	output += "  Backspace to delete filter character\n"
	output += "  Esc to cancel\n"
	output += "  Ctrl+C to quit\n"

	return output
}

// GroupSelectionUpdate handles input for the group selection screen
func GroupSelectionUpdate(data *GroupSelectionData, msg string) GroupSelectionUpdateResult {
	switch msg {
	case "up":
		// Move selection up
		if data.SelectedIndex > 0 {
			data.SelectedIndex--
		}
		return GroupSelectionUpdateResult{Screen: -2}

	case "down":
		// Move selection down
		if len(data.FilteredGroups) > 0 && data.SelectedIndex < len(data.FilteredGroups)-1 {
			data.SelectedIndex++
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
			// Reset selection to top
			data.SelectedIndex = 0
		}
		return GroupSelectionUpdateResult{Screen: -2}

	default:
		// Check if it's a printable character for filtering
		if len(msg) == 1 || msg == " " {
			// Add to filter text
			data.FilterText += msg
			data.FilteredGroups = filterGroups(data.AllGroups, data.FilterText)
			// Reset selection to top
			data.SelectedIndex = 0
			return GroupSelectionUpdateResult{Screen: -2}
		}

		// Unknown key, no action
		return GroupSelectionUpdateResult{Screen: -2}
	}
}
