// integration_test.go
package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"clip-tagger/renamer"
	"clip-tagger/scanner"
	"clip-tagger/state"
)

// TestFullWorkflow tests the complete classification workflow
// from startup to classification to review to completion
func TestFullWorkflow(t *testing.T) {
	// Setup: Create temp directory with test video files
	tmpDir := t.TempDir()
	createTestVideoFiles(t, tmpDir, []string{
		"clip1.mp4",
		"clip2.mp4",
		"clip3.mov",
	})

	// Step 1: Initialize new state
	st := state.NewState(tmpDir, state.SortByName)
	if st.Directory != tmpDir {
		t.Errorf("expected directory %s, got %s", tmpDir, st.Directory)
	}
	if st.CurrentIndex != 0 {
		t.Errorf("expected CurrentIndex 0, got %d", st.CurrentIndex)
	}

	// Step 2: Create groups
	intro := state.NewGroup("intro", 1)
	outro := state.NewGroup("outro", 2)
	st.Groups = append(st.Groups, intro, outro)

	if len(st.Groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(st.Groups))
	}

	// Step 3: Scan directory for video files
	scn := scanner.NewScanner(tmpDir)
	result, err := scn.Scan(scanner.SortByName)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if result.Total != 3 {
		t.Errorf("expected 3 video files, got %d", result.Total)
	}

	// Step 4: Classify files
	st.AddOrUpdateClassification("clip1.mp4", intro.ID)
	st.AddOrUpdateClassification("clip2.mp4", intro.ID)
	st.AddOrUpdateClassification("clip3.mov", outro.ID)

	if len(st.Classifications) != 3 {
		t.Errorf("expected 3 classifications, got %d", len(st.Classifications))
	}

	// Verify take numbers are correct
	c1, found := st.GetClassification("clip1.mp4")
	if !found {
		t.Fatal("clip1.mp4 not found")
	}
	if c1.TakeNumber != 1 {
		t.Errorf("expected take 1 for clip1, got %d", c1.TakeNumber)
	}

	c2, found := st.GetClassification("clip2.mp4")
	if !found {
		t.Fatal("clip2.mp4 not found")
	}
	if c2.TakeNumber != 2 {
		t.Errorf("expected take 2 for clip2, got %d", c2.TakeNumber)
	}

	c3, found := st.GetClassification("clip3.mov")
	if !found {
		t.Fatal("clip3.mov not found")
	}
	if c3.TakeNumber != 1 {
		t.Errorf("expected take 1 for clip3, got %d", c3.TakeNumber)
	}

	// Step 5: Generate rename operations
	var renames []renamer.Rename
	for _, c := range st.Classifications {
		group := st.FindGroupByID(c.GroupID)
		if group == nil {
			t.Fatalf("group not found for classification %s", c.File)
		}

		originalPath := filepath.Join(tmpDir, c.File)
		targetPath := renamer.GenerateTargetPath(tmpDir, originalPath, group.Order, c.TakeNumber, group.Name)

		renames = append(renames, renamer.Rename{
			OriginalPath: originalPath,
			TargetPath:   targetPath,
		})
	}

	if len(renames) != 3 {
		t.Errorf("expected 3 renames, got %d", len(renames))
	}

	// Step 6: Check for conflicts (should be none)
	conflicts := renamer.DetectConflicts(renames)
	if len(conflicts) != 0 {
		t.Errorf("expected no conflicts, got %d", len(conflicts))
	}

	// Step 7: Execute rename operations
	err = renamer.RenameInPlace(renames)
	if err != nil {
		t.Fatalf("rename failed: %v", err)
	}

	// Step 8: Verify files were renamed correctly
	expectedFiles := []string{
		"[01_01] intro.mp4",
		"[01_02] intro.mp4",
		"[02_01] outro.mov",
	}

	for _, filename := range expectedFiles {
		path := filepath.Join(tmpDir, filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist after rename", filename)
		}
	}

	// Step 9: Save final state
	statePath := state.StateFilePath(tmpDir)
	err = st.Save(statePath)
	if err != nil {
		t.Fatalf("save state failed: %v", err)
	}

	// Verify state file exists
	if !state.StateExists(tmpDir) {
		t.Error("expected state file to exist after save")
	}
}

