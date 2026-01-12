// ui/group_selection_test.go
package ui

import (
	"clip-tagger/state"
	"strings"
	"testing"
)

func TestGroupSelectionData_Creation(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	group1 := state.NewGroup("Scene 1", 1)
	group2 := state.NewGroup("Scene 2", 2)
	group3 := state.NewGroup("Interview", 3)
	appState.Groups = append(appState.Groups, group1, group2, group3)

	data := NewGroupSelectionData(appState, "test_file.mp4")

	if data.CurrentFile != "test_file.mp4" {
		t.Errorf("expected CurrentFile to be 'test_file.mp4', got '%s'", data.CurrentFile)
	}
	if len(data.AllGroups) != 3 {
		t.Errorf("expected 3 groups, got %d", len(data.AllGroups))
	}
	if len(data.FilteredGroups) != 3 {
		t.Errorf("expected 3 filtered groups initially, got %d", len(data.FilteredGroups))
	}
	if data.FilterText != "" {
		t.Errorf("expected empty filter text, got '%s'", data.FilterText)
	}
	if data.SelectedIndex != 0 {
		t.Errorf("expected SelectedIndex to be 0, got %d", data.SelectedIndex)
	}
}

func TestGroupSelectionData_EmptyGroups(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)

	data := NewGroupSelectionData(appState, "test_file.mp4")

	if len(data.AllGroups) != 0 {
		t.Errorf("expected 0 groups, got %d", len(data.AllGroups))
	}
	if len(data.FilteredGroups) != 0 {
		t.Errorf("expected 0 filtered groups, got %d", len(data.FilteredGroups))
	}
}

func TestFilterGroups_CaseInsensitive(t *testing.T) {
	groups := []state.Group{
		state.NewGroup("Scene 1", 1),
		state.NewGroup("Scene 2", 2),
		state.NewGroup("Interview", 3),
		state.NewGroup("B-Roll", 4),
	}

	t.Run("lowercase filter matches uppercase", func(t *testing.T) {
		filtered := filterGroups(groups, "scene")
		if len(filtered) != 2 {
			t.Errorf("expected 2 filtered groups, got %d", len(filtered))
		}
	})

	t.Run("uppercase filter matches lowercase", func(t *testing.T) {
		filtered := filterGroups(groups, "INTERVIEW")
		if len(filtered) != 1 {
			t.Errorf("expected 1 filtered group, got %d", len(filtered))
		}
		if filtered[0].Name != "Interview" {
			t.Errorf("expected 'Interview', got '%s'", filtered[0].Name)
		}
	})

	t.Run("mixed case filter", func(t *testing.T) {
		filtered := filterGroups(groups, "RoLl")
		if len(filtered) != 1 {
			t.Errorf("expected 1 filtered group, got %d", len(filtered))
		}
		if filtered[0].Name != "B-Roll" {
			t.Errorf("expected 'B-Roll', got '%s'", filtered[0].Name)
		}
	})
}

func TestFilterGroups_SubstringMatching(t *testing.T) {
	groups := []state.Group{
		state.NewGroup("Opening Scene", 1),
		state.NewGroup("Middle Scene", 2),
		state.NewGroup("Closing Scene", 3),
		state.NewGroup("Interview", 4),
	}

	t.Run("substring in middle", func(t *testing.T) {
		filtered := filterGroups(groups, "scene")
		if len(filtered) != 3 {
			t.Errorf("expected 3 filtered groups, got %d", len(filtered))
		}
	})

	t.Run("substring at start", func(t *testing.T) {
		filtered := filterGroups(groups, "open")
		if len(filtered) != 1 {
			t.Errorf("expected 1 filtered group, got %d", len(filtered))
		}
		if filtered[0].Name != "Opening Scene" {
			t.Errorf("expected 'Opening Scene', got '%s'", filtered[0].Name)
		}
	})

	t.Run("substring at end", func(t *testing.T) {
		filtered := filterGroups(groups, "view")
		if len(filtered) != 1 {
			t.Errorf("expected 1 filtered group, got %d", len(filtered))
		}
		if filtered[0].Name != "Interview" {
			t.Errorf("expected 'Interview', got '%s'", filtered[0].Name)
		}
	})
}

