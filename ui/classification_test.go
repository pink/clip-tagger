// ui/classification_test.go
package ui

import (
	"clip-tagger/state"
	"testing"
)

func TestClassificationData_NewSession(t *testing.T) {
	// Test classification screen with no existing classifications
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	files := []string{"file1.mp4", "file2.mov", "file3.avi"}

	data := NewClassificationData(appState, files, 0)

	if data.CurrentFile != "file1.mp4" {
		t.Errorf("expected CurrentFile to be 'file1.mp4', got '%s'", data.CurrentFile)
	}
	if data.CurrentIndex != 1 {
		t.Errorf("expected CurrentIndex to be 1, got %d", data.CurrentIndex)
	}
	if data.TotalFiles != 3 {
		t.Errorf("expected TotalFiles to be 3, got %d", data.TotalFiles)
	}
	if data.FilePath != "/test/dir/file1.mp4" {
		t.Errorf("expected FilePath to be '/test/dir/file1.mp4', got '%s'", data.FilePath)
	}
	if data.HasPreviousClassification {
		t.Error("expected HasPreviousClassification to be false for new session")
	}
	if data.PreviousGroupName != "" {
		t.Errorf("expected PreviousGroupName to be empty, got '%s'", data.PreviousGroupName)
	}
}

func TestClassificationData_WithPreviousClassification(t *testing.T) {
	// Test classification screen with a previous classification
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	group1 := state.NewGroup("Scene 1", 1)
	appState.Groups = append(appState.Groups, group1)
	appState.AddOrUpdateClassification("file1.mp4", group1.ID)

	files := []string{"file1.mp4", "file2.mov", "file3.avi"}

	data := NewClassificationData(appState, files, 1)

	if data.CurrentFile != "file2.mov" {
		t.Errorf("expected CurrentFile to be 'file2.mov', got '%s'", data.CurrentFile)
	}
	if data.CurrentIndex != 2 {
		t.Errorf("expected CurrentIndex to be 2, got %d", data.CurrentIndex)
	}
	if !data.HasPreviousClassification {
		t.Error("expected HasPreviousClassification to be true")
	}
	if data.PreviousGroupName != "Scene 1" {
		t.Errorf("expected PreviousGroupName to be 'Scene 1', got '%s'", data.PreviousGroupName)
	}
}

func TestClassificationData_FirstFileInResume(t *testing.T) {
	// Test when resuming at the first unclassified file
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	group1 := state.NewGroup("Scene 1", 1)
	appState.Groups = append(appState.Groups, group1)
	// No classifications yet in this case

	files := []string{"file1.mp4", "file2.mov"}

	data := NewClassificationData(appState, files, 0)

	if data.HasPreviousClassification {
		t.Error("expected HasPreviousClassification to be false when there are no classifications")
	}
}

func TestClassificationView_NewSession(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	files := []string{"file1.mp4", "file2.mov"}
	data := NewClassificationData(appState, files, 0)

	view := ClassificationView(data)

	if view == "" {
		t.Fatal("expected non-empty view")
	}

	// Check for file info
	if !contains(view, "file1.mp4") {
		t.Error("expected view to contain current filename")
	}
	if !contains(view, "File 1 of 2") {
		t.Error("expected view to show progress indicator")
	}
	if !contains(view, "/test/dir/file1.mp4") {
		t.Error("expected view to show full file path")
	}

	// Check for actions
	if !contains(view, "'p' - Preview file") {
		t.Error("expected view to show preview action")
	}
	if !contains(view, "'2' - Select from existing groups") {
		t.Error("expected view to show group selection action")
	}
	if !contains(view, "'3' - Create new group") {
		t.Error("expected view to show group creation action")
	}
	if !contains(view, "'s' - Skip this file") {
		t.Error("expected view to show skip action")
	}
	if !contains(view, "'q' - Quit") {
		t.Error("expected view to show quit action")
	}

	// Should NOT show "Same as last" option for first file
	if contains(view, "'1' - Same as last") {
		t.Error("expected view NOT to show 'Same as last' option for first file")
	}
}

func TestClassificationView_WithPreviousClassification(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	group1 := state.NewGroup("Scene 1", 1)
	appState.Groups = append(appState.Groups, group1)
	appState.AddOrUpdateClassification("file1.mp4", group1.ID)

	files := []string{"file1.mp4", "file2.mov"}
	data := NewClassificationData(appState, files, 1)

	view := ClassificationView(data)

	// Should show "Same as last" option with group name
	if !contains(view, "'1' - Same as last") {
		t.Error("expected view to show 'Same as last' option")
	}
	if !contains(view, "Scene 1") {
		t.Error("expected view to show previous group name")
	}
}

