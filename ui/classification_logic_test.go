// ui/classification_logic_test.go
package ui

import (
	"clip-tagger/state"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestClassificationLogic_SameAsLast tests the "Same as Last" classification action
func TestClassificationLogic_SameAsLast(t *testing.T) {
	t.Run("classifies with same group as previous file", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)
		group := state.NewGroup("intro", 1)
		appState.Groups = append(appState.Groups, group)

		// Add previous classification
		appState.AddOrUpdateClassification("file1.mp4", group.ID)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4", "file2.mp4"}
		model.currentFileIndex = 1 // Currently on file2.mp4
		model.currentScreen = ScreenClassification
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Perform "Same as Last" action
		updated := model.handleClassificationSameAsLast()

		// Verify file2.mp4 is classified with the same group
		c, found := updated.state.GetClassification("file2.mp4")
		if !found {
			t.Fatal("expected file2.mp4 to be classified")
		}
		if c.GroupID != group.ID {
			t.Errorf("expected group ID %s, got %s", group.ID, c.GroupID)
		}
		if c.TakeNumber != 2 {
			t.Errorf("expected take number 2, got %d", c.TakeNumber)
		}
	})

	t.Run("increments take number correctly", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)
		group := state.NewGroup("scene1", 1)
		appState.Groups = append(appState.Groups, group)

		// Add multiple previous classifications
		appState.AddOrUpdateClassification("file1.mp4", group.ID)
		appState.AddOrUpdateClassification("file2.mp4", group.ID)
		appState.AddOrUpdateClassification("file3.mp4", group.ID)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4", "file2.mp4", "file3.mp4", "file4.mp4"}
		model.currentFileIndex = 3 // Currently on file4.mp4
		model.currentScreen = ScreenClassification
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Perform "Same as Last" action
		updated := model.handleClassificationSameAsLast()

		// Verify take number is 4
		c, found := updated.state.GetClassification("file4.mp4")
		if !found {
			t.Fatal("expected file4.mp4 to be classified")
		}
		if c.TakeNumber != 4 {
			t.Errorf("expected take number 4, got %d", c.TakeNumber)
		}
	})

	t.Run("advances to next file", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)
		group := state.NewGroup("intro", 1)
		appState.Groups = append(appState.Groups, group)

		appState.AddOrUpdateClassification("file1.mp4", group.ID)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4", "file2.mp4", "file3.mp4"}
		model.currentFileIndex = 1
		model.currentScreen = ScreenClassification
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Perform "Same as Last" action
		updated := model.handleClassificationSameAsLast()

		// Verify currentFileIndex advanced
		if updated.currentFileIndex != 2 {
			t.Errorf("expected currentFileIndex to be 2, got %d", updated.currentFileIndex)
		}
	})

	t.Run("updates classification screen data after classification", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)
		group := state.NewGroup("intro", 1)
		appState.Groups = append(appState.Groups, group)

		appState.AddOrUpdateClassification("file1.mp4", group.ID)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4", "file2.mp4", "file3.mp4"}
		model.currentFileIndex = 1
		model.currentScreen = ScreenClassification
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Perform "Same as Last" action
		updated := model.handleClassificationSameAsLast()

		// Verify classification data is updated
		if updated.classificationData == nil {
			t.Fatal("expected classificationData to be updated")
		}
		if updated.classificationData.CurrentFile != "file3.mp4" {
			t.Errorf("expected current file to be file3.mp4, got %s", updated.classificationData.CurrentFile)
		}
		if updated.classificationData.CurrentIndex != 3 {
			t.Errorf("expected current index to be 3, got %d", updated.classificationData.CurrentIndex)
		}
	})

	t.Run("skips already classified file in history search", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)
		group1 := state.NewGroup("intro", 1)
		group2 := state.NewGroup("outro", 2)
		appState.Groups = append(appState.Groups, group1, group2)

		// file1 and file3 are classified, file2 is skipped
		appState.AddOrUpdateClassification("file1.mp4", group1.ID)
		appState.AddOrUpdateClassification("file3.mp4", group2.ID)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4", "file2.mp4", "file3.mp4", "file4.mp4"}
		model.currentFileIndex = 3 // Currently on file4.mp4
		model.currentScreen = ScreenClassification
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Perform "Same as Last" action
		updated := model.handleClassificationSameAsLast()

		// Verify file4.mp4 is classified with group2 (the most recent)
		c, found := updated.state.GetClassification("file4.mp4")
		if !found {
			t.Fatal("expected file4.mp4 to be classified")
		}
		if c.GroupID != group2.ID {
			t.Errorf("expected group ID %s, got %s", group2.ID, c.GroupID)
		}
	})
}