// TestResumeWorkflow tests resuming classification from existing state
func TestResumeWorkflow(t *testing.T) {
	// Setup: Create temp directory with test files
	tmpDir := t.TempDir()
	createTestVideoFiles(t, tmpDir, []string{
		"video1.mp4",
		"video2.mp4",
		"video3.mp4",
	})

	// Step 1: Create initial state with some classifications
	st := state.NewState(tmpDir, state.SortByName)
	intro := state.NewGroup("intro", 1)
	st.Groups = append(st.Groups, intro)
	st.AddOrUpdateClassification("video1.mp4", intro.ID)
	st.CurrentIndex = 1

	// Save state
	statePath := state.StateFilePath(tmpDir)
	err := st.Save(statePath)
	if err != nil {
		t.Fatalf("save state failed: %v", err)
	}

	// Step 2: Simulate app restart - load existing state
	loadedState, err := state.Load(statePath)
	if err != nil {
		t.Fatalf("load state failed: %v", err)
	}

	// Verify loaded state matches original
	if loadedState.Directory != tmpDir {
		t.Errorf("expected directory %s, got %s", tmpDir, loadedState.Directory)
	}
	if loadedState.CurrentIndex != 1 {
		t.Errorf("expected CurrentIndex 1, got %d", loadedState.CurrentIndex)
	}
	if len(loadedState.Groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(loadedState.Groups))
	}
	if len(loadedState.Classifications) != 1 {
		t.Errorf("expected 1 classification, got %d", len(loadedState.Classifications))
	}

	// Step 3: Continue classification from where we left off
	loadedState.AddOrUpdateClassification("video2.mp4", intro.ID)
	loadedState.CurrentIndex = 2

	if len(loadedState.Classifications) != 2 {
		t.Errorf("expected 2 classifications after resume, got %d", len(loadedState.Classifications))
	}

	// Step 4: Verify take numbers continue correctly
	c1, found := loadedState.GetClassification("video1.mp4")
	if !found {
		t.Fatal("video1.mp4 not found")
	}
	if c1.TakeNumber != 1 {
		t.Errorf("expected take 1, got %d", c1.TakeNumber)
	}

	c2, found := loadedState.GetClassification("video2.mp4")
	if !found {
		t.Fatal("video2.mp4 not found")
	}
	if c2.TakeNumber != 2 {
		t.Errorf("expected take 2, got %d", c2.TakeNumber)
	}

	// Step 5: Save updated state
	err = loadedState.Save(statePath)
	if err != nil {
		t.Fatalf("save updated state failed: %v", err)
	}

	// Step 6: Load again to verify persistence
	finalState, err := state.Load(statePath)
	if err != nil {
		t.Fatalf("load final state failed: %v", err)
	}

	if finalState.CurrentIndex != 2 {
		t.Errorf("expected CurrentIndex 2, got %d", finalState.CurrentIndex)
	}
	if len(finalState.Classifications) != 2 {
		t.Errorf("expected 2 classifications, got %d", len(finalState.Classifications))
	}
}

// TestConflictDetection tests end-to-end conflict detection
func TestConflictDetection(t *testing.T) {
	// Setup: Create temp directory with test files
	tmpDir := t.TempDir()
	createTestVideoFiles(t, tmpDir, []string{
		"clip1.mp4",
		"clip2.mp4",
	})

	// Create a file that will conflict with the target name
	conflictFile := filepath.Join(tmpDir, "[01_01] intro.mp4")
	if err := os.WriteFile(conflictFile, []byte("existing"), 0644); err != nil {
		t.Fatalf("failed to create conflict file: %v", err)
	}

	// Step 1: Initialize state and classify
	st := state.NewState(tmpDir, state.SortByName)
	intro := state.NewGroup("intro", 1)
	st.Groups = append(st.Groups, intro)
	st.AddOrUpdateClassification("clip1.mp4", intro.ID)

	// Step 2: Generate rename operations
	group := st.FindGroupByID(intro.ID)
	originalPath := filepath.Join(tmpDir, "clip1.mp4")
	targetPath := renamer.GenerateTargetPath(tmpDir, originalPath, group.Order, 1, group.Name)

	renames := []renamer.Rename{
		{
			OriginalPath: originalPath,
			TargetPath:   targetPath,
		},
	}

	// Step 3: Detect conflicts
	conflicts := renamer.DetectConflicts(renames)

	// Should detect the conflict
	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}

	if conflicts[0].TargetPath != conflictFile {
		t.Errorf("expected conflict on %s, got %s", conflictFile, conflicts[0].TargetPath)
	}

	// Step 4: Verify the existing file content before rename
	originalContent, err := os.ReadFile(conflictFile)
	if err != nil {
		t.Fatalf("failed to read conflict file: %v", err)
	}
	if string(originalContent) != "existing" {
		t.Errorf("expected conflict file to contain 'existing', got '%s'", string(originalContent))
	}

	// Note: os.Rename will overwrite on Unix systems, which is why we detect conflicts first.
	// In a real application, the UI would warn the user before proceeding with the rename.
}

