// ui/model_group_insertion_test.go
package ui

import (
	"clip-tagger/state"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModel_GroupInsertion_Integration(t *testing.T) {
	// Create initial state with some existing groups
	appState := state.NewState("/test", state.SortByName)
	appState.Groups = []state.Group{
		{ID: "g1", Name: "Group A", Order: 1},
		{ID: "g2", Name: "Group B", Order: 2},
	}

	files := []string{"clip01.mp4", "clip02.mp4"}
	m := Model{
		state:            appState,
		currentScreen:    ScreenClassification,
		files:            files,
		currentFileIndex: 0,
	}

	// Initialize classification screen
	model, _ := m.Update(ClassificationInitialized{
		Files:     files,
		FileIndex: 0,
	})
	m = model.(Model)

	// Press '3' to transition to group insertion
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m = model.(Model)

	// Should transition to group insertion screen
	if m.currentScreen != ScreenGroupInsertion {
		t.Errorf("Expected screen to be ScreenGroupInsertion, got %v", m.currentScreen)
	}

	// Execute the command to initialize group insertion data
	if cmd != nil {
		msg := cmd()
		model, _ = m.Update(msg)
		m = model.(Model)
	}

	// Verify group insertion data was initialized
	if m.groupInsertionData == nil {
		t.Fatalf("Expected groupInsertionData to be initialized")
	}

	if m.groupInsertionData.CurrentFile != "clip01.mp4" {
		t.Errorf("Expected CurrentFile 'clip01.mp4', got %s", m.groupInsertionData.CurrentFile)
	}

	if m.groupInsertionData.Mode != "name_entry" {
		t.Errorf("Expected Mode 'name_entry', got %s", m.groupInsertionData.Mode)
	}

	// Type a group name: "New Group"
	for _, char := range "New Group" {
		model, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}})
		m = model.(Model)
	}

	if m.groupInsertionData.GroupName != "New Group" {
		t.Errorf("Expected GroupName 'New Group', got %s", m.groupInsertionData.GroupName)
	}

	// Press Enter to proceed to position selection
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(Model)

	if m.groupInsertionData.Mode != "position_selection" {
		t.Errorf("Expected Mode 'position_selection', got %s", m.groupInsertionData.Mode)
	}

	// Press down arrow to move to position 1 (between groups)
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(Model)

	if m.groupInsertionData.SelectedPosition != 1 {
		t.Errorf("Expected SelectedPosition 1, got %d", m.groupInsertionData.SelectedPosition)
	}

	// Press Enter to confirm position
	model, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(Model)

	// Should transition back to classification screen
	if m.currentScreen != ScreenClassification {
		t.Errorf("Expected screen to be ScreenClassification, got %v", m.currentScreen)
	}

	// Execute the command to handle group insertion
	if cmd != nil {
		msg := cmd()
		model, _ = m.Update(msg)
		m = model.(Model)
	}

	// Verify group was inserted in state
	if len(m.state.Groups) != 3 {
		t.Fatalf("Expected 3 groups, got %d", len(m.state.Groups))
	}

	// Verify groups are in correct order
	expectedGroups := []struct {
		name  string
		order int
	}{
		{"Group A", 1},
		{"New Group", 2},
		{"Group B", 3},
	}

	for i, expected := range expectedGroups {
		if m.state.Groups[i].Name != expected.name {
			t.Errorf("Expected group[%d].Name '%s', got '%s'", i, expected.name, m.state.Groups[i].Name)
		}
		if m.state.Groups[i].Order != expected.order {
			t.Errorf("Expected group[%d].Order %d, got %d", i, expected.order, m.state.Groups[i].Order)
		}
	}
}

