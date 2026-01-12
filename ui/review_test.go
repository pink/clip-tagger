// ui/review_test.go
package ui

import (
	"clip-tagger/state"
	"path/filepath"
	"strings"
	"testing"
)

func TestReviewData_Creation(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)

	// Create groups
	group1 := state.NewGroup("Scene 1", 1)
	group2 := state.NewGroup("Scene 2", 2)
	appState.Groups = []state.Group{group1, group2}

	// Add classifications
	appState.AddOrUpdateClassification("file1.mp4", group1.ID)
	appState.AddOrUpdateClassification("file2.mp4", group1.ID)
	appState.AddOrUpdateClassification("file3.mp4", group2.ID)

	// Add skipped files
	appState.Skipped = []string{"file4.mp4"}

	files := []string{"file1.mp4", "file2.mp4", "file3.mp4", "file4.mp4"}
	data := NewReviewData(appState, files)

	if data.ClassifiedCount != 3 {
		t.Errorf("expected ClassifiedCount to be 3, got %d", data.ClassifiedCount)
	}
	if data.SkippedCount != 1 {
		t.Errorf("expected SkippedCount to be 1, got %d", data.SkippedCount)
	}
	if len(data.RenameItems) != 4 {
		t.Errorf("expected 4 rename items (3 classified + 1 skipped), got %d", len(data.RenameItems))
	}
	if data.SelectedIndex != 0 {
		t.Errorf("expected SelectedIndex to be 0, got %d", data.SelectedIndex)
	}
	if data.ScrollOffset != 0 {
		t.Errorf("expected ScrollOffset to be 0, got %d", data.ScrollOffset)
	}
}

func TestReviewData_RenameItems(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)

	// Create group
	group := state.NewGroup("Scene 1", 1)
	appState.Groups = []state.Group{group}

	// Add classification
	appState.AddOrUpdateClassification("video.mp4", group.ID)

	files := []string{"video.mp4"}
	data := NewReviewData(appState, files)

	if len(data.RenameItems) != 1 {
		t.Fatalf("expected 1 rename item, got %d", len(data.RenameItems))
	}

	item := data.RenameItems[0]
	if item.OriginalName != "video.mp4" {
		t.Errorf("expected OriginalName to be 'video.mp4', got '%s'", item.OriginalName)
	}
	if item.NewName != "[01_01] Scene 1.mp4" {
		t.Errorf("expected NewName to be '[01_01] Scene 1.mp4', got '%s'", item.NewName)
	}
	if item.IsSkipped {
		t.Errorf("expected IsSkipped to be false")
	}
	if item.ChangeType != "new" {
		t.Errorf("expected ChangeType 'new' for newly classified file, got '%s'", item.ChangeType)
	}
}

func TestReviewData_SkippedFiles(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)

	// Add skipped file
	appState.Skipped = []string{"skipped.mp4"}

	files := []string{"skipped.mp4"}
	data := NewReviewData(appState, files)

	if len(data.RenameItems) != 1 {
		t.Fatalf("expected 1 rename item, got %d", len(data.RenameItems))
	}

	item := data.RenameItems[0]
	if item.OriginalName != "skipped.mp4" {
		t.Errorf("expected OriginalName to be 'skipped.mp4', got '%s'", item.OriginalName)
	}
	if item.NewName != "skipped.mp4" {
		t.Errorf("expected NewName to be 'skipped.mp4' (unchanged), got '%s'", item.NewName)
	}
	if !item.IsSkipped {
		t.Errorf("expected IsSkipped to be true")
	}
	if item.ChangeType != "" {
		t.Errorf("expected empty ChangeType for skipped file, got '%s'", item.ChangeType)
	}
}