func TestFilterGroups_EmptyFilter(t *testing.T) {
	groups := []state.Group{
		state.NewGroup("Scene 1", 1),
		state.NewGroup("Scene 2", 2),
	}

	filtered := filterGroups(groups, "")
	if len(filtered) != 2 {
		t.Errorf("expected all 2 groups with empty filter, got %d", len(filtered))
	}
}

func TestFilterGroups_NoMatches(t *testing.T) {
	groups := []state.Group{
		state.NewGroup("Scene 1", 1),
		state.NewGroup("Scene 2", 2),
	}

	filtered := filterGroups(groups, "xyz")
	if len(filtered) != 0 {
		t.Errorf("expected 0 filtered groups for non-matching filter, got %d", len(filtered))
	}
}

func TestGroupSelectionView_WithGroups(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	group1 := state.NewGroup("Scene 1", 1)
	group2 := state.NewGroup("Scene 2", 2)
	appState.Groups = append(appState.Groups, group1, group2)

	data := NewGroupSelectionData(appState, "test_file.mp4")
	view := GroupSelectionView(data)

	if view == "" {
		t.Fatal("expected non-empty view")
	}

	// Check for file context
	if !contains(view, "Classifying: test_file.mp4") {
		t.Error("expected view to show current file context")
	}

	// Check for groups with order
	if !contains(view, "Scene 1") {
		t.Error("expected view to show 'Scene 1'")
	}
	if !contains(view, "Scene 2") {
		t.Error("expected view to show 'Scene 2'")
	}
	if !contains(view, "[1]") {
		t.Error("expected view to show order number [1]")
	}
	if !contains(view, "[2]") {
		t.Error("expected view to show order number [2]")
	}

	// Check for filter input
	if !contains(view, "Filter:") {
		t.Error("expected view to show filter input")
	}

	// Check for instructions
	if !contains(view, "arrow keys") {
		t.Error("expected view to show navigation instructions")
	}
	if !contains(view, "Enter to select") {
		t.Error("expected view to show selection instructions")
	}
	if !contains(view, "Esc to cancel") {
		t.Error("expected view to show cancel instructions")
	}
}

func TestGroupSelectionView_EmptyGroups(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	data := NewGroupSelectionData(appState, "test_file.mp4")
	view := GroupSelectionView(data)

	if !contains(view, "No groups available") {
		t.Error("expected view to show 'No groups available' message")
	}
}

func TestGroupSelectionView_FilterActive(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	group1 := state.NewGroup("Scene 1", 1)
	group2 := state.NewGroup("Interview", 2)
	appState.Groups = append(appState.Groups, group1, group2)

	data := NewGroupSelectionData(appState, "test_file.mp4")
	data.FilterText = "scene"
	data.FilteredGroups = filterGroups(data.AllGroups, data.FilterText)

	view := GroupSelectionView(data)

	// Filter text should be shown
	if !contains(view, "scene") {
		t.Error("expected view to show filter text")
	}

	// Only filtered groups should appear
	if !contains(view, "Scene 1") {
		t.Error("expected view to show filtered group 'Scene 1'")
	}
	if contains(view, "Interview") {
		t.Error("expected view NOT to show filtered out group 'Interview'")
	}
}

func TestGroupSelectionView_Selection(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	group1 := state.NewGroup("Scene 1", 1)
	group2 := state.NewGroup("Scene 2", 2)
	appState.Groups = append(appState.Groups, group1, group2)

	data := NewGroupSelectionData(appState, "test_file.mp4")
	data.SelectedIndex = 0

	view := GroupSelectionView(data)

	// First item should have selection indicator
	lines := strings.Split(view, "\n")
	foundSelection := false
	for _, line := range lines {
		if contains(line, "Scene 1") {
			if contains(line, ">") || contains(line, "*") {
				foundSelection = true
			}
		}
	}

	if !foundSelection {
		t.Error("expected view to show selection indicator on first item")
	}
}

