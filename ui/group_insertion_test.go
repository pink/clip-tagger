// ui/group_insertion_test.go
package ui

import (
	"clip-tagger/state"
	"testing"
)

func TestNewGroupInsertionData(t *testing.T) {
	tests := []struct {
		name           string
		groups         []state.Group
		currentFile    string
		expectedMode   string
		expectedGroups int
	}{
		{
			name: "empty groups",
			groups: []state.Group{},
			currentFile: "clip01.mp4",
			expectedMode: "name_entry",
			expectedGroups: 0,
		},
		{
			name: "with existing groups",
			groups: []state.Group{
				{ID: "1", Name: "Group A", Order: 1},
				{ID: "2", Name: "Group B", Order: 2},
			},
			currentFile: "clip02.mp4",
			expectedMode: "name_entry",
			expectedGroups: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appState := &state.State{
				Directory: "/test",
				SortBy:    state.SortByName,
				Groups:    tt.groups,
			}

			data := NewGroupInsertionData(appState, tt.currentFile)

			if data.CurrentFile != tt.currentFile {
				t.Errorf("Expected CurrentFile %s, got %s", tt.currentFile, data.CurrentFile)
			}

			if data.Mode != tt.expectedMode {
				t.Errorf("Expected Mode %s, got %s", tt.expectedMode, data.Mode)
			}

			if len(data.ExistingGroups) != tt.expectedGroups {
				t.Errorf("Expected %d groups, got %d", tt.expectedGroups, len(data.ExistingGroups))
			}

			if data.GroupName != "" {
				t.Errorf("Expected empty GroupName, got %s", data.GroupName)
			}

			if data.SelectedPosition != 0 {
				t.Errorf("Expected SelectedPosition 0, got %d", data.SelectedPosition)
			}
		})
	}
}

func TestGroupInsertionView_NameEntry(t *testing.T) {
	data := &GroupInsertionData{
		CurrentFile:    "clip01.mp4",
		Mode:           "name_entry",
		GroupName:      "Test",
		ExistingGroups: []state.Group{},
	}

	view := GroupInsertionView(data)

	// Check that view contains key elements
	expectedStrings := []string{
		"Group Insertion",
		"Classifying: clip01.mp4",
		"Enter new group name:",
		"Test",
		"Enter to proceed",
		"Esc to cancel",
	}

	for _, expected := range expectedStrings {
		if !contains(view, expected) {
			t.Errorf("View missing expected string: %s\nView:\n%s", expected, view)
		}
	}
}

func TestGroupInsertionView_PositionSelection(t *testing.T) {
	data := &GroupInsertionData{
		CurrentFile:      "clip01.mp4",
		Mode:             "position_selection",
		GroupName:        "New Group",
		SelectedPosition: 1,
		ExistingGroups: []state.Group{
			{ID: "1", Name: "Group A", Order: 1},
			{ID: "2", Name: "Group B", Order: 2},
		},
	}

	view := GroupInsertionView(data)

	// Check that view contains position selection elements
	expectedStrings := []string{
		"Group Insertion",
		"Classifying: clip01.mp4",
		"Position for 'New Group'",
		"Group A",
		"Group B",
		"Arrow keys to choose",
		"Enter to confirm",
		"Esc to go back",
	}

	for _, expected := range expectedStrings {
		if !contains(view, expected) {
			t.Errorf("View missing expected string: %s\nView:\n%s", expected, view)
		}
	}

	// Check for selection indicator
	if !contains(view, ">") {
		t.Errorf("View missing selection indicator")
	}
}

