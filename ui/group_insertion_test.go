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
		expectedMode   GroupInsertionMode
		expectedGroups int
	}{
		{
			name: "empty groups",
			groups: []state.Group{},
			currentFile: "clip01.mp4",
			expectedMode: ModeNameEntry,
			expectedGroups: 0,
		},
		{
			name: "with existing groups",
			groups: []state.Group{
				{ID: "1", Name: "Group A", Order: 1},
				{ID: "2", Name: "Group B", Order: 2},
			},
			currentFile: "clip02.mp4",
			expectedMode: ModeNameEntry,
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
				t.Errorf("Expected Mode %v, got %v", tt.expectedMode, data.Mode)
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
		Mode:           ModeNameEntry,
		GroupName:      "Test",
		ExistingGroups: []state.Group{},
	}

	view := GroupInsertionView(data)

	// Check that view contains key elements
	expectedStrings := []string{
		"Group Insertion",
		"Classifying",
		"clip01.mp4",
		"Enter new group name",
		"Test",
	}

	for _, expected := range expectedStrings {
		if !contains(view, expected) {
			t.Errorf("View missing expected string: %s\nView:\n%s", expected, view)
		}
	}
}

// TODO: Rewrite tests for new 3-mode flow (ModeInsertionChoice and ModeGroupSelection)
// The old position_selection mode tests are no longer valid with the new design
