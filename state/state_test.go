// state/state_test.go
package state

import (
	"os"
	"strings"
	"testing"
)

func TestState_NextTakeNumber(t *testing.T) {
	state := &State{
		Groups: []Group{
			{ID: "group1", Name: "intro", Order: 1},
		},
		Classifications: []Classification{
			{File: "vid1.mp4", GroupID: "group1", TakeNumber: 1},
			{File: "vid2.mp4", GroupID: "group1", TakeNumber: 2},
		},
	}

	nextTake := state.NextTakeNumber("group1")
	if nextTake != 3 {
		t.Errorf("expected take 3, got %d", nextTake)
	}
}

func TestState_FindGroupByID(t *testing.T) {
	group1 := Group{ID: "id1", Name: "intro", Order: 1}
	state := &State{
		Groups: []Group{group1},
	}

	found := state.FindGroupByID("id1")
	if found == nil {
		t.Fatal("expected to find group")
	}
	if found.Name != "intro" {
		t.Errorf("expected 'intro', got '%s'", found.Name)
	}

	notFound := state.FindGroupByID("nonexistent")
	if notFound != nil {
		t.Error("expected nil for nonexistent group")
	}
}

func TestState_NextTakeNumber_EdgeCases(t *testing.T) {
	t.Run("empty classifications returns 1", func(t *testing.T) {
		state := &State{
			Classifications: []Classification{},
		}
		nextTake := state.NextTakeNumber("group1")
		if nextTake != 1 {
			t.Errorf("expected take 1 for empty classifications, got %d", nextTake)
		}
	})

	t.Run("group with no classifications returns 1", func(t *testing.T) {
		state := &State{
			Classifications: []Classification{
				{File: "vid1.mp4", GroupID: "group1", TakeNumber: 5},
				{File: "vid2.mp4", GroupID: "group1", TakeNumber: 3},
			},
		}
		nextTake := state.NextTakeNumber("group2")
		if nextTake != 1 {
			t.Errorf("expected take 1 for group with no classifications, got %d", nextTake)
		}
	})
}

func TestState_GetClassification(t *testing.T) {
	t.Run("finding existing file", func(t *testing.T) {
		state := &State{
			Classifications: []Classification{
				{File: "vid1.mp4", GroupID: "group1", TakeNumber: 1},
				{File: "vid2.mp4", GroupID: "group2", TakeNumber: 2},
			},
		}

		c, found := state.GetClassification("vid1.mp4")
		if !found {
			t.Fatal("expected to find classification")
		}
		if c.File != "vid1.mp4" {
			t.Errorf("expected file 'vid1.mp4', got '%s'", c.File)
		}
		if c.GroupID != "group1" {
			t.Errorf("expected groupID 'group1', got '%s'", c.GroupID)
		}
		if c.TakeNumber != 1 {
			t.Errorf("expected take 1, got %d", c.TakeNumber)
		}
	})

	t.Run("missing file returns false", func(t *testing.T) {
		state := &State{
			Classifications: []Classification{
				{File: "vid1.mp4", GroupID: "group1", TakeNumber: 1},
			},
		}

		_, found := state.GetClassification("nonexistent.mp4")
		if found {
			t.Error("expected not to find classification")
		}
	})

	t.Run("empty state returns false", func(t *testing.T) {
		state := &State{
			Classifications: []Classification{},
		}

		_, found := state.GetClassification("vid1.mp4")
		if found {
			t.Error("expected not to find classification in empty state")
		}
	})
}

