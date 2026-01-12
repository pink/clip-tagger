// ui/group_insertion.go
package ui

import (
	"clip-tagger/state"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// GroupInsertionData contains the data needed to render the group insertion screen
type GroupInsertionData struct {
	CurrentFile      string
	Mode             string // "name_entry" or "position_selection"
	GroupName        string
	ExistingGroups   []state.Group
	SelectedPosition int // Index for insertion position (0 = before first, len(groups) = after last)
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
		Mode:             "name_entry",
		GroupName:        "",
		ExistingGroups:   appState.Groups,
		SelectedPosition: 0,
	}
}

// GroupInsertionView renders the group insertion screen
func GroupInsertionView(data *GroupInsertionData) string {
	var output strings.Builder

	// Header
	output.WriteString("=== Group Insertion ===\n\n")
	output.WriteString(fmt.Sprintf("Classifying: %s\n\n", data.CurrentFile))

	if data.Mode == "name_entry" {
		// Name entry mode
		output.WriteString("Enter new group name:\n")
		output.WriteString(fmt.Sprintf("> %s\n\n", data.GroupName))

		// Show existing groups if any
		if len(data.ExistingGroups) > 0 {
			output.WriteString("Existing groups:\n")
			for _, group := range data.ExistingGroups {
				output.WriteString(fmt.Sprintf("  [%d] %s\n", group.Order, group.Name))
			}
			output.WriteString("\n")
		}

		output.WriteString("Instructions:\n")
		output.WriteString("  Type to enter group name\n")
		output.WriteString("  Enter to proceed to position selection\n")
		output.WriteString("  Backspace to delete characters\n")
		output.WriteString("  Esc to cancel\n")
		output.WriteString("  Ctrl+C to quit\n")

	} else {
		// Position selection mode
		output.WriteString(fmt.Sprintf("Position for '%s':\n\n", data.GroupName))

		// Show insertion positions
		if len(data.ExistingGroups) == 0 {
			// No existing groups
			output.WriteString("  > Insert as first group\n\n")
		} else {
			// Show all possible insertion positions
			for i := 0; i <= len(data.ExistingGroups); i++ {
				indicator := "  "
				if i == data.SelectedPosition {
					indicator = "> "
				}

				if i == 0 {
					output.WriteString(fmt.Sprintf("%sInsert at beginning (before [%d] %s)\n",
						indicator, data.ExistingGroups[0].Order, data.ExistingGroups[0].Name))
				} else if i == len(data.ExistingGroups) {
					output.WriteString(fmt.Sprintf("%sInsert at end (after [%d] %s)\n",
						indicator, data.ExistingGroups[i-1].Order, data.ExistingGroups[i-1].Name))
				} else {
					output.WriteString(fmt.Sprintf("%sInsert between [%d] %s and [%d] %s\n",
						indicator,
						data.ExistingGroups[i-1].Order, data.ExistingGroups[i-1].Name,
						data.ExistingGroups[i].Order, data.ExistingGroups[i].Name))
				}
			}
			output.WriteString("\n")
		}

		output.WriteString("Instructions:\n")
		output.WriteString("  Arrow keys to choose position\n")
		output.WriteString("  Enter to confirm\n")
		output.WriteString("  Esc to go back to name entry\n")
		output.WriteString("  Ctrl+C to quit\n")
	}

	return output.String()
}

// GroupInsertionUpdate handles input for the group insertion screen
func GroupInsertionUpdate(data *GroupInsertionData, msg string) GroupInsertionUpdateResult {
	if data.Mode == "name_entry" {
		return handleNameEntry(data, msg)
	} else {
		return handlePositionSelection(data, msg)
	}
}

// handleNameEntry handles input in name entry mode
func handleNameEntry(data *GroupInsertionData, msg string) GroupInsertionUpdateResult {
	switch msg {
	case "enter":
		// Only proceed if name is not empty
		if data.GroupName != "" {
			// If no existing groups, skip position selection and create immediately
			if len(data.ExistingGroups) == 0 {
				data.Mode = "position_selection"
				data.SelectedPosition = 0
			} else {
				// Switch to position selection mode
				data.Mode = "position_selection"
				data.SelectedPosition = 0
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

// handlePositionSelection handles input in position selection mode
func handlePositionSelection(data *GroupInsertionData, msg string) GroupInsertionUpdateResult {
	maxPosition := len(data.ExistingGroups) // Maximum position index

	switch msg {
	case "up":
		// Move selection up
		if data.SelectedPosition > 0 {
			data.SelectedPosition--
		}
		return GroupInsertionUpdateResult{Screen: -2}

	case "down":
		// Move selection down
		if data.SelectedPosition < maxPosition {
			data.SelectedPosition++
		}
		return GroupInsertionUpdateResult{Screen: -2}

	case "enter":
		// Confirm position and create group
		groupID := uuid.New().String()
		order := calculateInsertionOrder(data.ExistingGroups, data.SelectedPosition)

		return GroupInsertionUpdateResult{
			Screen:            ScreenClassification,
			InsertedGroupID:   groupID,
			InsertedGroupName: data.GroupName,
			InsertedOrder:     order,
		}

	case "esc":
		// Go back to name entry mode
		data.Mode = "name_entry"
		return GroupInsertionUpdateResult{Screen: -2}

	case "ctrl+c":
		// Quit
		return GroupInsertionUpdateResult{Screen: -1}

	default:
		// Unknown key, no action
		return GroupInsertionUpdateResult{Screen: -2}
	}
}

// calculateInsertionOrder calculates the order number for a new group at the given position
func calculateInsertionOrder(existingGroups []state.Group, position int) int {
	if len(existingGroups) == 0 {
		return 1
	}

	// Position is where to insert (0 = before first, len = after last)
	// The order should be the position + 1, which is where it will be after insertion
	return position + 1
}