// TestAutoSaveCheckpoints tests that state persists at checkpoints
func TestAutoSaveCheckpoints(t *testing.T) {
	// Setup: Create temp directory
	tmpDir := t.TempDir()
	createTestVideoFiles(t, tmpDir, []string{
		"video1.mp4",
		"video2.mp4",
		"video3.mp4",
	})

	statePath := state.StateFilePath(tmpDir)

	// Checkpoint 1: Initial state with groups
	st := state.NewState(tmpDir, state.SortByName)
	intro := state.NewGroup("intro", 1)
	st.Groups = append(st.Groups, intro)

	err := st.Save(statePath)
	if err != nil {
		t.Fatalf("checkpoint 1 save failed: %v", err)
	}

	// Verify checkpoint 1
	loaded1, err := state.Load(statePath)
	if err != nil {
		t.Fatalf("checkpoint 1 load failed: %v", err)
	}
	if len(loaded1.Groups) != 1 {
		t.Errorf("checkpoint 1: expected 1 group, got %d", len(loaded1.Groups))
	}

	// Checkpoint 2: After first classification
	st.AddOrUpdateClassification("video1.mp4", intro.ID)
	st.CurrentIndex = 1

	err = st.Save(statePath)
	if err != nil {
		t.Fatalf("checkpoint 2 save failed: %v", err)
	}

	// Verify checkpoint 2
	loaded2, err := state.Load(statePath)
	if err != nil {
		t.Fatalf("checkpoint 2 load failed: %v", err)
	}
	if len(loaded2.Classifications) != 1 {
		t.Errorf("checkpoint 2: expected 1 classification, got %d", len(loaded2.Classifications))
	}
	if loaded2.CurrentIndex != 1 {
		t.Errorf("checkpoint 2: expected index 1, got %d", loaded2.CurrentIndex)
	}

	// Checkpoint 3: After more classifications
	st.AddOrUpdateClassification("video2.mp4", intro.ID)
	st.AddOrUpdateClassification("video3.mp4", intro.ID)
	st.CurrentIndex = 3

	err = st.Save(statePath)
	if err != nil {
		t.Fatalf("checkpoint 3 save failed: %v", err)
	}

	// Verify checkpoint 3
	loaded3, err := state.Load(statePath)
	if err != nil {
		t.Fatalf("checkpoint 3 load failed: %v", err)
	}
	if len(loaded3.Classifications) != 3 {
		t.Errorf("checkpoint 3: expected 3 classifications, got %d", len(loaded3.Classifications))
	}
	if loaded3.CurrentIndex != 3 {
		t.Errorf("checkpoint 3: expected index 3, got %d", loaded3.CurrentIndex)
	}

	// Checkpoint 4: Test backup functionality
	err = state.BackupState(tmpDir)
	if err != nil {
		t.Fatalf("backup failed: %v", err)
	}

	backupPath := statePath + ".bak"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("expected backup file to exist")
	}

	// Verify backup contains same data
	loadedBackup, err := state.Load(backupPath)
	if err != nil {
		t.Fatalf("load backup failed: %v", err)
	}
	if len(loadedBackup.Classifications) != 3 {
		t.Errorf("backup: expected 3 classifications, got %d", len(loadedBackup.Classifications))
	}
}