func TestState_AddOrUpdateClassification(t *testing.T) {
	t.Run("adding new classification", func(t *testing.T) {
		state := &State{
			Groups: []Group{
				{ID: "group1", Name: "intro", Order: 1},
			},
			Classifications: []Classification{},
		}

		state.AddOrUpdateClassification("vid1.mp4", "group1")

		if len(state.Classifications) != 1 {
			t.Fatalf("expected 1 classification, got %d", len(state.Classifications))
		}

		c := state.Classifications[0]
		if c.File != "vid1.mp4" {
			t.Errorf("expected file 'vid1.mp4', got '%s'", c.File)
		}
		if c.GroupID != "group1" {
			t.Errorf("expected groupID 'group1', got '%s'", c.GroupID)
		}
		if c.TakeNumber != 1 {
			t.Errorf("expected take 1, got %d", c.TakeNumber)
		}
	})

	t.Run("updating existing classification removes old", func(t *testing.T) {
		state := &State{
			Groups: []Group{
				{ID: "group1", Name: "intro", Order: 1},
				{ID: "group2", Name: "outro", Order: 2},
			},
			Classifications: []Classification{
				{File: "vid1.mp4", GroupID: "group1", TakeNumber: 1},
				{File: "vid2.mp4", GroupID: "group1", TakeNumber: 2},
			},
		}

		state.AddOrUpdateClassification("vid1.mp4", "group2")

		if len(state.Classifications) != 2 {
			t.Fatalf("expected 2 classifications, got %d", len(state.Classifications))
		}

		// Verify old classification was removed
		for _, c := range state.Classifications {
			if c.File == "vid1.mp4" {
				if c.GroupID != "group2" {
					t.Errorf("expected updated groupID 'group2', got '%s'", c.GroupID)
				}
				if c.TakeNumber != 1 {
					t.Errorf("expected take 1 for group2, got %d", c.TakeNumber)
				}
			}
		}
	})

	t.Run("take numbers increment correctly", func(t *testing.T) {
		state := &State{
			Groups: []Group{
				{ID: "group1", Name: "intro", Order: 1},
			},
			Classifications: []Classification{
				{File: "vid1.mp4", GroupID: "group1", TakeNumber: 1},
				{File: "vid2.mp4", GroupID: "group1", TakeNumber: 2},
			},
		}

		state.AddOrUpdateClassification("vid3.mp4", "group1")

		c, found := state.GetClassification("vid3.mp4")
		if !found {
			t.Fatal("expected to find new classification")
		}
		if c.TakeNumber != 3 {
			t.Errorf("expected take 3, got %d", c.TakeNumber)
		}
	})
}

func TestNewState(t *testing.T) {
	directory := "/test/dir"
	sortBy := SortByModifiedTime

	state := NewState(directory, sortBy)

	if state.Directory != directory {
		t.Errorf("expected directory '%s', got '%s'", directory, state.Directory)
	}
	if state.SortBy != sortBy {
		t.Errorf("expected sortBy '%s', got '%s'", sortBy, state.SortBy)
	}
	if state.CurrentIndex != 0 {
		t.Errorf("expected CurrentIndex 0, got %d", state.CurrentIndex)
	}
	if state.Groups == nil {
		t.Error("expected Groups to be initialized, got nil")
	}
	if len(state.Groups) != 0 {
		t.Errorf("expected empty Groups, got length %d", len(state.Groups))
	}
	if state.Classifications == nil {
		t.Error("expected Classifications to be initialized, got nil")
	}
	if len(state.Classifications) != 0 {
		t.Errorf("expected empty Classifications, got length %d", len(state.Classifications))
	}
	if state.Skipped == nil {
		t.Error("expected Skipped to be initialized, got nil")
	}
	if len(state.Skipped) != 0 {
		t.Errorf("expected empty Skipped, got length %d", len(state.Skipped))
	}
}

func TestNewGroup(t *testing.T) {
	name := "intro"
	order := 1

	group := NewGroup(name, order)

	if group.Name != name {
		t.Errorf("expected name '%s', got '%s'", name, group.Name)
	}
	if group.Order != order {
		t.Errorf("expected order %d, got %d", order, group.Order)
	}
	if group.ID == "" {
		t.Error("expected non-empty ID")
	}

	// Verify UUID format (basic check - should have dashes)
	if len(group.ID) < 32 {
		t.Errorf("expected UUID format, got '%s'", group.ID)
	}

	// Verify UUIDs are unique
	group2 := NewGroup("outro", 2)
	if group.ID == group2.ID {
		t.Error("expected unique IDs for different groups")
	}
}

func TestState_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := tmpDir + "/.clip-tagger-state.json"

	// Create and save state
	original := NewState(tmpDir, SortByModifiedTime)
	original.Groups = append(original.Groups, NewGroup("intro", 1))
	original.CurrentIndex = 5

	err := original.Save(statePath)
	if err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Load state
	loaded, err := Load(statePath)
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	if loaded.Directory != original.Directory {
		t.Errorf("directory mismatch")
	}
	if loaded.CurrentIndex != 5 {
		t.Errorf("expected index 5, got %d", loaded.CurrentIndex)
	}
	if len(loaded.Groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(loaded.Groups))
	}

	// Verify JSON uses 2-space indentation (not tabs)
	content, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}
	contentStr := string(content)

	// Check for 2-space indentation pattern
	if !strings.Contains(contentStr, "  \"directory\"") {
		t.Error("expected JSON to use 2-space indentation")
	}
	// Ensure no tabs are used
	if strings.Contains(contentStr, "\t") {
		t.Error("expected JSON to not contain tabs")
	}
}

