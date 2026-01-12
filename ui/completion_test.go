// ui/completion_test.go
package ui

import (
	"clip-tagger/renamer"
	"clip-tagger/state"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewCompletionData(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create test state
	appState := state.NewState(tmpDir, state.SortByModifiedTime)

	// Add groups
	group1 := state.NewGroup("intro", 1)
	group2 := state.NewGroup("middle", 2)
	appState.Groups = []state.Group{group1, group2}

	// Add classifications
	appState.Classifications = []state.Classification{
		{File: "clip1.mp4", GroupID: group1.ID, TakeNumber: 1},
		{File: "clip2.mp4", GroupID: group1.ID, TakeNumber: 2},
		{File: "clip3.mp4", GroupID: group2.ID, TakeNumber: 1},
	}

	// Create actual files (for conflict detection to work)
	for _, c := range appState.Classifications {
		filePath := filepath.Join(tmpDir, c.File)
		if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	data := NewCompletionData(appState)

	if data == nil {
		t.Fatal("expected non-nil completion data")
	}

	if data.SelectedMode != 0 {
		t.Errorf("expected selected mode to be 0, got %d", data.SelectedMode)
	}

	if len(data.Renames) != 3 {
		t.Errorf("expected 3 rename operations, got %d", len(data.Renames))
	}

	if data.OutputDirectory == "" {
		t.Error("expected non-empty output directory")
	}

	// Verify renames are properly created
	for _, r := range data.Renames {
		if r.OriginalPath == "" {
			t.Error("rename has empty original path")
		}
		if r.TargetPath == "" {
			t.Error("rename has empty target path")
		}
	}
}

func TestNewCompletionDataDetectsConflicts(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create test state
	appState := state.NewState(tmpDir, state.SortByModifiedTime)

	// Add group
	group := state.NewGroup("intro", 1)
	appState.Groups = []state.Group{group}

	// Add classifications
	appState.Classifications = []state.Classification{
		{File: "clip1.mp4", GroupID: group.ID, TakeNumber: 1},
	}

	// Create source file
	srcFile := filepath.Join(tmpDir, "clip1.mp4")
	if err := os.WriteFile(srcFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a file that conflicts with target name
	conflictFile := filepath.Join(tmpDir, "[01_01] intro.mp4")
	if err := os.WriteFile(conflictFile, []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}

	data := NewCompletionData(appState)

	if len(data.Conflicts) != 1 {
		t.Errorf("expected 1 conflict, got %d", len(data.Conflicts))
	}

	if data.HasConflicts != true {
		t.Error("expected HasConflicts to be true")
	}
}

func TestCompletionViewModeSelection(t *testing.T) {
	tmpDir := t.TempDir()
	appState := state.NewState(tmpDir, state.SortByModifiedTime)

	data := NewCompletionData(appState)
	data.SelectedMode = 0
	data.HasConflicts = false

	view := CompletionView(data)

	if !strings.Contains(view, "Mode Selection") {
		t.Error("view should contain 'Mode Selection' header")
	}

	if !strings.Contains(view, "1. Rename in place") {
		t.Error("view should contain 'Rename in place' option")
	}

	if !strings.Contains(view, "2. Copy to new directory") {
		t.Error("view should contain 'Copy to new directory' option")
	}

	// Check for selection indicator on first option
	lines := strings.Split(view, "\n")
	found := false
	for _, line := range lines {
		if strings.Contains(line, "> 1. Rename in place") {
			found = true
			break
		}
	}
	if !found {
		t.Error("view should show selection indicator on first option")
	}
}

func TestCompletionViewShowsConflicts(t *testing.T) {
	tmpDir := t.TempDir()
	appState := state.NewState(tmpDir, state.SortByModifiedTime)

	data := NewCompletionData(appState)
	data.HasConflicts = true
	data.Conflicts = []renamer.Rename{
		{
			OriginalPath: filepath.Join(tmpDir, "clip1.mp4"),
			TargetPath:   filepath.Join(tmpDir, "[01_01] intro.mp4"),
		},
	}

	view := CompletionView(data)

	if !strings.Contains(view, "WARNING") || !strings.Contains(view, "Conflicts") {
		t.Error("view should contain conflict warning")
	}

	if !strings.Contains(view, "clip1.mp4") {
		t.Error("view should show conflicting file")
	}
}

func TestCompletionUpdateNavigateModes(t *testing.T) {
	tmpDir := t.TempDir()
	appState := state.NewState(tmpDir, state.SortByModifiedTime)
	data := NewCompletionData(appState)

	// Test moving down
	result := CompletionUpdate(data, "down")
	if result.Screen != -2 {
		t.Error("down should not change screen")
	}
	if data.SelectedMode != 1 {
		t.Errorf("expected selected mode to be 1, got %d", data.SelectedMode)
	}

	// Test moving down at bottom (should not wrap)
	result = CompletionUpdate(data, "down")
	if data.SelectedMode != 1 {
		t.Errorf("should stay at mode 1, got %d", data.SelectedMode)
	}

	// Test moving up
	result = CompletionUpdate(data, "up")
	if result.Screen != -2 {
		t.Error("up should not change screen")
	}
	if data.SelectedMode != 0 {
		t.Errorf("expected selected mode to be 0, got %d", data.SelectedMode)
	}

	// Test moving up at top (should not wrap)
	result = CompletionUpdate(data, "up")
	if data.SelectedMode != 0 {
		t.Errorf("should stay at mode 0, got %d", data.SelectedMode)
	}
}

func TestCompletionUpdateExecuteRenameInPlace(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create test state
	appState := state.NewState(tmpDir, state.SortByModifiedTime)

	// Add group
	group := state.NewGroup("intro", 1)
	appState.Groups = []state.Group{group}

	// Add classifications
	appState.Classifications = []state.Classification{
		{File: "clip1.mp4", GroupID: group.ID, TakeNumber: 1},
	}

	// Create source file
	srcFile := filepath.Join(tmpDir, "clip1.mp4")
	if err := os.WriteFile(srcFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	data := NewCompletionData(appState)
	data.SelectedMode = 0 // Rename in place

	result := CompletionUpdate(data, "enter")

	if result.Screen != -2 {
		t.Error("should not change screen immediately")
	}

	if data.ExecutionResult == nil {
		t.Fatal("expected execution result to be set")
	}

	if data.ExecutionResult.Error != nil {
		t.Errorf("expected no error, got %v", data.ExecutionResult.Error)
	}

	// Check file was renamed
	targetPath := filepath.Join(tmpDir, "[01_01] intro.mp4")
	if _, err := os.Stat(targetPath); err != nil {
		t.Error("target file should exist")
	}

	if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
		t.Error("source file should not exist")
	}
}

func TestCompletionUpdateExecuteCopyToDirectory(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create test state
	appState := state.NewState(tmpDir, state.SortByModifiedTime)

	// Add group
	group := state.NewGroup("intro", 1)
	appState.Groups = []state.Group{group}

	// Add classifications
	appState.Classifications = []state.Classification{
		{File: "clip1.mp4", GroupID: group.ID, TakeNumber: 1},
	}

	// Create source file
	srcFile := filepath.Join(tmpDir, "clip1.mp4")
	if err := os.WriteFile(srcFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	data := NewCompletionData(appState)
	data.SelectedMode = 1 // Copy to directory

	result := CompletionUpdate(data, "enter")

	if result.Screen != -2 {
		t.Error("should not change screen immediately")
	}

	if data.ExecutionResult == nil {
		t.Fatal("expected execution result to be set")
	}

	if data.ExecutionResult.Error != nil {
		t.Errorf("expected no error, got %v", data.ExecutionResult.Error)
	}

	// Check file was copied to output directory
	targetPath := filepath.Join(data.OutputDirectory, "[01_01] intro.mp4")
	if _, err := os.Stat(targetPath); err != nil {
		t.Errorf("target file should exist at %s", targetPath)
	}

	// Source file should still exist
	if _, err := os.Stat(srcFile); err != nil {
		t.Error("source file should still exist")
	}
}

func TestCompletionUpdateAbortOnConflict(t *testing.T) {
	tmpDir := t.TempDir()
	appState := state.NewState(tmpDir, state.SortByModifiedTime)

	// Add group
	group := state.NewGroup("intro", 1)
	appState.Groups = []state.Group{group}

	// Add classifications
	appState.Classifications = []state.Classification{
		{File: "clip1.mp4", GroupID: group.ID, TakeNumber: 1},
	}

	// Create source file
	srcFile := filepath.Join(tmpDir, "clip1.mp4")
	if err := os.WriteFile(srcFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create conflict file
	conflictFile := filepath.Join(tmpDir, "[01_01] intro.mp4")
	if err := os.WriteFile(conflictFile, []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}

	data := NewCompletionData(appState)
	data.SelectedMode = 0

	// Should show conflict
	if !data.HasConflicts {
		t.Fatal("expected conflicts to be detected")
	}

	// Press Esc to abort
	result := CompletionUpdate(data, "esc")

	if result.Screen != ScreenReview {
		t.Errorf("expected to go back to review screen, got screen %d", result.Screen)
	}
}

func TestCompletionUpdateQuit(t *testing.T) {
	tmpDir := t.TempDir()
	appState := state.NewState(tmpDir, state.SortByModifiedTime)
	data := NewCompletionData(appState)

	result := CompletionUpdate(data, "q")
	if result.Screen != -1 {
		t.Errorf("expected quit screen (-1), got %d", result.Screen)
	}

	result = CompletionUpdate(data, "ctrl+c")
	if result.Screen != -1 {
		t.Errorf("expected quit screen (-1), got %d", result.Screen)
	}
}

func TestCompletionUpdateAfterExecution(t *testing.T) {
	tmpDir := t.TempDir()
	appState := state.NewState(tmpDir, state.SortByModifiedTime)

	// Add group
	group := state.NewGroup("intro", 1)
	appState.Groups = []state.Group{group}

	// Add classifications
	appState.Classifications = []state.Classification{
		{File: "clip1.mp4", GroupID: group.ID, TakeNumber: 1},
	}

	// Create source file
	srcFile := filepath.Join(tmpDir, "clip1.mp4")
	if err := os.WriteFile(srcFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	data := NewCompletionData(appState)
	data.SelectedMode = 0

	// Execute rename
	CompletionUpdate(data, "enter")

	if data.ExecutionResult == nil {
		t.Fatal("expected execution result")
	}

	// Now pressing any key should quit
	result := CompletionUpdate(data, "enter")
	if result.Screen != -1 {
		t.Errorf("expected quit after execution, got screen %d", result.Screen)
	}
}

func TestCompletionViewAfterSuccessfulExecution(t *testing.T) {
	tmpDir := t.TempDir()
	appState := state.NewState(tmpDir, state.SortByModifiedTime)

	data := NewCompletionData(appState)
	data.ExecutionResult = &CompletionExecutionResult{
		Success:      true,
		FilesChanged: 5,
		Mode:         "Rename in place",
		Error:        nil,
	}

	view := CompletionView(data)

	if !strings.Contains(view, "Success") {
		t.Error("view should contain 'Success' message")
	}

	if !strings.Contains(view, "5") {
		t.Error("view should show number of files changed")
	}

	if !strings.Contains(view, "Rename in place") {
		t.Error("view should show the mode used")
	}
}

func TestCompletionViewAfterFailedExecution(t *testing.T) {
	tmpDir := t.TempDir()
	appState := state.NewState(tmpDir, state.SortByModifiedTime)

	data := NewCompletionData(appState)
	data.ExecutionResult = &CompletionExecutionResult{
		Success:      false,
		FilesChanged: 0,
		Mode:         "Rename in place",
		Error:        os.ErrPermission,
	}

	view := CompletionView(data)

	if !strings.Contains(view, "Error") || !strings.Contains(view, "Failed") {
		t.Error("view should contain error message")
	}

	if !strings.Contains(view, "permission") {
		t.Error("view should show the error details")
	}
}