// TestMergeNewFiles tests handling new files added to directory
func TestMergeNewFiles(t *testing.T) {
	// Setup: Create temp directory with initial files
	tmpDir := t.TempDir()
	createTestVideoFiles(t, tmpDir, []string{
		"video1.mp4",
		"video2.mp4",
	})

	// Step 1: Create initial state
	st := state.NewState(tmpDir, state.SortByName)
	intro := state.NewGroup("intro", 1)
	st.Groups = append(st.Groups, intro)
	st.AddOrUpdateClassification("video1.mp4", intro.ID)

	// Step 2: Add new files to directory (simulate user adding more files)
	createTestVideoFiles(t, tmpDir, []string{
		"video3.mp4",
		"video4.mp4",
	})

	// Step 3: Scan directory
	scn := scanner.NewScanner(tmpDir)
	result, err := scn.Scan(scanner.SortByName)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	// Build filename list
	var scannedFiles []string
	for _, f := range result.Files {
		scannedFiles = append(scannedFiles, f.Name)
	}

	// Step 4: Merge with existing state
	mergeResult := state.MergeFiles(st, scannedFiles)

	// Verify merge results
	if len(mergeResult.NewFiles) != 3 {
		t.Errorf("expected 3 new files (video2, video3, video4), got %d", len(mergeResult.NewFiles))
	}
	if mergeResult.ExistingCount != 1 {
		t.Errorf("expected 1 existing file, got %d", mergeResult.ExistingCount)
	}
	if len(mergeResult.MissingFiles) != 0 {
		t.Errorf("expected 0 missing files, got %d", len(mergeResult.MissingFiles))
	}

	// Step 5: Classify new files
	for _, filename := range mergeResult.NewFiles {
		st.AddOrUpdateClassification(filename, intro.ID)
	}

	if len(st.Classifications) != 4 {
		t.Errorf("expected 4 classifications after merge, got %d", len(st.Classifications))
	}
}

// TestMissingFiles tests handling files that were removed from directory
func TestMissingFiles(t *testing.T) {
	// Setup: Create temp directory with files
	tmpDir := t.TempDir()
	createTestVideoFiles(t, tmpDir, []string{
		"video1.mp4",
		"video2.mp4",
		"video3.mp4",
	})

	// Step 1: Create state with all files classified
	st := state.NewState(tmpDir, state.SortByName)
	intro := state.NewGroup("intro", 1)
	st.Groups = append(st.Groups, intro)
	st.AddOrUpdateClassification("video1.mp4", intro.ID)
	st.AddOrUpdateClassification("video2.mp4", intro.ID)
	st.AddOrUpdateClassification("video3.mp4", intro.ID)

	// Step 2: Remove a file (simulate user deleting)
	removedFile := filepath.Join(tmpDir, "video2.mp4")
	if err := os.Remove(removedFile); err != nil {
		t.Fatalf("failed to remove file: %v", err)
	}

	// Step 3: Scan directory
	scn := scanner.NewScanner(tmpDir)
	result, err := scn.Scan(scanner.SortByName)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	var scannedFiles []string
	for _, f := range result.Files {
		scannedFiles = append(scannedFiles, f.Name)
	}

	// Step 4: Merge with existing state
	mergeResult := state.MergeFiles(st, scannedFiles)

	// Verify missing file was detected
	if len(mergeResult.MissingFiles) != 1 {
		t.Fatalf("expected 1 missing file, got %d", len(mergeResult.MissingFiles))
	}
	if mergeResult.MissingFiles[0] != "video2.mp4" {
		t.Errorf("expected missing file 'video2.mp4', got '%s'", mergeResult.MissingFiles[0])
	}
	if mergeResult.ExistingCount != 2 {
		t.Errorf("expected 2 existing files, got %d", mergeResult.ExistingCount)
	}
}