// TestClassificationLogic_SelectExisting tests selecting an existing group
func TestClassificationLogic_SelectExisting(t *testing.T) {
	t.Run("classifies file with selected group", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)
		group := state.NewGroup("scene1", 1)
		appState.Groups = append(appState.Groups, group)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4", "file2.mp4"}
		model.currentFileIndex = 0
		model.currentScreen = ScreenClassification
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Perform group selection
		updated := model.handleGroupSelected(group.ID)

		// Verify file1.mp4 is classified with selected group
		c, found := updated.state.GetClassification("file1.mp4")
		if !found {
			t.Fatal("expected file1.mp4 to be classified")
		}
		if c.GroupID != group.ID {
			t.Errorf("expected group ID %s, got %s", group.ID, c.GroupID)
		}
		if c.TakeNumber != 1 {
			t.Errorf("expected take number 1, got %d", c.TakeNumber)
		}
	})

	t.Run("increments take number for existing group", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)
		group := state.NewGroup("scene1", 1)
		appState.Groups = append(appState.Groups, group)

		// Add previous classifications to this group
		appState.AddOrUpdateClassification("file1.mp4", group.ID)
		appState.AddOrUpdateClassification("file2.mp4", group.ID)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4", "file2.mp4", "file3.mp4"}
		model.currentFileIndex = 2
		model.currentScreen = ScreenClassification
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Perform group selection
		updated := model.handleGroupSelected(group.ID)

		// Verify take number is 3
		c, found := updated.state.GetClassification("file3.mp4")
		if !found {
			t.Fatal("expected file3.mp4 to be classified")
		}
		if c.TakeNumber != 3 {
			t.Errorf("expected take number 3, got %d", c.TakeNumber)
		}
	})

	t.Run("advances to next file", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)
		group := state.NewGroup("scene1", 1)
		appState.Groups = append(appState.Groups, group)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4", "file2.mp4"}
		model.currentFileIndex = 0
		model.currentScreen = ScreenClassification
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Perform group selection
		updated := model.handleGroupSelected(group.ID)

		// Verify currentFileIndex advanced
		if updated.currentFileIndex != 1 {
			t.Errorf("expected currentFileIndex to be 1, got %d", updated.currentFileIndex)
		}
	})

	t.Run("updates classification screen data", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)
		group := state.NewGroup("scene1", 1)
		appState.Groups = append(appState.Groups, group)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4", "file2.mp4"}
		model.currentFileIndex = 0
		model.currentScreen = ScreenClassification
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Perform group selection
		updated := model.handleGroupSelected(group.ID)

		// Verify classification data is updated
		if updated.classificationData == nil {
			t.Fatal("expected classificationData to be updated")
		}
		if updated.classificationData.CurrentFile != "file2.mp4" {
			t.Errorf("expected current file to be file2.mp4, got %s", updated.classificationData.CurrentFile)
		}
	})

	t.Run("transitions back to classification screen", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)
		group := state.NewGroup("scene1", 1)
		appState.Groups = append(appState.Groups, group)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4", "file2.mp4"}
		model.currentFileIndex = 0
		model.currentScreen = ScreenGroupSelection // User is on group selection screen
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Perform group selection
		updated := model.handleGroupSelected(group.ID)

		// Verify screen transitions to classification
		if updated.currentScreen != ScreenClassification {
			t.Errorf("expected screen to be ScreenClassification, got %v", updated.currentScreen)
		}
	})
}

