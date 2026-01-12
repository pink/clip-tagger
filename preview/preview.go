// Package preview provides functionality to open files in their system default application.
// It supports macOS (using 'open'), Linux (using 'xdg-open'), and Windows (using 'start').
// The preview command runs in the background without blocking the UI.
package preview

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// OpenFile opens a file in the system's default application
// The command runs in the background and does not block
func OpenFile(filePath string) error {
	// Validate input
	if filePath == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	// Get the appropriate command for the current OS
	cmdName, args := getPreviewCommand(runtime.GOOS, filePath)

	// Create the command
	cmd := exec.Command(cmdName, args...)

	// Start the command in the background (don't wait for it to finish)
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	// Don't wait for the command to finish - let it run in background
	// This ensures the UI doesn't block while the video player is open
	go func() {
		// Wait in a goroutine to clean up process resources
		_ = cmd.Wait()
	}()

	return nil
}

// getPreviewCommand returns the command and arguments needed to open a file
// based on the operating system
func getPreviewCommand(osName, filePath string) (string, []string) {
	switch osName {
	case "darwin": // macOS
		return "open", []string{filePath}
	case "linux":
		return "xdg-open", []string{filePath}
	case "windows":
		// Windows requires special handling with cmd /c start
		// The empty string "" is required as the title argument for start command
		return "cmd", []string{"/c", "start", "", filePath}
	default:
		// Default to xdg-open for unknown platforms
		return "xdg-open", []string{filePath}
	}
}
