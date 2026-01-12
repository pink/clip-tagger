// main_test.go
package main

import (
	"clip-tagger/state"
	"os"
	"path/filepath"
	"testing"
)

func TestCleanMissingFiles(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.mp4")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create state with classifications for both existing and missing files
	appState := state.NewState(tmpDir, state.SortByModifiedTime)
	appState.Groups = []state.Group{
		{ID: "group1", Name: "Group 1", Order: 1},
	}
	appState.Classifications = []state.Classification{
		{File: "test.mp4", GroupID: "group1", TakeNumber: 1},
		{File: "missing.mp4", GroupID: "group1", TakeNumber: 2},
		{File: "also-missing.mov", GroupID: "group1", TakeNumber: 3},
	}

	// Clean missing files
	cleanedCount := cleanMissingFiles(appState)

	// Should have removed 2 missing files
	if cleanedCount != 2 {
		t.Errorf("expected to clean 2 files, got %d", cleanedCount)
	}

	// Should have 1 classification left
	if len(appState.Classifications) != 1 {
		t.Errorf("expected 1 classification remaining, got %d", len(appState.Classifications))
	}

	// Remaining classification should be for test.mp4
	if appState.Classifications[0].File != "test.mp4" {
		t.Errorf("expected remaining file to be 'test.mp4', got '%s'", appState.Classifications[0].File)
	}
}

func TestCleanMissingFiles_AllExist(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create test files
	testFile1 := filepath.Join(tmpDir, "test1.mp4")
	testFile2 := filepath.Join(tmpDir, "test2.mp4")
	if err := os.WriteFile(testFile1, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testFile2, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create state with classifications for existing files
	appState := state.NewState(tmpDir, state.SortByModifiedTime)
	appState.Groups = []state.Group{
		{ID: "group1", Name: "Group 1", Order: 1},
	}
	appState.Classifications = []state.Classification{
		{File: "test1.mp4", GroupID: "group1", TakeNumber: 1},
		{File: "test2.mp4", GroupID: "group1", TakeNumber: 2},
	}

	// Clean missing files
	cleanedCount := cleanMissingFiles(appState)

	// Should have removed 0 files
	if cleanedCount != 0 {
		t.Errorf("expected to clean 0 files, got %d", cleanedCount)
	}

	// Should have 2 classifications left
	if len(appState.Classifications) != 2 {
		t.Errorf("expected 2 classifications remaining, got %d", len(appState.Classifications))
	}
}

func TestCleanMissingFiles_AllMissing(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create state with classifications for missing files
	appState := state.NewState(tmpDir, state.SortByModifiedTime)
	appState.Groups = []state.Group{
		{ID: "group1", Name: "Group 1", Order: 1},
	}
	appState.Classifications = []state.Classification{
		{File: "missing1.mp4", GroupID: "group1", TakeNumber: 1},
		{File: "missing2.mp4", GroupID: "group1", TakeNumber: 2},
	}

	// Clean missing files
	cleanedCount := cleanMissingFiles(appState)

	// Should have removed 2 files
	if cleanedCount != 2 {
		t.Errorf("expected to clean 2 files, got %d", cleanedCount)
	}

	// Should have 0 classifications left
	if len(appState.Classifications) != 0 {
		t.Errorf("expected 0 classifications remaining, got %d", len(appState.Classifications))
	}
}

func TestCleanMissingFiles_EmptyState(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create empty state
	appState := state.NewState(tmpDir, state.SortByModifiedTime)

	// Clean missing files
	cleanedCount := cleanMissingFiles(appState)

	// Should have removed 0 files
	if cleanedCount != 0 {
		t.Errorf("expected to clean 0 files, got %d", cleanedCount)
	}

	// Should have 0 classifications
	if len(appState.Classifications) != 0 {
		t.Errorf("expected 0 classifications, got %d", len(appState.Classifications))
	}
}

func TestShowPreview_EmptyState(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create empty state
	appState := state.NewState(tmpDir, state.SortByModifiedTime)

	// Should not panic
	showPreview(appState)
}

func TestShowPreview_WithClassifications(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create test files
	testFile1 := filepath.Join(tmpDir, "test1.mp4")
	testFile2 := filepath.Join(tmpDir, "test2.mp4")
	if err := os.WriteFile(testFile1, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testFile2, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create state with classifications
	appState := state.NewState(tmpDir, state.SortByModifiedTime)
	appState.Groups = []state.Group{
		{ID: "group1", Name: "Group 1", Order: 1},
	}
	appState.Classifications = []state.Classification{
		{File: "test1.mp4", GroupID: "group1", TakeNumber: 1},
		{File: "test2.mp4", GroupID: "group1", TakeNumber: 2},
	}

	// Should not panic
	showPreview(appState)
}