func TestGroupSelectionUpdate_ArrowKeys(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	group1 := state.NewGroup("Scene 1", 1)
	group2 := state.NewGroup("Scene 2", 2)
	group3 := state.NewGroup("Scene 3", 3)
	appState.Groups = append(appState.Groups, group1, group2, group3)

	data := NewGroupSelectionData(appState, "test_file.mp4")

	t.Run("down arrow moves down", func(t *testing.T) {
		result := GroupSelectionUpdate(data, "down")
		if result.Screen != -2 {
			t.Errorf("expected no screen change, got %v", result.Screen)
		}
		if data.SelectedIndex != 1 {
			t.Errorf("expected SelectedIndex to be 1, got %d", data.SelectedIndex)
		}
	})

	t.Run("down arrow at bottom stays at bottom", func(t *testing.T) {
		data.SelectedIndex = 2
		result := GroupSelectionUpdate(data, "down")
		if result.Screen != -2 {
			t.Errorf("expected no screen change, got %v", result.Screen)
		}
		if data.SelectedIndex != 2 {
			t.Errorf("expected SelectedIndex to stay at 2, got %d", data.SelectedIndex)
		}
	})

	t.Run("up arrow moves up", func(t *testing.T) {
		data.SelectedIndex = 2
		result := GroupSelectionUpdate(data, "up")
		if result.Screen != -2 {
			t.Errorf("expected no screen change, got %v", result.Screen)
		}
		if data.SelectedIndex != 1 {
			t.Errorf("expected SelectedIndex to be 1, got %d", data.SelectedIndex)
		}
	})

	t.Run("up arrow at top stays at top", func(t *testing.T) {
		data.SelectedIndex = 0
		result := GroupSelectionUpdate(data, "up")
		if result.Screen != -2 {
			t.Errorf("expected no screen change, got %v", result.Screen)
		}
		if data.SelectedIndex != 0 {
			t.Errorf("expected SelectedIndex to stay at 0, got %d", data.SelectedIndex)
		}
	})
}

func TestGroupSelectionUpdate_Enter(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	group1 := state.NewGroup("Scene 1", 1)
	group2 := state.NewGroup("Scene 2", 2)
	appState.Groups = append(appState.Groups, group1, group2)

	data := NewGroupSelectionData(appState, "test_file.mp4")
	data.SelectedIndex = 1

	result := GroupSelectionUpdate(data, "enter")

	if result.Screen != ScreenClassification {
		t.Errorf("expected transition to ScreenClassification, got %v", result.Screen)
	}
	if result.SelectedGroupID == "" {
		t.Error("expected SelectedGroupID to be set")
	}
	if result.SelectedGroupID != group2.ID {
		t.Errorf("expected SelectedGroupID to be %s, got %s", group2.ID, result.SelectedGroupID)
	}
	if result.SelectedGroupName != "Scene 2" {
		t.Errorf("expected SelectedGroupName to be 'Scene 2', got '%s'", result.SelectedGroupName)
	}
}

func TestGroupSelectionUpdate_Enter_NoGroups(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	data := NewGroupSelectionData(appState, "test_file.mp4")

	result := GroupSelectionUpdate(data, "enter")

	// Should not crash, should stay on screen
	if result.Screen != -2 {
		t.Errorf("expected no screen change when no groups, got %v", result.Screen)
	}
}

func TestGroupSelectionUpdate_Escape(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	group1 := state.NewGroup("Scene 1", 1)
	appState.Groups = append(appState.Groups, group1)

	data := NewGroupSelectionData(appState, "test_file.mp4")

	result := GroupSelectionUpdate(data, "esc")

	if result.Screen != ScreenClassification {
		t.Errorf("expected transition to ScreenClassification, got %v", result.Screen)
	}
	if result.SelectedGroupID != "" {
		t.Error("expected SelectedGroupID to be empty on cancel")
	}
	if result.SelectedGroupName != "" {
		t.Error("expected SelectedGroupName to be empty on cancel")
	}
}