func TestReviewData_ChangeTypes(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)

	// Create groups
	group1 := state.NewGroup("Scene 1", 1)
	group2 := state.NewGroup("Scene 2", 2)
	appState.Groups = []state.Group{group1, group2}

	// Simulate files that have been renamed before
	// We need to create files with the old naming pattern
	files := []string{
		"[01_01] Scene 1.mp4", // Already has correct name (no change)
		"[01_02] Scene 1.mp4", // Position changed in Scene 1
		"[02_01] Scene 1.mp4", // Moved from Scene 2 to Scene 1
		"new_file.mp4",        // New file being classified
	}

	// Classify files
	appState.AddOrUpdateClassification("[01_01] Scene 1.mp4", group1.ID) // Take 1
	appState.AddOrUpdateClassification("[01_02] Scene 1.mp4", group1.ID) // Take 2
	appState.AddOrUpdateClassification("[02_01] Scene 1.mp4", group1.ID) // Take 3 (moved from group 2)
	appState.AddOrUpdateClassification("new_file.mp4", group2.ID)        // Take 1 in group 2

	data := NewReviewData(appState, files)

	if len(data.RenameItems) != 4 {
		t.Fatalf("expected 4 rename items, got %d", len(data.RenameItems))
	}

	// Test change detection
	// File 1: [01_01] Scene 1.mp4 -> [01_01] Scene 1.mp4 (no change)
	if data.RenameItems[0].ChangeType != "" {
		t.Errorf("file 1: expected no change type, got '%s'", data.RenameItems[0].ChangeType)
	}

	// File 2: [01_02] Scene 1.mp4 -> [01_02] Scene 1.mp4 (no change)
	if data.RenameItems[1].ChangeType != "" {
		t.Errorf("file 2: expected no change type, got '%s'", data.RenameItems[1].ChangeType)
	}

	// File 3: [02_01] Scene 1.mp4 -> [01_03] Scene 1.mp4 (moved)
	if data.RenameItems[2].ChangeType != "moved" {
		t.Errorf("file 3: expected 'moved' change type, got '%s'", data.RenameItems[2].ChangeType)
	}

	// File 4: new_file.mp4 -> [02_01] Scene 2.mp4 (new)
	if data.RenameItems[3].ChangeType != "new" {
		t.Errorf("file 4: expected 'new' change type, got '%s'", data.RenameItems[3].ChangeType)
	}
}

func TestReviewView_DisplaysCorrectly(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)

	// Create group
	group := state.NewGroup("Scene 1", 1)
	appState.Groups = []state.Group{group}

	// Add classification and skip
	appState.AddOrUpdateClassification("file1.mp4", group.ID)
	appState.Skipped = []string{"file2.mp4"}

	files := []string{"file1.mp4", "file2.mp4"}
	data := NewReviewData(appState, files)

	view := ReviewView(data)

	// Check for key elements in the view
	if !strings.Contains(view, "=== Review Changes ===") {
		t.Error("view should contain header")
	}
	if !strings.Contains(view, "1 file classified, 1 skipped") {
		t.Errorf("view should contain summary, got: %s", view)
	}
	if !strings.Contains(view, "file1.mp4") {
		t.Error("view should contain original filename")
	}
	if !strings.Contains(view, "[01_01] Scene 1.mp4") {
		t.Error("view should contain new filename")
	}
	if !strings.Contains(view, "file2.mp4") {
		t.Error("view should contain skipped file")
	}
	if !strings.Contains(view, "[SKIPPED]") {
		t.Error("view should mark skipped files")
	}
	if !strings.Contains(view, "Enter") {
		t.Error("view should show Enter key option")
	}
	if !strings.Contains(view, "Esc") {
		t.Error("view should show Esc key option")
	}
}

func TestReviewView_ShowsChangeTags(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)

	// Create groups
	group1 := state.NewGroup("Scene 1", 1)
	appState.Groups = []state.Group{group1}

	// Classify a file that was previously in different position
	appState.AddOrUpdateClassification("[02_01] Scene 1.mp4", group1.ID)

	files := []string{"[02_01] Scene 1.mp4"}
	data := NewReviewData(appState, files)

	// The file should show as moved since it's going from [02_01] to [01_01]
	view := ReviewView(data)

	if !strings.Contains(view, "[moved]") {
		t.Error("view should show [moved] tag for position changes")
	}
}

func TestReviewUpdate_NavigationKeys(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	group := state.NewGroup("Scene 1", 1)
	appState.Groups = []state.Group{group}

	// Add multiple files
	files := []string{"file1.mp4", "file2.mp4", "file3.mp4"}
	for _, f := range files {
		appState.AddOrUpdateClassification(f, group.ID)
	}

	data := NewReviewData(appState, files)

	// Test down navigation
	result := ReviewUpdate(data, "down")
	if result.Screen != -2 {
		t.Error("down key should not change screen")
	}
	if data.SelectedIndex != 1 {
		t.Errorf("expected SelectedIndex to be 1 after down, got %d", data.SelectedIndex)
	}

	// Test up navigation
	result = ReviewUpdate(data, "up")
	if result.Screen != -2 {
		t.Error("up key should not change screen")
	}
	if data.SelectedIndex != 0 {
		t.Errorf("expected SelectedIndex to be 0 after up, got %d", data.SelectedIndex)
	}

	// Test up at top (should not go below 0)
	result = ReviewUpdate(data, "up")
	if data.SelectedIndex != 0 {
		t.Errorf("expected SelectedIndex to stay at 0, got %d", data.SelectedIndex)
	}
}