// TestClassificationLogic_CreateNewGroup tests creating a new group and classifying
func TestClassificationLogic_CreateNewGroup(t *testing.T) {
	t.Run("classifies file with new group", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4", "file2.mp4"}
		model.currentFileIndex = 0
		model.currentScreen = ScreenClassification
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Create new group and classify
		newGroup := state.NewGroup("new_scene", 1)
		updated := model.handleGroupInserted(newGroup.ID, newGroup.Name, 1)

		// Verify file1.mp4 is classified with new group
		c, found := updated.state.GetClassification("file1.mp4")
		if !found {
			t.Fatal("expected file1.mp4 to be classified")
		}
		if c.GroupID != newGroup.ID {
			t.Errorf("expected group ID %s, got %s", newGroup.ID, c.GroupID)
		}
		if c.TakeNumber != 1 {
			t.Errorf("expected take number 1 (first file in new group), got %d", c.TakeNumber)
		}
	})

	t.Run("new group is added to state", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)
		existingGroup := state.NewGroup("intro", 1)
		appState.Groups = append(appState.Groups, existingGroup)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4"}
		model.currentFileIndex = 0
		model.currentScreen = ScreenClassification
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Create new group and add it to state (this mimics what GroupInserted message does)
		newGroup := state.NewGroup("outro", 2)
		model.state.Groups = append(model.state.Groups, newGroup)

		// Now handle classification with the new group
		updated := model.handleGroupInserted(newGroup.ID, newGroup.Name, 2)

		// Verify new group is in state
		if len(updated.state.Groups) != 2 {
			t.Fatalf("expected 2 groups, got %d", len(updated.state.Groups))
		}
		found := false
		for _, g := range updated.state.Groups {
			if g.ID == newGroup.ID {
				found = true
				if g.Name != newGroup.Name {
					t.Errorf("expected group name %s, got %s", newGroup.Name, g.Name)
				}
			}
		}
		if !found {
			t.Error("expected new group to be in state")
		}
	})

	t.Run("advances to next file", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4", "file2.mp4"}
		model.currentFileIndex = 0
		model.currentScreen = ScreenClassification
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Create new group
		newGroup := state.NewGroup("scene1", 1)
		updated := model.handleGroupInserted(newGroup.ID, newGroup.Name, 1)

		// Verify currentFileIndex advanced
		if updated.currentFileIndex != 1 {
			t.Errorf("expected currentFileIndex to be 1, got %d", updated.currentFileIndex)
		}
	})

	t.Run("updates classification screen data", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4", "file2.mp4"}
		model.currentFileIndex = 0
		model.currentScreen = ScreenClassification
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Create new group
		newGroup := state.NewGroup("scene1", 1)
		updated := model.handleGroupInserted(newGroup.ID, newGroup.Name, 1)

		// Verify classification data is updated
		if updated.classificationData == nil {
			t.Fatal("expected classificationData to be updated")
		}
		if updated.classificationData.CurrentFile != "file2.mp4" {
			t.Errorf("expected current file to be file2.mp4, got %s", updated.classificationData.CurrentFile)
		}
	})

	t.Run("transitions back to classification screen when more files remain", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4", "file2.mp4"}
		model.currentFileIndex = 0
		model.currentScreen = ScreenGroupInsertion // User is on group insertion screen
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Create new group
		newGroup := state.NewGroup("scene1", 1)
		updated := model.handleGroupInserted(newGroup.ID, newGroup.Name, 1)

		// Verify screen transitions to classification
		if updated.currentScreen != ScreenClassification {
			t.Errorf("expected screen to be ScreenClassification, got %v", updated.currentScreen)
		}
	})
}