func TestModel_GroupInsertion_EmptyGroups(t *testing.T) {
	// Create initial state with no groups
	appState := state.NewState("/test", state.SortByName)

	files := []string{"clip01.mp4"}
	m := Model{
		state:            appState,
		currentScreen:    ScreenClassification,
		files:            files,
		currentFileIndex: 0,
	}

	// Initialize classification screen
	model, _ := m.Update(ClassificationInitialized{
		Files:     files,
		FileIndex: 0,
	})
	m = model.(Model)

	// Press '3' to transition to group insertion
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m = model.(Model)

	// Execute initialization command
	if cmd != nil {
		msg := cmd()
		model, _ = m.Update(msg)
		m = model.(Model)
	}

	// Type a group name
	for _, char := range "First Group" {
		model, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}})
		m = model.(Model)
	}

	// Press Enter to proceed to position selection
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(Model)

	// Should be in position selection mode
	if m.groupInsertionData.Mode != "position_selection" {
		t.Errorf("Expected Mode 'position_selection', got %s", m.groupInsertionData.Mode)
	}

	// Press Enter to confirm (only position available is 0)
	model, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(Model)

	// Execute the command to handle group insertion
	if cmd != nil {
		msg := cmd()
		model, _ = m.Update(msg)
		m = model.(Model)
	}

	// Verify group was inserted
	if len(m.state.Groups) != 1 {
		t.Fatalf("Expected 1 group, got %d", len(m.state.Groups))
	}

	if m.state.Groups[0].Name != "First Group" {
		t.Errorf("Expected group name 'First Group', got '%s'", m.state.Groups[0].Name)
	}

	if m.state.Groups[0].Order != 1 {
		t.Errorf("Expected group order 1, got %d", m.state.Groups[0].Order)
	}
}

func TestModel_GroupInsertion_Cancel(t *testing.T) {
	// Create initial state
	appState := state.NewState("/test", state.SortByName)
	appState.Groups = []state.Group{
		{ID: "g1", Name: "Group A", Order: 1},
	}

	files := []string{"clip01.mp4"}
	m := Model{
		state:            appState,
		currentScreen:    ScreenClassification,
		files:            files,
		currentFileIndex: 0,
	}

	// Initialize classification screen
	model, _ := m.Update(ClassificationInitialized{
		Files:     files,
		FileIndex: 0,
	})
	m = model.(Model)

	// Press '3' to transition to group insertion
	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m = model.(Model)

	// Execute initialization command
	if cmd != nil {
		msg := cmd()
		model, _ = m.Update(msg)
		m = model.(Model)
	}

	// Type a group name
	for _, char := range "Test" {
		model, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}})
		m = model.(Model)
	}

	// Press Esc to cancel
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = model.(Model)

	// Should transition back to classification screen
	if m.currentScreen != ScreenClassification {
		t.Errorf("Expected screen to be ScreenClassification, got %v", m.currentScreen)
	}

	// Verify no group was added
	if len(m.state.Groups) != 1 {
		t.Errorf("Expected 1 group (no change), got %d", len(m.state.Groups))
	}
}

func TestModel_GroupInsertion_InsertAtBeginning(t *testing.T) {
	// Create initial state
	appState := state.NewState("/test", state.SortByName)
	appState.Groups = []state.Group{
		{ID: "g1", Name: "Group B", Order: 1},
		{ID: "g2", Name: "Group C", Order: 2},
	}

	files := []string{"clip01.mp4"}
	m := Model{
		state:            appState,
		currentScreen:    ScreenClassification,
		files:            files,
		currentFileIndex: 0,
	}

	// Initialize and navigate to group insertion
	model, _ := m.Update(ClassificationInitialized{Files: files, FileIndex: 0})
	m = model.(Model)

	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m = model.(Model)

	if cmd != nil {
		model, _ = m.Update(cmd())
		m = model.(Model)
	}

	// Type "Group A"
	for _, char := range "Group A" {
		model, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}})
		m = model.(Model)
	}

	// Enter to position selection
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(Model)

	// Position 0 is already selected (at beginning)
	// Press Enter to confirm
	model, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(Model)

	if cmd != nil {
		model, _ = m.Update(cmd())
		m = model.(Model)
	}

	// Verify order
	if len(m.state.Groups) != 3 {
		t.Fatalf("Expected 3 groups, got %d", len(m.state.Groups))
	}

	expected := []string{"Group A", "Group B", "Group C"}
	for i, name := range expected {
		if m.state.Groups[i].Name != name {
			t.Errorf("Expected group[%d].Name '%s', got '%s'", i, name, m.state.Groups[i].Name)
		}
		if m.state.Groups[i].Order != i+1 {
			t.Errorf("Expected group[%d].Order %d, got %d", i, i+1, m.state.Groups[i].Order)
		}
	}
}