func TestGroupInsertionUpdate_NameEntry(t *testing.T) {
	tests := []struct {
		name           string
		initialName    string
		input          string
		expectedName   string
		expectedMode   string
		expectedScreen Screen
	}{
		{
			name:           "add letter",
			initialName:    "Test",
			input:          "a",
			expectedName:   "Testa",
			expectedMode:   "name_entry",
			expectedScreen: -2,
		},
		{
			name:           "add space",
			initialName:    "Test",
			input:          " ",
			expectedName:   "Test ",
			expectedMode:   "name_entry",
			expectedScreen: -2,
		},
		{
			name:           "add number",
			initialName:    "Test",
			input:          "1",
			expectedName:   "Test1",
			expectedMode:   "name_entry",
			expectedScreen: -2,
		},
		{
			name:           "backspace",
			initialName:    "Test",
			input:          "backspace",
			expectedName:   "Tes",
			expectedMode:   "name_entry",
			expectedScreen: -2,
		},
		{
			name:           "backspace on empty",
			initialName:    "",
			input:          "backspace",
			expectedName:   "",
			expectedMode:   "name_entry",
			expectedScreen: -2,
		},
		{
			name:           "enter without groups proceeds to position selection",
			initialName:    "TestGroup",
			input:          "enter",
			expectedName:   "TestGroup",
			expectedMode:   "position_selection",
			expectedScreen: -2,
		},
		{
			name:           "enter with empty name does nothing",
			initialName:    "",
			input:          "enter",
			expectedName:   "",
			expectedMode:   "name_entry",
			expectedScreen: -2,
		},
		{
			name:           "escape cancels",
			initialName:    "Test",
			input:          "esc",
			expectedName:   "Test",
			expectedMode:   "name_entry",
			expectedScreen: ScreenClassification,
		},
		{
			name:           "ctrl+c quits",
			initialName:    "Test",
			input:          "ctrl+c",
			expectedName:   "Test",
			expectedMode:   "name_entry",
			expectedScreen: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &GroupInsertionData{
				CurrentFile:    "clip01.mp4",
				Mode:           "name_entry",
				GroupName:      tt.initialName,
				ExistingGroups: []state.Group{},
			}

			result := GroupInsertionUpdate(data, tt.input)

			if data.GroupName != tt.expectedName {
				t.Errorf("Expected GroupName %s, got %s", tt.expectedName, data.GroupName)
			}

			if data.Mode != tt.expectedMode {
				t.Errorf("Expected Mode %s, got %s", tt.expectedMode, data.Mode)
			}

			if result.Screen != tt.expectedScreen {
				t.Errorf("Expected Screen %d, got %d", tt.expectedScreen, result.Screen)
			}
		})
	}
}

func TestGroupInsertionUpdate_PositionSelection(t *testing.T) {
	tests := []struct {
		name             string
		initialPosition  int
		existingGroups   []state.Group
		input            string
		expectedPosition int
		expectedMode     string
		expectedScreen   Screen
		expectGroupID    bool
	}{
		{
			name:            "arrow down moves selection",
			initialPosition: 0,
			existingGroups: []state.Group{
				{ID: "1", Name: "Group A", Order: 1},
				{ID: "2", Name: "Group B", Order: 2},
			},
			input:            "down",
			expectedPosition: 1,
			expectedMode:     "position_selection",
			expectedScreen:   -2,
			expectGroupID:    false,
		},
		{
			name:            "arrow up moves selection",
			initialPosition: 2,
			existingGroups: []state.Group{
				{ID: "1", Name: "Group A", Order: 1},
				{ID: "2", Name: "Group B", Order: 2},
			},
			input:            "up",
			expectedPosition: 1,
			expectedMode:     "position_selection",
			expectedScreen:   -2,
			expectGroupID:    false,
		},
		{
			name:            "arrow up at top does nothing",
			initialPosition: 0,
			existingGroups: []state.Group{
				{ID: "1", Name: "Group A", Order: 1},
			},
			input:            "up",
			expectedPosition: 0,
			expectedMode:     "position_selection",
			expectedScreen:   -2,
			expectGroupID:    false,
		},
		{
			name:            "arrow down at bottom does nothing",
			initialPosition: 2,
			existingGroups: []state.Group{
				{ID: "1", Name: "Group A", Order: 1},
				{ID: "2", Name: "Group B", Order: 2},
			},
			input:            "down",
			expectedPosition: 2,
			expectedMode:     "position_selection",
			expectedScreen:   -2,
			expectGroupID:    false,
		},
		{
			name:            "enter confirms position",
			initialPosition: 1,
			existingGroups: []state.Group{
				{ID: "1", Name: "Group A", Order: 1},
				{ID: "2", Name: "Group B", Order: 2},
			},
			input:            "enter",
			expectedPosition: 1,
			expectedMode:     "position_selection",
			expectedScreen:   ScreenClassification,
			expectGroupID:    true,
		},
		{
			name:            "escape goes back to name entry",
			initialPosition: 1,
			existingGroups: []state.Group{
				{ID: "1", Name: "Group A", Order: 1},
			},
			input:            "esc",
			expectedPosition: 1,
			expectedMode:     "name_entry",
			expectedScreen:   -2,
			expectGroupID:    false,
		},
		{
			name:            "ctrl+c quits",
			initialPosition: 0,
			existingGroups:  []state.Group{},
			input:           "ctrl+c",
			expectedPosition: 0,
			expectedMode:     "position_selection",
			expectedScreen:   -1,
			expectGroupID:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &GroupInsertionData{
				CurrentFile:      "clip01.mp4",
				Mode:             "position_selection",
				GroupName:        "New Group",
				SelectedPosition: tt.initialPosition,
				ExistingGroups:   tt.existingGroups,
			}

			result := GroupInsertionUpdate(data, tt.input)

			if data.SelectedPosition != tt.expectedPosition {
				t.Errorf("Expected SelectedPosition %d, got %d", tt.expectedPosition, data.SelectedPosition)
			}

			if data.Mode != tt.expectedMode {
				t.Errorf("Expected Mode %s, got %s", tt.expectedMode, data.Mode)
			}

			if result.Screen != tt.expectedScreen {
				t.Errorf("Expected Screen %d, got %d", tt.expectedScreen, result.Screen)
			}

			if tt.expectGroupID {
				if result.InsertedGroupID == "" {
					t.Errorf("Expected InsertedGroupID to be set")
				}
				if result.InsertedGroupName != "New Group" {
					t.Errorf("Expected InsertedGroupName 'New Group', got %s", result.InsertedGroupName)
				}
			} else {
				if result.InsertedGroupID != "" {
					t.Errorf("Expected InsertedGroupID to be empty, got %s", result.InsertedGroupID)
				}
			}
		})
	}
}

