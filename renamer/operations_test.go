// renamer/operations_test.go
package renamer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRenameInPlace(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file
	src := filepath.Join(tmpDir, "source.mp4")
	if err := os.WriteFile(src, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(tmpDir, "[01_01] intro.mp4")
	rename := Rename{OriginalPath: src, TargetPath: dst}

	err := RenameInPlace([]Rename{rename})
	if err != nil {
		t.Fatalf("rename failed: %v", err)
	}

	// Check source gone, target exists
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("source still exists")
	}

	if _, err := os.Stat(dst); err != nil {
		t.Error("target doesn't exist")
	}

	// Verify content
	content, _ := os.ReadFile(dst)
	if string(content) != "content" {
		t.Error("content mismatch")
	}
}

func TestCopyToDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")

	// Create source file
	src := filepath.Join(tmpDir, "source.mp4")
	if err := os.WriteFile(src, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(outputDir, "[01_01] intro.mp4")
	rename := Rename{OriginalPath: src, TargetPath: dst}

	err := CopyToDirectory([]Rename{rename}, outputDir)
	if err != nil {
		t.Fatalf("copy failed: %v", err)
	}

	// Check source still exists
	if _, err := os.Stat(src); err != nil {
		t.Error("source was removed")
	}

	// Check target exists
	if _, err := os.Stat(dst); err != nil {
		t.Error("target doesn't exist")
	}

	// Verify content
	content, _ := os.ReadFile(dst)
	if string(content) != "content" {
		t.Error("content mismatch")
	}
}
