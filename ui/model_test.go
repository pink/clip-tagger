// ui/model_test.go
package ui

import (
	"clip-tagger/state"
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	model := NewModel(appState, "/test/dir")

	if model.state != appState {
		t.Error("expected model to have correct state reference")
	}
	if model.currentScreen != ScreenStartup {
		t.Errorf("expected initial screen to be ScreenStartup, got %v", model.currentScreen)
	}
	if model.directory != "/test/dir" {
		t.Errorf("expected directory to be /test/dir, got %s", model.directory)
	}
}

func TestScreen_String(t *testing.T) {
	tests := []struct {
		screen   Screen
		expected string
	}{
		{ScreenStartup, "startup"},
		{ScreenClassification, "classification"},
		{ScreenGroupSelection, "group_selection"},
		{ScreenGroupInsertion, "group_insertion"},
		{ScreenReview, "review"},
		{ScreenComplete, "complete"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.screen.String()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestModel_Init(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	model := NewModel(appState, "/test/dir")

	// Init should return a command for initialization
	cmd := model.Init()
	if cmd == nil {
		t.Error("expected Init to return a command")
	}
}

func TestModel_Update_ScreenTransitions(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	model := NewModel(appState, "/test/dir")

	t.Run("transition to classification screen", func(t *testing.T) {
		msg := TransitionToScreen{Screen: ScreenClassification}
		updated, _ := model.Update(msg)
		updatedModel := updated.(Model)

		if updatedModel.currentScreen != ScreenClassification {
			t.Errorf("expected screen to be ScreenClassification, got %v", updatedModel.currentScreen)
		}
	})

	t.Run("transition to group selection screen", func(t *testing.T) {
		msg := TransitionToScreen{Screen: ScreenGroupSelection}
		updated, _ := model.Update(msg)
		updatedModel := updated.(Model)

		if updatedModel.currentScreen != ScreenGroupSelection {
			t.Errorf("expected screen to be ScreenGroupSelection, got %v", updatedModel.currentScreen)
		}
	})

	t.Run("transition to complete screen", func(t *testing.T) {
		msg := TransitionToScreen{Screen: ScreenComplete}
		updated, _ := model.Update(msg)
		updatedModel := updated.(Model)

		if updatedModel.currentScreen != ScreenComplete {
			t.Errorf("expected screen to be ScreenComplete, got %v", updatedModel.currentScreen)
		}
	})
}

func TestModel_Update_QuitMessage(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	model := NewModel(appState, "/test/dir")

	t.Run("quit with ctrl+c", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyCtrlC}
		_, cmd := model.Update(msg)

		if cmd == nil {
			t.Fatal("expected quit command to be returned")
		}

		// Note: We can't easily test that it's exactly tea.Quit without exposing internals,
		// but we verify a command is returned
	})

	t.Run("no quit on regular key", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		_, cmd := model.Update(msg)

		// For now, enter key should not do anything in base model
		// (screen-specific handlers will be added later)
		// We can only check that cmd is nil for non-quit operations
		if cmd != nil {
			t.Error("expected regular key to return nil command")
		}
	})
}

func TestModel_View(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	model := NewModel(appState, "/test/dir")

	t.Run("view returns non-empty string", func(t *testing.T) {
		view := model.View()
		if view == "" {
			t.Error("expected View to return non-empty string")
		}
	})

	t.Run("view includes screen name", func(t *testing.T) {
		model.currentScreen = ScreenStartup
		view := model.View()
		// Should show something about the current screen
		if len(view) < 5 {
			t.Error("expected View to return meaningful content")
		}
	})

	t.Run("view changes with screen", func(t *testing.T) {
		model.currentScreen = ScreenStartup
		startupView := model.View()

		model.currentScreen = ScreenClassification
		classificationView := model.View()

		// Views should be different for different screens
		if startupView == classificationView {
			t.Error("expected different views for different screens")
		}
	})
}