func TestModel_GroupInsertion_InsertAtEnd(t *testing.T) {
	// Create initial state
	appState := state.NewState("/test", state.SortByName)
	appState.Groups = []state.Group{
		{ID: "g1", Name: "Group A", Order: 1},
		{ID: "g2", Name: "Group B", Order: 2},
	}

	files := []string{"clip01.mp4"}
	m := Model{
		state:            appState,
		currentScreen:    ScreenClassification,
		files:            files,
		currentFileIndex: 0,
	}

	// Initialize and navigate to group insertion
	model, _ := m.Update(ClassificationInitialized{Files: files, FileIndex: 0})
	m = model.(Model)

	model, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m = model.(Model)

	if cmd != nil {
		model, _ = m.Update(cmd())
		m = model.(Model)
	}

	// Type "Group C"
	for _, char := range "Group C" {
		model, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}})
		m = model.(Model)
	}

	// Enter to position selection
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(Model)

	// Move to position 2 (at end)
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(Model)
	model, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = model.(Model)

	// Press Enter to confirm
	model, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = model.(Model)

	if cmd != nil {
		model, _ = m.Update(cmd())
		m = model.(Model)
	}

	// Verify order
	if len(m.state.Groups) != 3 {
		t.Fatalf("Expected 3 groups, got %d", len(m.state.Groups))
	}

	expected := []string{"Group A", "Group B", "Group C"}
	for i, name := range expected {
		if m.state.Groups[i].Name != name {
			t.Errorf("Expected group[%d].Name '%s', got '%s'", i, name, m.state.Groups[i].Name)
		}
		if m.state.Groups[i].Order != i+1 {
			t.Errorf("Expected group[%d].Order %d, got %d", i, i+1, m.state.Groups[i].Order)
		}
	}
}

func TestInsertGroupAtPosition(t *testing.T) {
	tests := []struct {
		name           string
		initialGroups  []state.Group
		newGroup       state.Group
		order          int
		expectedOrder  []string
		expectedOrders []int
	}{
		{
			name:          "insert into empty list",
			initialGroups: []state.Group{},
			newGroup:      state.Group{ID: "new", Name: "New Group", Order: 1},
			order:         1,
			expectedOrder: []string{"New Group"},
			expectedOrders: []int{1},
		},
		{
			name: "insert at beginning",
			initialGroups: []state.Group{
				{ID: "1", Name: "B", Order: 1},
				{ID: "2", Name: "C", Order: 2},
			},
			newGroup:       state.Group{ID: "new", Name: "A", Order: 1},
			order:          1,
			expectedOrder:  []string{"A", "B", "C"},
			expectedOrders: []int{1, 2, 3},
		},
		{
			name: "insert in middle",
			initialGroups: []state.Group{
				{ID: "1", Name: "A", Order: 1},
				{ID: "2", Name: "C", Order: 2},
			},
			newGroup:       state.Group{ID: "new", Name: "B", Order: 2},
			order:          2,
			expectedOrder:  []string{"A", "B", "C"},
			expectedOrders: []int{1, 2, 3},
		},
		{
			name: "insert at end",
			initialGroups: []state.Group{
				{ID: "1", Name: "A", Order: 1},
				{ID: "2", Name: "B", Order: 2},
			},
			newGroup:       state.Group{ID: "new", Name: "C", Order: 3},
			order:          3,
			expectedOrder:  []string{"A", "B", "C"},
			expectedOrders: []int{1, 2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups := make([]state.Group, len(tt.initialGroups))
			copy(groups, tt.initialGroups)

			insertGroupAtPosition(&groups, tt.newGroup, tt.order)

			if len(groups) != len(tt.expectedOrder) {
				t.Fatalf("Expected %d groups, got %d", len(tt.expectedOrder), len(groups))
			}

			for i := range groups {
				if groups[i].Name != tt.expectedOrder[i] {
					t.Errorf("Expected group[%d].Name '%s', got '%s'", i, tt.expectedOrder[i], groups[i].Name)
				}
				if groups[i].Order != tt.expectedOrders[i] {
					t.Errorf("Expected group[%d].Order %d, got %d", i, tt.expectedOrders[i], groups[i].Order)
				}
			}
		})
	}
}