// TestReclassification tests changing a file's classification
func TestReclassification(t *testing.T) {
	// Setup: Create temp directory with files
	tmpDir := t.TempDir()
	createTestVideoFiles(t, tmpDir, []string{
		"video1.mp4",
	})

	// Step 1: Create state and classify file
	st := state.NewState(tmpDir, state.SortByName)
	intro := state.NewGroup("intro", 1)
	outro := state.NewGroup("outro", 2)
	st.Groups = append(st.Groups, intro, outro)

	// Initial classification to intro
	st.AddOrUpdateClassification("video1.mp4", intro.ID)

	c1, found := st.GetClassification("video1.mp4")
	if !found {
		t.Fatal("expected to find initial classification")
	}
	if c1.GroupID != intro.ID {
		t.Errorf("expected groupID %s, got %s", intro.ID, c1.GroupID)
	}
	if c1.TakeNumber != 1 {
		t.Errorf("expected take 1, got %d", c1.TakeNumber)
	}

	// Step 2: Reclassify to outro
	st.AddOrUpdateClassification("video1.mp4", outro.ID)

	// Should only have one classification
	if len(st.Classifications) != 1 {
		t.Errorf("expected 1 classification after reclassification, got %d", len(st.Classifications))
	}

	c2, found := st.GetClassification("video1.mp4")
	if !found {
		t.Fatal("expected to find reclassification")
	}
	if c2.GroupID != outro.ID {
		t.Errorf("expected groupID %s after reclassification, got %s", outro.ID, c2.GroupID)
	}
	if c2.TakeNumber != 1 {
		t.Errorf("expected take 1 for outro group, got %d", c2.TakeNumber)
	}
}

// TestCopyToDirectory tests the copy operation
func TestCopyToDirectory(t *testing.T) {
	// Setup: Create temp directory with test files
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")

	createTestVideoFiles(t, tmpDir, []string{
		"video1.mp4",
		"video2.mp4",
	})

	// Step 1: Create state and classify
	st := state.NewState(tmpDir, state.SortByName)
	intro := state.NewGroup("intro", 1)
	st.Groups = append(st.Groups, intro)
	st.AddOrUpdateClassification("video1.mp4", intro.ID)
	st.AddOrUpdateClassification("video2.mp4", intro.ID)

	// Step 2: Generate renames
	var renames []renamer.Rename
	for _, c := range st.Classifications {
		group := st.FindGroupByID(c.GroupID)
		originalPath := filepath.Join(tmpDir, c.File)
		targetPath := renamer.GenerateTargetPath(tmpDir, originalPath, group.Order, c.TakeNumber, group.Name)

		renames = append(renames, renamer.Rename{
			OriginalPath: originalPath,
			TargetPath:   targetPath,
		})
	}

	// Step 3: Copy to output directory
	err := renamer.CopyToDirectory(renames, outputDir)
	if err != nil {
		t.Fatalf("copy failed: %v", err)
	}

	// Step 4: Verify files were copied
	expectedFiles := []string{
		"[01_01] intro.mp4",
		"[01_02] intro.mp4",
	}

	for _, filename := range expectedFiles {
		path := filepath.Join(outputDir, filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist in output directory", filename)
		}
	}

	// Step 5: Verify original files still exist
	originalFiles := []string{"video1.mp4", "video2.mp4"}
	for _, filename := range originalFiles {
		path := filepath.Join(tmpDir, filename)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected original file %s to still exist", filename)
		}
	}
}

// TestSkippedFiles tests tracking skipped files
func TestSkippedFiles(t *testing.T) {
	tmpDir := t.TempDir()
	createTestVideoFiles(t, tmpDir, []string{
		"video1.mp4",
		"video2.mp4",
		"video3.mp4",
	})

	st := state.NewState(tmpDir, state.SortByName)
	intro := state.NewGroup("intro", 1)
	st.Groups = append(st.Groups, intro)

	// Classify some, skip others
	st.AddOrUpdateClassification("video1.mp4", intro.ID)
	st.Skipped = append(st.Skipped, "video2.mp4")
	st.AddOrUpdateClassification("video3.mp4", intro.ID)

	if len(st.Classifications) != 2 {
		t.Errorf("expected 2 classifications, got %d", len(st.Classifications))
	}
	if len(st.Skipped) != 1 {
		t.Errorf("expected 1 skipped file, got %d", len(st.Skipped))
	}

	// Save and reload
	statePath := state.StateFilePath(tmpDir)
	err := st.Save(statePath)
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := state.Load(statePath)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if len(loaded.Skipped) != 1 {
		t.Errorf("expected 1 skipped file after reload, got %d", len(loaded.Skipped))
	}
	if loaded.Skipped[0] != "video2.mp4" {
		t.Errorf("expected skipped file 'video2.mp4', got '%s'", loaded.Skipped[0])
	}
}