func TestModel_StateUpdate(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	model := NewModel(appState, "/test/dir")

	t.Run("state update message updates state", func(t *testing.T) {
		newState := state.NewState("/new/dir", state.SortByName)
		msg := StateUpdate{State: newState}

		updated, _ := model.Update(msg)
		updatedModel := updated.(Model)

		if updatedModel.state.Directory != "/new/dir" {
			t.Errorf("expected directory to be updated to '/new/dir', got '%s'", updatedModel.state.Directory)
		}
	})
}

func TestAllScreensEnumerated(t *testing.T) {
	// Verify all 6 screens are defined
	screens := []Screen{
		ScreenStartup,
		ScreenClassification,
		ScreenGroupSelection,
		ScreenGroupInsertion,
		ScreenReview,
		ScreenComplete,
	}

	// Verify each screen has a unique value
	seen := make(map[Screen]bool)
	for _, screen := range screens {
		if seen[screen] {
			t.Errorf("duplicate screen value: %v", screen)
		}
		seen[screen] = true
	}

	if len(screens) != 6 {
		t.Errorf("expected 6 screens, found %d", len(screens))
	}
}

func TestModel_PreviewAction_FileNotFound(t *testing.T) {
	// Test preview action when file doesn't exist
	appState := state.NewState("/tmp", state.SortByModifiedTime)
	model := NewModel(appState, "/tmp")

	// Set up classification screen with non-existent file
	files := []string{"nonexistent-file.mp4"}
	model.files = files
	model.currentFileIndex = 0
	model.currentScreen = ScreenClassification
	model.classificationData = NewClassificationData(appState, files, 0)

	// Trigger preview action with 'p' key
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	updated, _ := model.Update(keyMsg)
	updatedModel := updated.(Model)

	// Should set error message
	if updatedModel.err == "" {
		t.Error("expected error message when previewing non-existent file")
	}
	if !contains(updatedModel.err, "Failed to preview file") {
		t.Errorf("expected error message to contain 'Failed to preview file', got: %s", updatedModel.err)
	}
}

func TestModel_PreviewAction_Success(t *testing.T) {
	// Test preview action with existing file
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-video.mp4")

	f, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	f.Close()

	appState := state.NewState(tmpDir, state.SortByModifiedTime)
	model := NewModel(appState, tmpDir)

	// Set up classification screen
	files := []string{"test-video.mp4"}
	model.files = files
	model.currentFileIndex = 0
	model.currentScreen = ScreenClassification
	model.classificationData = NewClassificationData(appState, files, 0)

	// Trigger preview action with 'p' key
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	updated, _ := model.Update(keyMsg)
	updatedModel := updated.(Model)

	// On systems with a default player, no error should be set
	// On headless/CI systems, an error may be set (acceptable)
	if updatedModel.err != "" {
		t.Logf("Preview returned error (may be expected in test environment): %s", updatedModel.err)
	}

	// Should remain on classification screen
	if updatedModel.currentScreen != ScreenClassification {
		t.Errorf("expected to remain on classification screen, got %v", updatedModel.currentScreen)
	}
}

func TestModel_PreviewAction_ScreenNoChange(t *testing.T) {
	// Test that preview action doesn't change screen
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.mp4")

	f, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	f.Close()

	appState := state.NewState(tmpDir, state.SortByModifiedTime)
	model := NewModel(appState, tmpDir)

	// Set up classification screen
	files := []string{"test.mp4"}
	model.files = files
	model.currentFileIndex = 0
	model.currentScreen = ScreenClassification
	model.classificationData = NewClassificationData(appState, files, 0)

	// Trigger preview action
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	updated, cmd := model.Update(keyMsg)
	updatedModel := updated.(Model)

	// Should remain on classification screen
	if updatedModel.currentScreen != ScreenClassification {
		t.Errorf("expected screen to remain ScreenClassification, got %v", updatedModel.currentScreen)
	}

	// Should not return any command (no screen transition)
	if cmd != nil {
		t.Error("expected no command for preview action (stays on same screen)")
	}
}