func TestClassificationView_Progress(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	files := []string{"file1.mp4", "file2.mov", "file3.avi", "file4.mp4"}

	t.Run("first file", func(t *testing.T) {
		data := NewClassificationData(appState, files, 0)
		view := ClassificationView(data)
		if !contains(view, "File 1 of 4") {
			t.Error("expected view to show 'File 1 of 4'")
		}
	})

	t.Run("middle file", func(t *testing.T) {
		data := NewClassificationData(appState, files, 2)
		view := ClassificationView(data)
		if !contains(view, "File 3 of 4") {
			t.Error("expected view to show 'File 3 of 4'")
		}
	})

	t.Run("last file", func(t *testing.T) {
		data := NewClassificationData(appState, files, 3)
		view := ClassificationView(data)
		if !contains(view, "File 4 of 4") {
			t.Error("expected view to show 'File 4 of 4'")
		}
	})
}

func TestClassificationUpdate_PreviewKey(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	files := []string{"file1.mp4"}
	data := NewClassificationData(appState, files, 0)

	result := ClassificationUpdate(data, "p")

	if result.Action != ClassificationActionPreview {
		t.Errorf("expected action to be ClassificationActionPreview, got %v", result.Action)
	}
	if result.Screen != -2 {
		t.Errorf("expected no screen transition, got %v", result.Screen)
	}
}

func TestClassificationUpdate_SameAsLastKey(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	group1 := state.NewGroup("Scene 1", 1)
	appState.Groups = append(appState.Groups, group1)
	appState.AddOrUpdateClassification("file1.mp4", group1.ID)

	files := []string{"file1.mp4", "file2.mov"}
	data := NewClassificationData(appState, files, 1)

	result := ClassificationUpdate(data, "1")

	if result.Action != ClassificationActionSameAsLast {
		t.Errorf("expected action to be ClassificationActionSameAsLast, got %v", result.Action)
	}
	if result.Screen != -2 {
		t.Errorf("expected no screen transition, got %v", result.Screen)
	}
}

func TestClassificationUpdate_SameAsLastKey_NotAvailable(t *testing.T) {
	// Test that '1' does nothing when there's no previous classification
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	files := []string{"file1.mp4"}
	data := NewClassificationData(appState, files, 0)

	result := ClassificationUpdate(data, "1")

	if result.Action != ClassificationActionNone {
		t.Errorf("expected action to be ClassificationActionNone, got %v", result.Action)
	}
	if result.Screen != -2 {
		t.Errorf("expected no screen transition, got %v", result.Screen)
	}
}

func TestClassificationUpdate_SelectGroupKey(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	files := []string{"file1.mp4"}
	data := NewClassificationData(appState, files, 0)

	result := ClassificationUpdate(data, "2")

	if result.Action != ClassificationActionSelectGroup {
		t.Errorf("expected action to be ClassificationActionSelectGroup, got %v", result.Action)
	}
	if result.Screen != ScreenGroupSelection {
		t.Errorf("expected transition to ScreenGroupSelection, got %v", result.Screen)
	}
}

func TestClassificationUpdate_CreateGroupKey(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	files := []string{"file1.mp4"}
	data := NewClassificationData(appState, files, 0)

	result := ClassificationUpdate(data, "3")

	if result.Action != ClassificationActionCreateGroup {
		t.Errorf("expected action to be ClassificationActionCreateGroup, got %v", result.Action)
	}
	if result.Screen != ScreenGroupInsertion {
		t.Errorf("expected transition to ScreenGroupInsertion, got %v", result.Screen)
	}
}

func TestClassificationUpdate_SkipKey(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	files := []string{"file1.mp4"}
	data := NewClassificationData(appState, files, 0)

	result := ClassificationUpdate(data, "s")

	if result.Action != ClassificationActionSkip {
		t.Errorf("expected action to be ClassificationActionSkip, got %v", result.Action)
	}
	if result.Screen != -2 {
		t.Errorf("expected no screen transition, got %v", result.Screen)
	}
}

func TestClassificationUpdate_QuitKey(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	files := []string{"file1.mp4"}
	data := NewClassificationData(appState, files, 0)

	t.Run("q key should quit", func(t *testing.T) {
		result := ClassificationUpdate(data, "q")

		if result.Action != ClassificationActionNone {
			t.Errorf("expected action to be ClassificationActionNone for quit, got %v", result.Action)
		}
		if result.Screen != -1 {
			t.Errorf("expected quit signal (-1), got %v", result.Screen)
		}
	})

	t.Run("ctrl+c should quit", func(t *testing.T) {
		result := ClassificationUpdate(data, "ctrl+c")

		if result.Action != ClassificationActionNone {
			t.Errorf("expected action to be ClassificationActionNone for quit, got %v", result.Action)
		}
		if result.Screen != -1 {
			t.Errorf("expected quit signal (-1), got %v", result.Screen)
		}
	})
}

func TestClassificationUpdate_OtherKeys(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	files := []string{"file1.mp4"}
	data := NewClassificationData(appState, files, 0)

	// Other keys should do nothing
	result := ClassificationUpdate(data, "x")

	if result.Action != ClassificationActionNone {
		t.Errorf("expected action to be ClassificationActionNone, got %v", result.Action)
	}
	if result.Screen != -2 {
		t.Errorf("expected no action (-2), got %v", result.Screen)
	}
}