func TestGroupInsertionUpdate_PositionSelection_EmptyGroups(t *testing.T) {
	// When there are no existing groups, pressing Enter should immediately create group
	data := &GroupInsertionData{
		CurrentFile:      "clip01.mp4",
		Mode:             "name_entry",
		GroupName:        "First Group",
		ExistingGroups:   []state.Group{},
		SelectedPosition: 0,
	}

	// First, press enter to move to position selection
	result := GroupInsertionUpdate(data, "enter")
	if result.Screen != -2 {
		t.Fatalf("Expected to stay in screen (mode change), got screen %d", result.Screen)
	}
	if data.Mode != "position_selection" {
		t.Fatalf("Expected mode 'position_selection', got %s", data.Mode)
	}

	// Now press enter again in position selection mode (with no groups)
	result = GroupInsertionUpdate(data, "enter")
	if result.Screen != ScreenClassification {
		t.Errorf("Expected to return to classification screen, got screen %d", result.Screen)
	}
	if result.InsertedGroupID == "" {
		t.Errorf("Expected InsertedGroupID to be set")
	}
	if result.InsertedGroupName != "First Group" {
		t.Errorf("Expected InsertedGroupName 'First Group', got %s", result.InsertedGroupName)
	}
	if result.InsertedOrder != 1 {
		t.Errorf("Expected InsertedOrder 1, got %d", result.InsertedOrder)
	}
}

func TestCalculateInsertionOrder(t *testing.T) {
	tests := []struct {
		name             string
		existingGroups   []state.Group
		selectedPosition int
		expectedOrder    int
	}{
		{
			name:             "empty groups list",
			existingGroups:   []state.Group{},
			selectedPosition: 0,
			expectedOrder:    1,
		},
		{
			name: "insert at beginning",
			existingGroups: []state.Group{
				{ID: "1", Name: "Group A", Order: 1},
				{ID: "2", Name: "Group B", Order: 2},
			},
			selectedPosition: 0,
			expectedOrder:    1,
		},
		{
			name: "insert in middle",
			existingGroups: []state.Group{
				{ID: "1", Name: "Group A", Order: 1},
				{ID: "2", Name: "Group B", Order: 2},
			},
			selectedPosition: 1,
			expectedOrder:    2,
		},
		{
			name: "insert at end",
			existingGroups: []state.Group{
				{ID: "1", Name: "Group A", Order: 1},
				{ID: "2", Name: "Group B", Order: 2},
			},
			selectedPosition: 2,
			expectedOrder:    3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := calculateInsertionOrder(tt.existingGroups, tt.selectedPosition)
			if order != tt.expectedOrder {
				t.Errorf("Expected order %d, got %d", tt.expectedOrder, order)
			}
		})
	}
}