// TestModel_AutoSaveOnGroupSelected tests that state is saved after a group is selected
func TestModel_AutoSaveOnGroupSelected(t *testing.T) {
	tmpDir := t.TempDir()
	appState := state.NewState(tmpDir, state.SortByModifiedTime)
	appState.Groups = append(appState.Groups, state.NewGroup("intro", 1))

	model := NewModel(appState, tmpDir)
	model.currentScreen = ScreenClassification

	// Send GroupSelected message
	msg := GroupSelected{
		GroupID:   appState.Groups[0].ID,
		GroupName: "intro",
	}

	_, _ = model.Update(msg)

	// Verify state file was created
	statePath := state.StateFilePath(tmpDir)
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Error("expected state file to be created after GroupSelected")
	}

	// Verify state can be loaded
	loaded, err := state.Load(statePath)
	if err != nil {
		t.Fatalf("failed to load saved state: %v", err)
	}

	if len(loaded.Groups) != 1 {
		t.Errorf("expected 1 group in loaded state, got %d", len(loaded.Groups))
	}
}

// TestModel_AutoSaveOnGroupInserted tests that state is saved after a group is inserted
func TestModel_AutoSaveOnGroupInserted(t *testing.T) {
	tmpDir := t.TempDir()
	appState := state.NewState(tmpDir, state.SortByModifiedTime)
	appState.Groups = append(appState.Groups, state.NewGroup("intro", 1))

	model := NewModel(appState, tmpDir)
	model.currentScreen = ScreenClassification

	// Send GroupInserted message
	newGroup := state.NewGroup("outro", 2)
	msg := GroupInserted{
		GroupID:   newGroup.ID,
		GroupName: newGroup.Name,
		Order:     2,
	}

	_, _ = model.Update(msg)

	// Verify state file was created
	statePath := state.StateFilePath(tmpDir)
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Error("expected state file to be created after GroupInserted")
	}

	// Verify state can be loaded and has 2 groups
	loaded, err := state.Load(statePath)
	if err != nil {
		t.Fatalf("failed to load saved state: %v", err)
	}

	if len(loaded.Groups) != 2 {
		t.Errorf("expected 2 groups in loaded state, got %d", len(loaded.Groups))
	}
}

// TestModel_SaveErrorHandling tests that save errors are handled gracefully
func TestModel_SaveErrorHandling(t *testing.T) {
	// Use a read-only directory to force a save error
	tmpDir := t.TempDir()
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	err := os.Mkdir(readOnlyDir, 0444) // read-only permissions
	if err != nil {
		t.Fatalf("failed to create read-only dir: %v", err)
	}

	appState := state.NewState(readOnlyDir, state.SortByModifiedTime)
	appState.Groups = append(appState.Groups, state.NewGroup("intro", 1))

	model := NewModel(appState, readOnlyDir)
	model.currentScreen = ScreenClassification

	// Send GroupSelected message - should fail to save
	msg := GroupSelected{
		GroupID:   appState.Groups[0].ID,
		GroupName: "intro",
	}

	updated, _ := model.Update(msg)
	updatedModel := updated.(Model)

	// Should have error message but not crash
	if updatedModel.err == "" {
		t.Error("expected error message when save fails")
	}

	if !contains(updatedModel.err, "Failed to save state") {
		t.Errorf("expected error message to contain 'Failed to save state', got: %s", updatedModel.err)
	}

	// Should still be able to continue using the app (not crashed)
	if updatedModel.state == nil {
		t.Error("expected state to still be available after save error")
	}
}