// TestClassificationLogic_Skip tests skipping files
func TestClassificationLogic_Skip(t *testing.T) {
	t.Run("adds file to skipped list", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4", "file2.mp4"}
		model.currentFileIndex = 0
		model.currentScreen = ScreenClassification
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Perform skip action
		updated := model.handleClassificationSkip()

		// Verify file1.mp4 is in skipped list
		if len(updated.state.Skipped) != 1 {
			t.Fatalf("expected 1 skipped file, got %d", len(updated.state.Skipped))
		}
		if updated.state.Skipped[0] != "file1.mp4" {
			t.Errorf("expected file1.mp4 to be skipped, got %s", updated.state.Skipped[0])
		}
	})

	t.Run("does not classify skipped file", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4", "file2.mp4"}
		model.currentFileIndex = 0
		model.currentScreen = ScreenClassification
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Perform skip action
		updated := model.handleClassificationSkip()

		// Verify file1.mp4 is NOT classified
		_, found := updated.state.GetClassification("file1.mp4")
		if found {
			t.Error("expected file1.mp4 to NOT be classified")
		}
	})

	t.Run("advances to next file", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4", "file2.mp4"}
		model.currentFileIndex = 0
		model.currentScreen = ScreenClassification
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Perform skip action
		updated := model.handleClassificationSkip()

		// Verify currentFileIndex advanced
		if updated.currentFileIndex != 1 {
			t.Errorf("expected currentFileIndex to be 1, got %d", updated.currentFileIndex)
		}
	})

	t.Run("updates classification screen data", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4", "file2.mp4"}
		model.currentFileIndex = 0
		model.currentScreen = ScreenClassification
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Perform skip action
		updated := model.handleClassificationSkip()

		// Verify classification data is updated
		if updated.classificationData == nil {
			t.Fatal("expected classificationData to be updated")
		}
		if updated.classificationData.CurrentFile != "file2.mp4" {
			t.Errorf("expected current file to be file2.mp4, got %s", updated.classificationData.CurrentFile)
		}
	})

	t.Run("can skip multiple files", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4", "file2.mp4", "file3.mp4"}
		model.currentFileIndex = 0
		model.currentScreen = ScreenClassification
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Skip first file
		updated := model.handleClassificationSkip()
		// Skip second file
		updated = updated.handleClassificationSkip()

		// Verify both files are in skipped list
		if len(updated.state.Skipped) != 2 {
			t.Fatalf("expected 2 skipped files, got %d", len(updated.state.Skipped))
		}
		if updated.state.Skipped[0] != "file1.mp4" {
			t.Errorf("expected first skipped file to be file1.mp4, got %s", updated.state.Skipped[0])
		}
		if updated.state.Skipped[1] != "file2.mp4" {
			t.Errorf("expected second skipped file to be file2.mp4, got %s", updated.state.Skipped[1])
		}
	})
}