func TestGroupSelectionUpdate_TextInput(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	group1 := state.NewGroup("Scene 1", 1)
	group2 := state.NewGroup("Interview", 2)
	appState.Groups = append(appState.Groups, group1, group2)

	data := NewGroupSelectionData(appState, "test_file.mp4")

	t.Run("typing letter adds to filter", func(t *testing.T) {
		result := GroupSelectionUpdate(data, "s")
		if result.Screen != -2 {
			t.Errorf("expected no screen change, got %v", result.Screen)
		}
		if data.FilterText != "s" {
			t.Errorf("expected FilterText to be 's', got '%s'", data.FilterText)
		}
		if len(data.FilteredGroups) != 1 {
			t.Errorf("expected 1 filtered group, got %d", len(data.FilteredGroups))
		}
		if data.SelectedIndex != 0 {
			t.Errorf("expected SelectedIndex to reset to 0, got %d", data.SelectedIndex)
		}
	})

	t.Run("typing more letters continues filtering", func(t *testing.T) {
		data.FilterText = "s"
		result := GroupSelectionUpdate(data, "c")
		if result.Screen != -2 {
			t.Errorf("expected no screen change, got %v", result.Screen)
		}
		if data.FilterText != "sc" {
			t.Errorf("expected FilterText to be 'sc', got '%s'", data.FilterText)
		}
		if len(data.FilteredGroups) != 1 {
			t.Errorf("expected 1 filtered group, got %d", len(data.FilteredGroups))
		}
	})

	t.Run("typing space is allowed", func(t *testing.T) {
		data.FilterText = ""
		data.FilteredGroups = data.AllGroups
		result := GroupSelectionUpdate(data, " ")
		if result.Screen != -2 {
			t.Errorf("expected no screen change, got %v", result.Screen)
		}
		if data.FilterText != " " {
			t.Errorf("expected FilterText to be ' ', got '%s'", data.FilterText)
		}
	})
}

func TestGroupSelectionUpdate_Backspace(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	group1 := state.NewGroup("Scene 1", 1)
	appState.Groups = append(appState.Groups, group1)

	data := NewGroupSelectionData(appState, "test_file.mp4")
	data.FilterText = "scene"
	data.FilteredGroups = filterGroups(data.AllGroups, data.FilterText)

	result := GroupSelectionUpdate(data, "backspace")

	if result.Screen != -2 {
		t.Errorf("expected no screen change, got %v", result.Screen)
	}
	if data.FilterText != "scen" {
		t.Errorf("expected FilterText to be 'scen', got '%s'", data.FilterText)
	}
	if len(data.FilteredGroups) != 1 {
		t.Errorf("expected filtered groups to be updated, got %d", len(data.FilteredGroups))
	}
}

func TestGroupSelectionUpdate_Backspace_EmptyFilter(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	group1 := state.NewGroup("Scene 1", 1)
	appState.Groups = append(appState.Groups, group1)

	data := NewGroupSelectionData(appState, "test_file.mp4")
	data.FilterText = ""

	result := GroupSelectionUpdate(data, "backspace")

	if result.Screen != -2 {
		t.Errorf("expected no screen change, got %v", result.Screen)
	}
	if data.FilterText != "" {
		t.Errorf("expected FilterText to remain empty, got '%s'", data.FilterText)
	}
}

func TestGroupSelectionUpdate_Quit(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	data := NewGroupSelectionData(appState, "test_file.mp4")

	result := GroupSelectionUpdate(data, "ctrl+c")

	if result.Screen != -1 {
		t.Errorf("expected quit signal (-1), got %v", result.Screen)
	}
}

func TestGroupSelectionUpdate_FilteringResetsSelection(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	group1 := state.NewGroup("Scene 1", 1)
	group2 := state.NewGroup("Scene 2", 2)
	group3 := state.NewGroup("Interview", 3)
	appState.Groups = append(appState.Groups, group1, group2, group3)

	data := NewGroupSelectionData(appState, "test_file.mp4")
	data.SelectedIndex = 2 // Select the third item

	// Type a filter that changes the list
	GroupSelectionUpdate(data, "s")

	// Selection should reset to 0
	if data.SelectedIndex != 0 {
		t.Errorf("expected SelectedIndex to reset to 0 after filtering, got %d", data.SelectedIndex)
	}
}