func TestReviewUpdate_EnterKey(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	group := state.NewGroup("Scene 1", 1)
	appState.Groups = []state.Group{group}
	appState.AddOrUpdateClassification("file1.mp4", group.ID)

	files := []string{"file1.mp4"}
	data := NewReviewData(appState, files)

	result := ReviewUpdate(data, "enter")
	if result.Screen != ScreenComplete {
		t.Errorf("expected Enter to transition to ScreenComplete (rename confirmation), got screen %d", result.Screen)
	}
}

func TestReviewUpdate_EscKey(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	group := state.NewGroup("Scene 1", 1)
	appState.Groups = []state.Group{group}
	appState.AddOrUpdateClassification("file1.mp4", group.ID)

	files := []string{"file1.mp4"}
	data := NewReviewData(appState, files)

	result := ReviewUpdate(data, "esc")
	if result.Screen != ScreenClassification {
		t.Errorf("expected Esc to transition to ScreenClassification, got screen %d", result.Screen)
	}
}

func TestReviewUpdate_QuitKeys(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	files := []string{}
	data := NewReviewData(appState, files)

	t.Run("q key quits", func(t *testing.T) {
		result := ReviewUpdate(data, "q")
		if result.Screen != -1 {
			t.Errorf("expected q to quit (screen -1), got %d", result.Screen)
		}
	})

	t.Run("ctrl+c quits", func(t *testing.T) {
		result := ReviewUpdate(data, "ctrl+c")
		if result.Screen != -1 {
			t.Errorf("expected ctrl+c to quit (screen -1), got %d", result.Screen)
		}
	})
}

func TestDetectChangeType(t *testing.T) {
	tests := []struct {
		name         string
		originalPath string
		newPath      string
		expected     string
	}{
		{
			name:         "new file without pattern",
			originalPath: "/dir/video.mp4",
			newPath:      "/dir/[01_01] Scene 1.mp4",
			expected:     "new",
		},
		{
			name:         "no change - same name",
			originalPath: "/dir/[01_01] Scene 1.mp4",
			newPath:      "/dir/[01_01] Scene 1.mp4",
			expected:     "",
		},
		{
			name:         "moved - different group",
			originalPath: "/dir/[02_01] Scene 1.mp4",
			newPath:      "/dir/[01_01] Scene 1.mp4",
			expected:     "moved",
		},
		{
			name:         "updated - different take",
			originalPath: "/dir/[01_01] Scene 1.mp4",
			newPath:      "/dir/[01_02] Scene 1.mp4",
			expected:     "updated",
		},
		{
			name:         "moved - group and take changed",
			originalPath: "/dir/[02_03] Scene 1.mp4",
			newPath:      "/dir/[01_01] Scene 1.mp4",
			expected:     "moved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectChangeType(tt.originalPath, tt.newPath)
			if result != tt.expected {
				t.Errorf("detectChangeType(%s, %s) = %s, want %s",
					filepath.Base(tt.originalPath), filepath.Base(tt.newPath),
					result, tt.expected)
			}
		})
	}
}

func TestReviewData_WithEmptyState(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	files := []string{"file1.mp4", "file2.mp4"}

	// No classifications, no skips
	data := NewReviewData(appState, files)

	if data.ClassifiedCount != 0 {
		t.Errorf("expected ClassifiedCount to be 0, got %d", data.ClassifiedCount)
	}
	if data.SkippedCount != 0 {
		t.Errorf("expected SkippedCount to be 0, got %d", data.SkippedCount)
	}
	// Files without classification or skip are not shown in review
	if len(data.RenameItems) != 0 {
		t.Errorf("expected 0 rename items for unclassified files, got %d", len(data.RenameItems))
	}
}

func TestReviewData_ScrollWindow(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	group := state.NewGroup("Scene 1", 1)
	appState.Groups = []state.Group{group}

	// Create many files to test scrolling
	files := make([]string, 20)
	for i := 0; i < 20; i++ {
		filename := filepath.Join("file", string(rune('A'+i))+".mp4")
		files[i] = filename
		appState.AddOrUpdateClassification(filename, group.ID)
	}

	data := NewReviewData(appState, files)

	// Initially at top
	if data.SelectedIndex != 0 {
		t.Errorf("expected SelectedIndex to start at 0, got %d", data.SelectedIndex)
	}
	if data.ScrollOffset != 0 {
		t.Errorf("expected ScrollOffset to start at 0, got %d", data.ScrollOffset)
	}

	// Move down several times to trigger scrolling
	for i := 0; i < 15; i++ {
		ReviewUpdate(data, "down")
	}

	if data.SelectedIndex != 15 {
		t.Errorf("expected SelectedIndex to be 15, got %d", data.SelectedIndex)
	}

	// ScrollOffset should have adjusted to keep selection visible
	// (implementation will determine exact behavior)
	if data.ScrollOffset < 0 {
		t.Error("ScrollOffset should not be negative")
	}
}
