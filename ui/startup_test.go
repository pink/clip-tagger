// ui/startup_test.go
package ui

import (
	"clip-tagger/state"
	"testing"
)

func TestStartupData_NewSession(t *testing.T) {
	// Test new session with no existing state
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	files := []string{"file1.mp4", "file2.mov", "file3.avi"}

	data := NewStartupData(appState, files, nil)

	if data.IsResume {
		t.Error("expected IsResume to be false for new session")
	}
	if data.ClassifiedCount != 0 {
		t.Errorf("expected ClassifiedCount to be 0, got %d", data.ClassifiedCount)
	}
	if data.RemainingCount != 3 {
		t.Errorf("expected RemainingCount to be 3, got %d", data.RemainingCount)
	}
	if data.TotalFiles != 3 {
		t.Errorf("expected TotalFiles to be 3, got %d", data.TotalFiles)
	}
	if data.NewFilesCount != 0 {
		t.Error("expected NewFilesCount to be 0 for new session")
	}
	if data.MissingFilesCount != 0 {
		t.Error("expected MissingFilesCount to be 0 for new session")
	}
}

func TestStartupData_ResumeSession(t *testing.T) {
	// Test resuming with existing classifications
	appState := state.NewState("/test/dir", state.SortByModifiedTime)

	// Add some groups and classifications
	group1 := state.NewGroup("Scene 1", 1)
	appState.Groups = append(appState.Groups, group1)
	appState.AddOrUpdateClassification("file1.mp4", group1.ID)
	appState.AddOrUpdateClassification("file2.mov", group1.ID)

	files := []string{"file1.mp4", "file2.mov", "file3.avi"}

	data := NewStartupData(appState, files, nil)

	if !data.IsResume {
		t.Error("expected IsResume to be true when state has classifications")
	}
	if data.ClassifiedCount != 2 {
		t.Errorf("expected ClassifiedCount to be 2, got %d", data.ClassifiedCount)
	}
	if data.RemainingCount != 1 {
		t.Errorf("expected RemainingCount to be 1, got %d", data.RemainingCount)
	}
	if data.TotalFiles != 3 {
		t.Errorf("expected TotalFiles to be 3, got %d", data.TotalFiles)
	}
}

func TestStartupData_WithMergeResult(t *testing.T) {
	// Test with new and missing files
	appState := state.NewState("/test/dir", state.SortByModifiedTime)

	group1 := state.NewGroup("Scene 1", 1)
	appState.Groups = append(appState.Groups, group1)
	appState.AddOrUpdateClassification("file1.mp4", group1.ID)

	files := []string{"file1.mp4", "file2.mov", "file3.avi"}

	mergeResult := &state.MergeResult{
		NewFiles:      []string{"file2.mov", "file3.avi"},
		MissingFiles:  []string{"file0.mp4"},
		ExistingCount: 1,
	}

	data := NewStartupData(appState, files, mergeResult)

	if !data.IsResume {
		t.Error("expected IsResume to be true")
	}
	if data.NewFilesCount != 2 {
		t.Errorf("expected NewFilesCount to be 2, got %d", data.NewFilesCount)
	}
	if data.MissingFilesCount != 1 {
		t.Errorf("expected MissingFilesCount to be 1, got %d", data.MissingFilesCount)
	}
	if data.ClassifiedCount != 1 {
		t.Errorf("expected ClassifiedCount to be 1, got %d", data.ClassifiedCount)
	}
}

func TestStartupView_NewSession(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	files := []string{"file1.mp4", "file2.mov"}
	data := NewStartupData(appState, files, nil)

	view := StartupView(data)

	// Check for expected content
	if view == "" {
		t.Fatal("expected non-empty view")
	}

	// Should indicate new session
	if !contains(view, "New session") {
		t.Error("expected view to contain 'New session'")
	}

	// Should show file count
	if !contains(view, "2 files") {
		t.Error("expected view to show file count")
	}

	// Should show sorting info
	if !contains(view, "modified_time") || !contains(view, "Sorted by") {
		t.Error("expected view to show sorting information")
	}

	// Should show instructions
	if !contains(view, "Press Enter") {
		t.Error("expected view to show Enter key instruction")
	}
	if !contains(view, "Press 'q' or Ctrl+C to quit") {
		t.Error("expected view to show quit instructions")
	}
}

func TestStartupView_ResumeSession(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByName)

	group1 := state.NewGroup("Scene 1", 1)
	appState.Groups = append(appState.Groups, group1)
	appState.AddOrUpdateClassification("file1.mp4", group1.ID)

	files := []string{"file1.mp4", "file2.mov", "file3.avi"}

	mergeResult := &state.MergeResult{
		NewFiles:      []string{"file2.mov", "file3.avi"},
		MissingFiles:  []string{},
		ExistingCount: 1,
	}

	data := NewStartupData(appState, files, mergeResult)
	view := StartupView(data)

	// Should indicate resuming
	if !contains(view, "Found existing session") {
		t.Error("expected view to contain 'Found existing session'")
	}

	// Should show classified and remaining counts
	if !contains(view, "1 classified") {
		t.Error("expected view to show classified count")
	}
	if !contains(view, "2 remaining") {
		t.Error("expected view to show remaining count")
	}

	// Should show new files count
	if !contains(view, "2 new files") {
		t.Error("expected view to show new files count")
	}

	// Should show continue instruction
	if !contains(view, "Press Enter") {
		t.Error("expected view to show Enter key instruction")
	}
}

func TestStartupView_WithMissingFiles(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByName)

	group1 := state.NewGroup("Scene 1", 1)
	appState.Groups = append(appState.Groups, group1)
	appState.AddOrUpdateClassification("file1.mp4", group1.ID)

	files := []string{"file1.mp4", "file2.mov"}

	mergeResult := &state.MergeResult{
		NewFiles:      []string{"file2.mov"},
		MissingFiles:  []string{"file0.mp4"},
		ExistingCount: 1,
	}

	data := NewStartupData(appState, files, mergeResult)
	view := StartupView(data)

	// Should show missing files warning
	if !contains(view, "1 missing file") {
		t.Error("expected view to show missing files count")
	}
}

func TestStartupUpdate_EnterKey(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	files := []string{"file1.mp4"}
	data := NewStartupData(appState, files, nil)

	// Simulate Enter key press
	msg := "enter"
	transition := StartupUpdate(data, msg)

	if transition != ScreenClassification {
		t.Errorf("expected transition to ScreenClassification, got %v", transition)
	}
}

func TestStartupUpdate_QuitKeys(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	files := []string{"file1.mp4"}
	data := NewStartupData(appState, files, nil)

	t.Run("q key should quit", func(t *testing.T) {
		msg := "q"
		transition := StartupUpdate(data, msg)

		if transition != -1 {
			t.Errorf("expected quit signal (-1), got %v", transition)
		}
	})

	t.Run("ctrl+c should quit", func(t *testing.T) {
		msg := "ctrl+c"
		transition := StartupUpdate(data, msg)

		if transition != -1 {
			t.Errorf("expected quit signal (-1), got %v", transition)
		}
	})
}

func TestStartupUpdate_OtherKeys(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	files := []string{"file1.mp4"}
	data := NewStartupData(appState, files, nil)

	// Other keys should do nothing (return -2 for no action)
	msg := "a"
	transition := StartupUpdate(data, msg)

	if transition != -2 {
		t.Errorf("expected no action (-2), got %v", transition)
	}
}

// Helper function for string containment check
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) && hasSubstring(s, substr))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
