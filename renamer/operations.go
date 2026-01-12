// renamer/operations.go
package renamer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// RenameInPlace renames files in their current directory
func RenameInPlace(renames []Rename) error {
	for _, r := range renames {
		// Skip if no actual change
		if r.OriginalPath == r.TargetPath {
			continue
		}

		if err := os.Rename(r.OriginalPath, r.TargetPath); err != nil {
			return fmt.Errorf("rename %s -> %s: %w",
				filepath.Base(r.OriginalPath),
				filepath.Base(r.TargetPath),
				err)
		}
	}
	return nil
}

// CopyToDirectory copies files to a new directory
func CopyToDirectory(renames []Rename, outputDir string) error {
	// Create output directory if needed
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	for _, r := range renames {
		targetPath := filepath.Join(outputDir, filepath.Base(r.TargetPath))

		if err := copyFile(r.OriginalPath, targetPath); err != nil {
			return fmt.Errorf("copy %s -> %s: %w",
				filepath.Base(r.OriginalPath),
				filepath.Base(targetPath),
				err)
		}
	}
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return dstFile.Sync()
}
