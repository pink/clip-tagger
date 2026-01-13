// ui/group_insertion.go
package ui

import (
	"clip-tagger/state"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// GroupInsertionMode represents the mode of group insertion
type GroupInsertionMode int

const (
	ModeNameEntry GroupInsertionMode = iota
	ModeInsertionChoice
	ModeGroupSelection
)

// GroupInsertionData contains the data needed to render the group insertion screen
type GroupInsertionData struct {
	CurrentFile      string
	Mode             GroupInsertionMode
	GroupName        string
	ExistingGroups   []state.Group
	FilteredGroups   []state.Group // Filtered list based on FilterQuery
	FilterQuery      string        // Query text for filtering groups
	SelectedPosition int           // Index for cursor position in choice/selection modes
	ScrollOffset     int           // Track scroll position for group lists
	ViewportHeight   int           // Number of items to show (default: 10)
}

// GroupInsertionUpdateResult contains the result of a group insertion update
type GroupInsertionUpdateResult struct {
	Screen            Screen // -1 for quit, -2 for no screen change, >= 0 for screen transition
	InsertedGroupID   string
	InsertedGroupName string
	InsertedOrder     int
}

// NewGroupInsertionData creates group insertion data from state and current file
func NewGroupInsertionData(appState *state.State, currentFile string) *GroupInsertionData {
	return &GroupInsertionData{
		CurrentFile:      currentFile,
		Mode:             ModeNameEntry,
		GroupName:        "",
		ExistingGroups:   appState.Groups,
		FilteredGroups:   appState.Groups, // Initially show all groups
		FilterQuery:      "",
		SelectedPosition: 0,
		ScrollOffset:     0,
		ViewportHeight:   10,
	}
}

// GroupInsertionView renders the group insertion screen
func GroupInsertionView(data *GroupInsertionData) string {
	var output strings.Builder

	// Header
	output.WriteString(RenderHeader("=== Group Insertion ===") + "\n\n")
	output.WriteString(fmt.Sprintf("%s %s\n\n", RenderMuted("Classifying:"), RenderSubheader(data.CurrentFile)))

	switch data.Mode {
	case ModeNameEntry:
		// Name entry mode
		output.WriteString(RenderHighlight("Enter new group name:") + "\n")
		output.WriteString(fmt.Sprintf("%s %s\n\n", RenderCursor(">"), RenderSubheader(data.GroupName)))

		// Show existing groups if any (with viewport to keep input on screen)
		if len(data.ExistingGroups) > 0 {
			output.WriteString(RenderMuted("Existing groups:") + "\n")

			// Show maximum 5 groups to keep input on screen
			maxShow := 5
			if len(data.ExistingGroups) > maxShow {
				for i := 0; i < maxShow; i++ {
					group := data.ExistingGroups[i]
					output.WriteString(fmt.Sprintf("  %s %s\n",
						RenderMuted(fmt.Sprintf("[%d]", group.Order)),
						group.Name))
				}
				remaining := len(data.ExistingGroups) - maxShow
				output.WriteString(RenderMuted(fmt.Sprintf("  ... and %d more groups\n", remaining)))
			} else {
				for _, group := range data.ExistingGroups {
					output.WriteString(fmt.Sprintf("  %s %s\n",
						RenderMuted(fmt.Sprintf("[%d]", group.Order)),
						group.Name))
				}
			}
			output.WriteString("\n")
		}

		output.WriteString(RenderMuted("Instructions:") + "\n")
		output.WriteString(RenderKeyHint("  Type to enter group name") + "\n")
		output.WriteString(RenderKeyHint("  Enter to proceed") + "\n")
		output.WriteString(RenderKeyHint("  Backspace to delete characters") + "\n")
		output.WriteString(RenderKeyHint("  Esc to cancel") + "\n")
		output.WriteString(RenderKeyHint("  Ctrl+C to quit") + "\n")

	case ModeInsertionChoice:
		// Insertion choice mode
		output.WriteString(RenderHighlight(fmt.Sprintf("Where should \"%s\" be added?", data.GroupName)) + "\n\n")

		// Show options
		options := []string{
			"1. Add to end",
			"2. Insert after existing group",
		}

		for i, option := range options {
			if i == data.SelectedPosition {
				output.WriteString(fmt.Sprintf("%s %s\n", RenderCursor(">"), RenderHighlight(option)))
			} else {
				output.WriteString(fmt.Sprintf("  %s\n", option))
			}
		}

		output.WriteString("\n")
		output.WriteString(RenderMuted("Instructions:") + "\n")
		output.WriteString(RenderKeyHint("  1-2 or Up/Down to select") + "\n")
		output.WriteString(RenderKeyHint("  Enter to confirm") + "\n")
		output.WriteString(RenderKeyHint("  Esc to go back") + "\n")
		output.WriteString(RenderKeyHint("  Ctrl+C to quit") + "\n")

	case ModeGroupSelection:
		// Group selection mode
		output.WriteString(RenderHighlight("Select group to insert after:") + "\n\n")

		// Filter input
		output.WriteString(fmt.Sprintf("%s %s\n\n", RenderMuted("Filter:"), RenderSubheader(data.FilterQuery)))

		// Show filtered groups with viewport
		if len(data.FilteredGroups) == 0 {
			output.WriteString(RenderWarning("No groups match your filter.") + "\n")
		} else {
			// Calculate visible window
			startIdx := data.ScrollOffset
			endIdx := data.ScrollOffset + data.ViewportHeight
			if endIdx > len(data.FilteredGroups) {
				endIdx = len(data.FilteredGroups)
			}

			// Show scroll indicator if needed
			if data.ScrollOffset > 0 {
				output.WriteString(RenderMuted("  ... (more items above)") + "\n")
			}

			// Display groups in viewport
			for i := startIdx; i < endIdx; i++ {
				group := data.FilteredGroups[i]
				if i == data.SelectedPosition {
					output.WriteString(fmt.Sprintf("%s %s %s\n",
						RenderCursor(">"),
						RenderMuted(fmt.Sprintf("[%d]", group.Order)),
						RenderHighlight(group.Name)))
				} else {
					output.WriteString(fmt.Sprintf("  %s %s\n",
						RenderMuted(fmt.Sprintf("[%d]", group.Order)),
						group.Name))
				}
			}

			// Show scroll indicator if needed
			if endIdx < len(data.FilteredGroups) {
				output.WriteString(RenderMuted("  ... (more items below)") + "\n")
			}
		}

		output.WriteString("\n")
		output.WriteString(RenderMuted("Instructions:") + "\n")
		output.WriteString(RenderKeyHint("  Type to filter groups") + "\n")
		output.WriteString(RenderKeyHint("  Up/Down to navigate") + "\n")
		output.WriteString(RenderKeyHint("  Enter to select") + "\n")
		output.WriteString(RenderKeyHint("  Backspace to delete filter character") + "\n")
		output.WriteString(RenderKeyHint("  Esc to go back") + "\n")
		output.WriteString(RenderKeyHint("  Ctrl+C to quit") + "\n")
	}

	return output.String()
}

// GroupInsertionUpdate handles input for the group insertion screen
func GroupInsertionUpdate(data *GroupInsertionData, msg string) GroupInsertionUpdateResult {
	switch data.Mode {
	case ModeNameEntry:
		return handleNameEntry(data, msg)
	case ModeInsertionChoice:
		return handleInsertionChoice(data, msg)
	case ModeGroupSelection:
		return handleGroupSelection(data, msg)
	default:
		return GroupInsertionUpdateResult{Screen: -2}
	}
}

// handleNameEntry handles input in name entry mode
func handleNameEntry(data *GroupInsertionData, msg string) GroupInsertionUpdateResult {
	switch msg {
	case "enter":
		// Only proceed if name is not empty
		if data.GroupName != "" {
			// If no existing groups, create immediately at order 1
			if len(data.ExistingGroups) == 0 {
				groupID := uuid.New().String()
				return GroupInsertionUpdateResult{
					Screen:            ScreenClassification,
					InsertedGroupID:   groupID,
					InsertedGroupName: data.GroupName,
					InsertedOrder:     1,
				}
			} else {
				// Switch to insertion choice mode
				data.Mode = ModeInsertionChoice
				data.SelectedPosition = 0 // Default to first option (add to end)
			}
		}
		return GroupInsertionUpdateResult{Screen: -2}

	case "esc":
		// Cancel and return to classification screen
		return GroupInsertionUpdateResult{Screen: ScreenClassification}

	case "ctrl+c":
		// Quit
		return GroupInsertionUpdateResult{Screen: -1}

	case "backspace":
		// Remove last character from group name
		if len(data.GroupName) > 0 {
			data.GroupName = data.GroupName[:len(data.GroupName)-1]
		}
		return GroupInsertionUpdateResult{Screen: -2}

	default:
		// Check if it's a printable character for name entry
		if len(msg) == 1 || msg == " " {
			// Add to group name
			data.GroupName += msg
		}
		return GroupInsertionUpdateResult{Screen: -2}
	}
}

// handleInsertionChoice handles input in insertion choice mode
func handleInsertionChoice(data *GroupInsertionData, msg string) GroupInsertionUpdateResult {
	switch msg {
	case "up":
		// Move selection up
		if data.SelectedPosition > 0 {
			data.SelectedPosition--
		}
		return GroupInsertionUpdateResult{Screen: -2}

	case "down":
		// Move selection down
		if data.SelectedPosition < 1 {
			data.SelectedPosition++
		}
		return GroupInsertionUpdateResult{Screen: -2}

	case "1":
		// Option 1: Add to end
		groupID := uuid.New().String()
		order := len(data.ExistingGroups) + 1

		return GroupInsertionUpdateResult{
			Screen:            ScreenClassification,
			InsertedGroupID:   groupID,
			InsertedGroupName: data.GroupName,
			InsertedOrder:     order,
		}

	case "2":
		// Option 2: Insert after existing group
		data.Mode = ModeGroupSelection
		data.SelectedPosition = 0
		data.ScrollOffset = 0
		data.FilterQuery = ""
		data.FilteredGroups = data.ExistingGroups
		return GroupInsertionUpdateResult{Screen: -2}

	case "enter":
		// Confirm current selection
		if data.SelectedPosition == 0 {
			// Option 1: Add to end
			groupID := uuid.New().String()
			order := len(data.ExistingGroups) + 1

			return GroupInsertionUpdateResult{
				Screen:            ScreenClassification,
				InsertedGroupID:   groupID,
				InsertedGroupName: data.GroupName,
				InsertedOrder:     order,
			}
		} else {
			// Option 2: Insert after existing group
			data.Mode = ModeGroupSelection
			data.SelectedPosition = 0
			data.ScrollOffset = 0
			data.FilterQuery = ""
			data.FilteredGroups = data.ExistingGroups
			return GroupInsertionUpdateResult{Screen: -2}
		}

	case "esc":
		// Go back to name entry mode
		data.Mode = ModeNameEntry
		return GroupInsertionUpdateResult{Screen: -2}

	case "ctrl+c":
		// Quit
		return GroupInsertionUpdateResult{Screen: -1}

	default:
		// Unknown key, no action
		return GroupInsertionUpdateResult{Screen: -2}
	}
}

// handleGroupSelection handles input in group selection mode
func handleGroupSelection(data *GroupInsertionData, msg string) GroupInsertionUpdateResult {
	switch msg {
	case "up":
		// Move selection up
		if data.SelectedPosition > 0 {
			data.SelectedPosition--

			// Adjust scroll offset if selection moves above viewport
			if data.SelectedPosition < data.ScrollOffset {
				data.ScrollOffset = data.SelectedPosition
			}
		}
		return GroupInsertionUpdateResult{Screen: -2}

	case "down":
		// Move selection down
		if len(data.FilteredGroups) > 0 && data.SelectedPosition < len(data.FilteredGroups)-1 {
			data.SelectedPosition++

			// Adjust scroll offset if selection moves below viewport
			if data.SelectedPosition >= data.ScrollOffset+data.ViewportHeight {
				data.ScrollOffset = data.SelectedPosition - data.ViewportHeight + 1
			}
		}
		return GroupInsertionUpdateResult{Screen: -2}

	case "enter":
		// Select current group and insert after it
		if len(data.FilteredGroups) > 0 && data.SelectedPosition < len(data.FilteredGroups) {
			selectedGroup := data.FilteredGroups[data.SelectedPosition]
			groupID := uuid.New().String()
			// Insert after selected group: new order = selected group order + 1
			order := selectedGroup.Order + 1

			return GroupInsertionUpdateResult{
				Screen:            ScreenClassification,
				InsertedGroupID:   groupID,
				InsertedGroupName: data.GroupName,
				InsertedOrder:     order,
			}
		}
		return GroupInsertionUpdateResult{Screen: -2}

	case "backspace":
		// Remove last character from filter
		if len(data.FilterQuery) > 0 {
			data.FilterQuery = data.FilterQuery[:len(data.FilterQuery)-1]
			data.FilteredGroups = filterGroupsByName(data.ExistingGroups, data.FilterQuery)
			// Reset selection AND scroll to top
			data.SelectedPosition = 0
			data.ScrollOffset = 0
		}
		return GroupInsertionUpdateResult{Screen: -2}

	case "esc":
		// Go back to insertion choice mode
		data.Mode = ModeInsertionChoice
		data.SelectedPosition = 0
		return GroupInsertionUpdateResult{Screen: -2}

	case "ctrl+c":
		// Quit
		return GroupInsertionUpdateResult{Screen: -1}

	default:
		// Check if it's a printable character for filtering
		if len(msg) == 1 || msg == " " {
			// Add to filter query
			data.FilterQuery += msg
			data.FilteredGroups = filterGroupsByName(data.ExistingGroups, data.FilterQuery)
			// Reset selection AND scroll to top
			data.SelectedPosition = 0
			data.ScrollOffset = 0
			return GroupInsertionUpdateResult{Screen: -2}
		}

		// Unknown key, no action
		return GroupInsertionUpdateResult{Screen: -2}
	}
}

// filterGroupsByName filters groups by case-insensitive substring matching
func filterGroupsByName(groups []state.Group, query string) []state.Group {
	if query == "" {
		return groups
	}

	lowerQuery := strings.ToLower(query)
	filtered := make([]state.Group, 0)

	for _, group := range groups {
		if strings.Contains(strings.ToLower(group.Name), lowerQuery) {
			filtered = append(filtered, group)
		}
	}

	return filtered
}