// TestModel_StatePersistenceAcrossRestarts tests that state can be persisted and loaded
func TestModel_StatePersistenceAcrossRestarts(t *testing.T) {
	tmpDir := t.TempDir()

	// Create initial state and save it
	originalState := state.NewState(tmpDir, state.SortByModifiedTime)
	originalState.Groups = append(originalState.Groups, state.NewGroup("intro", 1))
	originalState.Groups = append(originalState.Groups, state.NewGroup("outro", 2))
	originalState.CurrentIndex = 5

	model := NewModel(originalState, tmpDir)

	// Trigger a save by sending GroupSelected message
	msg := GroupSelected{
		GroupID:   originalState.Groups[0].ID,
		GroupName: "intro",
	}
	_, _ = model.Update(msg)

	// Verify state file exists
	statePath := state.StateFilePath(tmpDir)
	if !state.StateExists(tmpDir) {
		t.Fatal("expected state file to exist")
	}

	// Load state (simulating app restart)
	loaded, err := state.Load(statePath)
	if err != nil {
		t.Fatalf("failed to load state: %v", err)
	}

	// Verify loaded state matches original
	if loaded.Directory != originalState.Directory {
		t.Errorf("directory mismatch: expected %s, got %s", originalState.Directory, loaded.Directory)
	}
	if loaded.CurrentIndex != originalState.CurrentIndex {
		t.Errorf("index mismatch: expected %d, got %d", originalState.CurrentIndex, loaded.CurrentIndex)
	}
	if len(loaded.Groups) != len(originalState.Groups) {
		t.Errorf("groups count mismatch: expected %d, got %d", len(originalState.Groups), len(loaded.Groups))
	}
}

// TestModel_StateFileLocation tests that state file is created in the correct location
func TestModel_StateFileLocation(t *testing.T) {
	tmpDir := t.TempDir()
	appState := state.NewState(tmpDir, state.SortByModifiedTime)
	appState.Groups = append(appState.Groups, state.NewGroup("intro", 1))

	model := NewModel(appState, tmpDir)

	// Trigger a save
	msg := GroupSelected{
		GroupID:   appState.Groups[0].ID,
		GroupName: "intro",
	}
	_, _ = model.Update(msg)

	// Verify state file is in the working directory with correct name
	expectedPath := filepath.Join(tmpDir, ".clip-tagger-state.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected state file at %s, but it does not exist", expectedPath)
	}
}

// TestModel_NoSaveOnNonStateChangingActions tests that state is NOT saved for actions that don't change state
func TestModel_NoSaveOnNonStateChangingActions(t *testing.T) {
	tmpDir := t.TempDir()
	appState := state.NewState(tmpDir, state.SortByModifiedTime)

	model := NewModel(appState, tmpDir)
	model.currentScreen = ScreenStartup

	// Send a screen transition message from startup to review (shouldn't trigger save)
	msg := TransitionToScreen{Screen: ScreenReview}
	_, _ = model.Update(msg)

	// Verify state file was NOT created
	statePath := state.StateFilePath(tmpDir)
	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Error("expected state file to NOT be created for screen transition from non-classification screen")
	}
}

// TestModel_AutoSaveOnClassificationActionSameAsLast tests that state is saved after "Same as Last" action
func TestModel_AutoSaveOnClassificationActionSameAsLast(t *testing.T) {
	tmpDir := t.TempDir()
	appState := state.NewState(tmpDir, state.SortByModifiedTime)
	group := state.NewGroup("intro", 1)
	appState.Groups = append(appState.Groups, group)

	// Add a previous classification so "Same as Last" is available
	appState.Classifications = append(appState.Classifications, state.Classification{
		File:       "file1.mp4",
		GroupID:    group.ID,
		TakeNumber: 1,
	})

	model := NewModel(appState, tmpDir)
	model.currentScreen = ScreenClassification
	model.files = []string{"file1.mp4", "file2.mp4"}
	model.currentFileIndex = 1 // Currently on file2.mp4, file1.mp4 has classification
	model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex)
	model.actionsPerSave = 1 // Save after every action for this test

	// Verify the classification data has previous classification (prerequisite)
	if !model.classificationData.HasPreviousClassification {
		t.Fatal("expected classificationData to have previous classification for this test to work")
	}

	// Trigger "Same as Last" action with '1' key
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}
	_, _ = model.Update(keyMsg)

	// Verify state file was created
	statePath := state.StateFilePath(tmpDir)
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Error("expected state file to be created after ClassificationActionSameAsLast")
	}

	// Verify state can be loaded
	loaded, err := state.Load(statePath)
	if err != nil {
		t.Fatalf("failed to load saved state: %v", err)
	}

	if len(loaded.Groups) != 1 {
		t.Errorf("expected 1 group in loaded state, got %d", len(loaded.Groups))
	}
}