// TestFileSorting tests that files are scanned in correct sort order
func TestFileSorting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create files with different timestamps
	files := []string{"c.mp4", "a.mp4", "b.mp4"}
	for i, filename := range files {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte("data"), 0644); err != nil {
			t.Fatalf("failed to create file: %v", err)
		}

		// Set different modification times
		modTime := time.Now().Add(time.Duration(i) * time.Hour)
		if err := os.Chtimes(path, modTime, modTime); err != nil {
			t.Fatalf("failed to set mod time: %v", err)
		}
	}

	// Test sort by name
	scn := scanner.NewScanner(tmpDir)
	result, err := scn.Scan(scanner.SortByName)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	expectedOrder := []string{"a.mp4", "b.mp4", "c.mp4"}
	for i, expected := range expectedOrder {
		if result.Files[i].Name != expected {
			t.Errorf("expected file %d to be %s, got %s", i, expected, result.Files[i].Name)
		}
	}

	// Test sort by modified time
	result, err = scn.Scan(scanner.SortByModifiedTime)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	// Should be in order of creation (c, a, b by timestamp)
	expectedTimeOrder := []string{"c.mp4", "a.mp4", "b.mp4"}
	for i, expected := range expectedTimeOrder {
		if result.Files[i].Name != expected {
			t.Errorf("expected file %d to be %s, got %s", i, expected, result.Files[i].Name)
		}
	}
}

// TestMultipleGroups tests classification with multiple groups
func TestMultipleGroups(t *testing.T) {
	tmpDir := t.TempDir()
	createTestVideoFiles(t, tmpDir, []string{
		"video1.mp4",
		"video2.mp4",
		"video3.mp4",
		"video4.mp4",
	})

	st := state.NewState(tmpDir, state.SortByName)
	intro := state.NewGroup("intro", 1)
	main := state.NewGroup("main", 2)
	outro := state.NewGroup("outro", 3)
	st.Groups = append(st.Groups, intro, main, outro)

	// Classify files to different groups
	st.AddOrUpdateClassification("video1.mp4", intro.ID)
	st.AddOrUpdateClassification("video2.mp4", main.ID)
	st.AddOrUpdateClassification("video3.mp4", main.ID)
	st.AddOrUpdateClassification("video4.mp4", outro.ID)

	// Verify take numbers per group
	c1, _ := st.GetClassification("video1.mp4")
	if c1.TakeNumber != 1 {
		t.Errorf("expected take 1 for intro, got %d", c1.TakeNumber)
	}

	c2, _ := st.GetClassification("video2.mp4")
	if c2.TakeNumber != 1 {
		t.Errorf("expected take 1 for main, got %d", c2.TakeNumber)
	}

	c3, _ := st.GetClassification("video3.mp4")
	if c3.TakeNumber != 2 {
		t.Errorf("expected take 2 for main, got %d", c3.TakeNumber)
	}

	c4, _ := st.GetClassification("video4.mp4")
	if c4.TakeNumber != 1 {
		t.Errorf("expected take 1 for outro, got %d", c4.TakeNumber)
	}

	// Generate and verify filenames
	expectedNames := map[string]string{
		"video1.mp4": "[01_01] intro.mp4",
		"video2.mp4": "[02_01] main.mp4",
		"video3.mp4": "[02_02] main.mp4",
		"video4.mp4": "[03_01] outro.mp4",
	}

	for _, c := range st.Classifications {
		group := st.FindGroupByID(c.GroupID)
		filename := renamer.GenerateFilename(group.Order, c.TakeNumber, group.Name, ".mp4")
		expected := expectedNames[c.File]
		if filename != expected {
			t.Errorf("expected filename %s, got %s", expected, filename)
		}
	}
}

// Helper function to create test video files
func createTestVideoFiles(t *testing.T, dir string, filenames []string) {
	t.Helper()
	for _, filename := range filenames {
		path := filepath.Join(dir, filename)
		if err := os.WriteFile(path, []byte("test video data"), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", filename, err)
		}
	}
}