func TestLoad_NonExistent(t *testing.T) {
	_, err := Load("/nonexistent/path/.clip-tagger-state.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestStateFilePath(t *testing.T) {
	t.Run("returns correct path with directory", func(t *testing.T) {
		dir := "/test/dir"
		expected := "/test/dir/.clip-tagger-state.json"
		result := StateFilePath(dir)

		if result != expected {
			t.Errorf("expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("handles directory with trailing slash", func(t *testing.T) {
		dir := "/test/dir/"
		result := StateFilePath(dir)

		// filepath.Join should handle trailing slashes correctly
		if !strings.Contains(result, ".clip-tagger-state.json") {
			t.Errorf("expected path to contain state filename, got '%s'", result)
		}
	})

	t.Run("handles relative directory", func(t *testing.T) {
		dir := "."
		result := StateFilePath(dir)
		expected := ".clip-tagger-state.json"

		if result != expected {
			t.Errorf("expected '%s', got '%s'", expected, result)
		}
	})
}

func TestStateExists(t *testing.T) {
	t.Run("returns true when state file exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		statePath := StateFilePath(tmpDir)

		// Create a state file
		state := NewState(tmpDir, SortByModifiedTime)
		err := state.Save(statePath)
		if err != nil {
			t.Fatalf("failed to create state file: %v", err)
		}

		// Verify StateExists returns true
		if !StateExists(tmpDir) {
			t.Error("expected StateExists to return true when file exists")
		}
	})

	t.Run("returns false when state file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Verify StateExists returns false
		if StateExists(tmpDir) {
			t.Error("expected StateExists to return false when file does not exist")
		}
	})

	t.Run("returns false for nonexistent directory", func(t *testing.T) {
		dir := "/nonexistent/directory/path"

		if StateExists(dir) {
			t.Error("expected StateExists to return false for nonexistent directory")
		}
	})
}

func TestBackupState(t *testing.T) {
	t.Run("returns error when state file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := BackupState(tmpDir)
		if err == nil {
			t.Fatal("expected error when state file does not exist")
		}

		expectedErrMsg := "state file does not exist"
		if !strings.Contains(err.Error(), expectedErrMsg) {
			t.Errorf("expected error message to contain '%s', got '%s'", expectedErrMsg, err.Error())
		}
	})

	t.Run("creates backup successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		statePath := StateFilePath(tmpDir)
		backupPath := statePath + ".bak"

		// Create a state file
		original := NewState(tmpDir, SortByModifiedTime)
		original.Groups = append(original.Groups, NewGroup("intro", 1))
		original.CurrentIndex = 10

		err := original.Save(statePath)
		if err != nil {
			t.Fatalf("failed to create state file: %v", err)
		}

		// Create backup
		err = BackupState(tmpDir)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		// Verify backup file exists
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			t.Fatal("expected backup file to exist")
		}
	})

	t.Run("backup content matches original", func(t *testing.T) {
		tmpDir := t.TempDir()
		statePath := StateFilePath(tmpDir)
		backupPath := statePath + ".bak"

		// Create a state file with specific content
		original := NewState(tmpDir, SortByModifiedTime)
		original.Groups = append(original.Groups, NewGroup("intro", 1))
		original.Groups = append(original.Groups, NewGroup("outro", 2))
		original.CurrentIndex = 15
		original.Classifications = append(original.Classifications, Classification{
			File:       "test.mp4",
			GroupID:    "group1",
			TakeNumber: 3,
		})

		err := original.Save(statePath)
		if err != nil {
			t.Fatalf("failed to create state file: %v", err)
		}

		// Create backup
		err = BackupState(tmpDir)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		// Read original and backup content
		originalContent, err := os.ReadFile(statePath)
		if err != nil {
			t.Fatalf("failed to read original file: %v", err)
		}

		backupContent, err := os.ReadFile(backupPath)
		if err != nil {
			t.Fatalf("failed to read backup file: %v", err)
		}

		// Verify content matches
		if string(originalContent) != string(backupContent) {
			t.Error("expected backup content to match original content")
		}

		// Verify both can be loaded as valid state
		loadedBackup, err := Load(backupPath)
		if err != nil {
			t.Fatalf("failed to load backup: %v", err)
		}

		if loadedBackup.CurrentIndex != 15 {
			t.Errorf("expected CurrentIndex 15, got %d", loadedBackup.CurrentIndex)
		}
		if len(loadedBackup.Groups) != 2 {
			t.Errorf("expected 2 groups, got %d", len(loadedBackup.Groups))
		}
		if len(loadedBackup.Classifications) != 1 {
			t.Errorf("expected 1 classification, got %d", len(loadedBackup.Classifications))
		}
	})
}
