// state/merger_test.go
package state

import (
	"testing"
)

func TestMergeFiles(t *testing.T) {
	state := &State{
		Directory:       "/test",
		Classifications: []Classification{
			{File: "existing1.mp4", GroupID: "g1", TakeNumber: 1},
			{File: "existing2.mp4", GroupID: "g2", TakeNumber: 1},
		},
	}

	scannedFiles := []string{
		"existing1.mp4",
		"existing2.mp4",
		"new1.mp4",
		"new2.mp4",
	}

	result := MergeFiles(state, scannedFiles)

	if len(result.NewFiles) != 2 {
		t.Errorf("expected 2 new files, got %d", len(result.NewFiles))
	}

	if len(result.MissingFiles) != 0 {
		t.Errorf("expected 0 missing files, got %d", len(result.MissingFiles))
	}

	if result.ExistingCount != 2 {
		t.Errorf("expected 2 existing, got %d", result.ExistingCount)
	}
}

func TestMergeFiles_WithMissing(t *testing.T) {
	state := &State{
		Directory: "/test",
		Classifications: []Classification{
			{File: "old1.mp4", GroupID: "g1", TakeNumber: 1},
			{File: "old2.mp4", GroupID: "g2", TakeNumber: 1},
		},
	}

	scannedFiles := []string{"new1.mp4"}

	result := MergeFiles(state, scannedFiles)

	if len(result.MissingFiles) != 2 {
		t.Errorf("expected 2 missing files, got %d", len(result.MissingFiles))
	}

	expectedMissing := map[string]bool{"old1.mp4": true, "old2.mp4": true}
	for _, f := range result.MissingFiles {
		if !expectedMissing[f] {
			t.Errorf("unexpected missing file: %s", f)
		}
	}
}