// TestModel_AutoSaveOnClassificationActionSkip tests that state is saved after "Skip" action
func TestModel_AutoSaveOnClassificationActionSkip(t *testing.T) {
	tmpDir := t.TempDir()
	appState := state.NewState(tmpDir, state.SortByModifiedTime)
	appState.Groups = append(appState.Groups, state.NewGroup("scene1", 1))

	model := NewModel(appState, tmpDir)
	model.currentScreen = ScreenClassification
	model.files = []string{"file1.mp4"}
	model.currentFileIndex = 0
	model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex)
	model.actionsPerSave = 1 // Save after every action for this test

	// Trigger "Skip" action with 's' key
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	_, _ = model.Update(keyMsg)

	// Verify state file was created
	statePath := state.StateFilePath(tmpDir)
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Error("expected state file to be created after ClassificationActionSkip")
	}

	// Verify state can be loaded
	loaded, err := state.Load(statePath)
	if err != nil {
		t.Fatalf("failed to load saved state: %v", err)
	}

	if len(loaded.Groups) != 1 {
		t.Errorf("expected 1 group in loaded state, got %d", len(loaded.Groups))
	}
}

// TestModel_PeriodicAutoSave tests that state is saved every N actions
func TestModel_PeriodicAutoSave(t *testing.T) {
	tmpDir := t.TempDir()
	appState := state.NewState(tmpDir, state.SortByModifiedTime)
	appState.Groups = append(appState.Groups, state.NewGroup("scene1", 1))

	model := NewModel(appState, tmpDir)
	model.currentScreen = ScreenClassification
	model.files = []string{"file1.mp4", "file2.mp4", "file3.mp4", "file4.mp4", "file5.mp4"}
	model.currentFileIndex = 0
	model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex)
	model.actionsPerSave = 3 // Save every 3 actions

	statePath := state.StateFilePath(tmpDir)

	// Action 1 - should not save yet
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	updated, _ := model.Update(keyMsg)
	model = updated.(Model)
	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Error("expected state file to NOT exist after 1 action (threshold is 3)")
	}

	// Action 2 - should not save yet
	updated, _ = model.Update(keyMsg)
	model = updated.(Model)
	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Error("expected state file to NOT exist after 2 actions (threshold is 3)")
	}

	// Action 3 - should trigger save
	updated, _ = model.Update(keyMsg)
	model = updated.(Model)
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Error("expected state file to exist after 3 actions (threshold reached)")
	}

	// Verify counter was reset and action 4 doesn't save
	os.Remove(statePath) // Remove state file to test counter reset
	updated, _ = model.Update(keyMsg)
	model = updated.(Model)
	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Error("expected state file to NOT exist after counter reset (action 4)")
	}

	// Action 5 - should not save yet
	updated, _ = model.Update(keyMsg)
	model = updated.(Model)
	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Error("expected state file to NOT exist after 5 actions (2 since reset)")
	}

	// Action 6 - should trigger save again
	updated, _ = model.Update(keyMsg)
	model = updated.(Model)
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Error("expected state file to exist after 6 actions (3 since reset)")
	}
}