// TestClassificationLogic_TransitionToReview tests transitioning to review screen when done
func TestClassificationLogic_TransitionToReview(t *testing.T) {
	t.Run("transitions to review when all files processed", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)
		group := state.NewGroup("scene1", 1)
		appState.Groups = append(appState.Groups, group)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4"} // Only one file
		model.currentFileIndex = 0
		model.currentScreen = ScreenClassification
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Classify the last file
		updated := model.handleGroupSelected(group.ID)

		// Verify screen transitions to review
		if updated.currentScreen != ScreenReview {
			t.Errorf("expected screen to be ScreenReview, got %v", updated.currentScreen)
		}
	})

	t.Run("stays on classification when more files remain", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)
		group := state.NewGroup("scene1", 1)
		appState.Groups = append(appState.Groups, group)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4", "file2.mp4"}
		model.currentFileIndex = 0
		model.currentScreen = ScreenClassification
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Classify first file
		updated := model.handleGroupSelected(group.ID)

		// Verify screen stays on classification
		if updated.currentScreen != ScreenClassification {
			t.Errorf("expected screen to be ScreenClassification, got %v", updated.currentScreen)
		}
	})

	t.Run("transitions to review after skip on last file", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4"}
		model.currentFileIndex = 0
		model.currentScreen = ScreenClassification
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Skip the last file
		updated := model.handleClassificationSkip()

		// Verify screen transitions to review
		if updated.currentScreen != ScreenReview {
			t.Errorf("expected screen to be ScreenReview, got %v", updated.currentScreen)
		}
	})

	t.Run("transitions to review after same as last on last file", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)
		group := state.NewGroup("scene1", 1)
		appState.Groups = append(appState.Groups, group)
		appState.AddOrUpdateClassification("file1.mp4", group.ID)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4", "file2.mp4"}
		model.currentFileIndex = 1
		model.currentScreen = ScreenClassification
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Same as last on the last file
		updated := model.handleClassificationSameAsLast()

		// Verify screen transitions to review
		if updated.currentScreen != ScreenReview {
			t.Errorf("expected screen to be ScreenReview, got %v", updated.currentScreen)
		}
	})
}

// TestClassificationLogic_Integration tests the integration with Update method
func TestClassificationLogic_Integration(t *testing.T) {
	t.Run("full classification flow with multiple files", func(t *testing.T) {
		tmpDir := t.TempDir()
		appState := state.NewState(tmpDir, state.SortByModifiedTime)
		group1 := state.NewGroup("intro", 1)
		group2 := state.NewGroup("outro", 2)
		appState.Groups = append(appState.Groups, group1, group2)

		model := NewModel(appState, tmpDir)
		model.files = []string{"file1.mp4", "file2.mp4", "file3.mp4", "file4.mp4"}
		model.currentFileIndex = 0
		model.currentScreen = ScreenClassification
		model.classificationData = NewClassificationData(appState, model.files, model.currentFileIndex, "")

		// Classify file1 with group1
		msg1 := GroupSelected{GroupID: group1.ID, GroupName: group1.Name}
		updated, _ := model.Update(msg1)
		model = updated.(Model)

		// Verify file1 is classified
		c1, found := model.state.GetClassification("file1.mp4")
		if !found {
			t.Fatal("expected file1.mp4 to be classified")
		}
		if c1.GroupID != group1.ID || c1.TakeNumber != 1 {
			t.Error("file1 classification incorrect")
		}

		// Skip file2
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
		updated, _ = model.Update(keyMsg)
		model = updated.(Model)

		// Verify file2 is skipped
		if len(model.state.Skipped) != 1 || model.state.Skipped[0] != "file2.mp4" {
			t.Error("expected file2.mp4 to be skipped")
		}

		// Classify file3 with group2
		msg2 := GroupSelected{GroupID: group2.ID, GroupName: group2.Name}
		updated, _ = model.Update(msg2)
		model = updated.(Model)

		// Verify file3 is classified
		c3, found := model.state.GetClassification("file3.mp4")
		if !found {
			t.Fatal("expected file3.mp4 to be classified")
		}
		if c3.GroupID != group2.ID || c3.TakeNumber != 1 {
			t.Error("file3 classification incorrect")
		}

		// Same as last for file4 (should use group2)
		keyMsg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}
		updated, _ = model.Update(keyMsg)
		model = updated.(Model)

		// Verify file4 is classified with group2
		c4, found := model.state.GetClassification("file4.mp4")
		if !found {
			t.Fatal("expected file4.mp4 to be classified")
		}
		if c4.GroupID != group2.ID || c4.TakeNumber != 2 {
			t.Error("file4 classification incorrect")
		}

		// Verify we're on review screen
		if model.currentScreen != ScreenReview {
			t.Errorf("expected to be on review screen, got %v", model.currentScreen)
		}
	})
}