// TestModel_ScreenTransitionSave_FromClassification tests that state is saved when leaving classification screen
func TestModel_ScreenTransitionSave_FromClassification(t *testing.T) {
	tmpDir := t.TempDir()
	appState := state.NewState(tmpDir, state.SortByModifiedTime)
	appState.Groups = append(appState.Groups, state.NewGroup("scene1", 1))

	model := NewModel(appState, tmpDir)
	model.currentScreen = ScreenClassification
	model.files = []string{"file1.mp4"}
	model.currentFileIndex = 0
	model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex)

	// Transition from classification to review screen
	msg := TransitionToScreen{Screen: ScreenReview}
	_, _ = model.Update(msg)

	// Verify state file was created
	statePath := state.StateFilePath(tmpDir)
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Error("expected state file to be created when leaving classification screen")
	}

	// Verify state can be loaded
	loaded, err := state.Load(statePath)
	if err != nil {
		t.Fatalf("failed to load saved state: %v", err)
	}

	if len(loaded.Groups) != 1 {
		t.Errorf("expected 1 group in loaded state, got %d", len(loaded.Groups))
	}
}

// TestModel_ScreenTransitionSave_ToGroupSelection tests that state is saved when transitioning to group selection
func TestModel_ScreenTransitionSave_ToGroupSelection(t *testing.T) {
	tmpDir := t.TempDir()
	appState := state.NewState(tmpDir, state.SortByModifiedTime)
	appState.Groups = append(appState.Groups, state.NewGroup("scene1", 1))

	model := NewModel(appState, tmpDir)
	model.currentScreen = ScreenClassification
	model.files = []string{"file1.mp4"}
	model.currentFileIndex = 0
	model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex)

	// Trigger transition to group selection with '2' key
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}
	_, _ = model.Update(keyMsg)

	// Verify state file was created
	statePath := state.StateFilePath(tmpDir)
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Error("expected state file to be created when transitioning from classification to group selection")
	}
}

// TestModel_ScreenTransitionSave_ToGroupInsertion tests that state is saved when transitioning to group insertion
func TestModel_ScreenTransitionSave_ToGroupInsertion(t *testing.T) {
	tmpDir := t.TempDir()
	appState := state.NewState(tmpDir, state.SortByModifiedTime)
	appState.Groups = append(appState.Groups, state.NewGroup("scene1", 1))

	model := NewModel(appState, tmpDir)
	model.currentScreen = ScreenClassification
	model.files = []string{"file1.mp4"}
	model.currentFileIndex = 0
	model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex)

	// Trigger transition to group insertion with '3' key
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}}
	_, _ = model.Update(keyMsg)

	// Verify state file was created
	statePath := state.StateFilePath(tmpDir)
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Error("expected state file to be created when transitioning from classification to group insertion")
	}
}

// TestModel_NoSaveOnTransitionWithinNonClassification tests that transitions between non-classification screens don't trigger save
func TestModel_NoSaveOnTransitionWithinNonClassification(t *testing.T) {
	tmpDir := t.TempDir()
	appState := state.NewState(tmpDir, state.SortByModifiedTime)

	model := NewModel(appState, tmpDir)
	model.currentScreen = ScreenReview

	// Transition from review to complete (no classification screen involved)
	msg := TransitionToScreen{Screen: ScreenComplete}
	_, _ = model.Update(msg)

	// Verify state file was NOT created
	statePath := state.StateFilePath(tmpDir)
	if _, err := os.Stat(statePath); !os.IsNotExist(err) {
		t.Error("expected state file to NOT be created for transition between non-classification screens")
	}
}

// TestModel_ActionCounterInitialization tests that action counter is initialized correctly
func TestModel_ActionCounterInitialization(t *testing.T) {
	tmpDir := t.TempDir()
	appState := state.NewState(tmpDir, state.SortByModifiedTime)

	model := NewModel(appState, tmpDir)

	if model.actionCounter != 0 {
		t.Errorf("expected action counter to be 0, got %d", model.actionCounter)
	}

	if model.actionsPerSave != 5 {
		t.Errorf("expected actionsPerSave to be 5 (default), got %d", model.actionsPerSave)
	}
}
