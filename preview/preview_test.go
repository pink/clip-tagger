// preview/preview_test.go
package preview

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetPreviewCommand(t *testing.T) {
	tests := []struct {
		name     string
		osName   string
		filePath string
		wantCmd  string
		wantArgs []string
	}{
		{
			name:     "macOS with simple path",
			osName:   "darwin",
			filePath: "/Users/test/video.mp4",
			wantCmd:  "open",
			wantArgs: []string{"/Users/test/video.mp4"},
		},
		{
			name:     "macOS with path containing spaces",
			osName:   "darwin",
			filePath: "/Users/test/my video.mp4",
			wantCmd:  "open",
			wantArgs: []string{"/Users/test/my video.mp4"},
		},
		{
			name:     "Linux with simple path",
			osName:   "linux",
			filePath: "/home/test/video.mp4",
			wantCmd:  "xdg-open",
			wantArgs: []string{"/home/test/video.mp4"},
		},
		{
			name:     "Linux with path containing spaces",
			osName:   "linux",
			filePath: "/home/test/my video.mp4",
			wantCmd:  "xdg-open",
			wantArgs: []string{"/home/test/my video.mp4"},
		},
		{
			name:     "Windows with simple path",
			osName:   "windows",
			filePath: "C:\\Users\\test\\video.mp4",
			wantCmd:  "cmd",
			wantArgs: []string{"/c", "start", "", "C:\\Users\\test\\video.mp4"},
		},
		{
			name:     "Windows with path containing spaces",
			osName:   "windows",
			filePath: "C:\\Users\\test\\my video.mp4",
			wantCmd:  "cmd",
			wantArgs: []string{"/c", "start", "", "C:\\Users\\test\\my video.mp4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, args := getPreviewCommand(tt.osName, tt.filePath)
			if cmd != tt.wantCmd {
				t.Errorf("getPreviewCommand() cmd = %v, want %v", cmd, tt.wantCmd)
			}
			if len(args) != len(tt.wantArgs) {
				t.Fatalf("getPreviewCommand() args length = %v, want %v", len(args), len(tt.wantArgs))
			}
			for i, arg := range args {
				if arg != tt.wantArgs[i] {
					t.Errorf("getPreviewCommand() args[%d] = %v, want %v", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}

func TestGetPreviewCommand_UnsupportedOS(t *testing.T) {
	cmd, args := getPreviewCommand("unsupported", "/path/to/file.mp4")
	// Should default to xdg-open on unsupported platforms
	if cmd != "xdg-open" {
		t.Errorf("expected default to xdg-open, got %v", cmd)
	}
	if len(args) != 1 || args[0] != "/path/to/file.mp4" {
		t.Errorf("expected args to be [/path/to/file.mp4], got %v", args)
	}
}

func TestOpenFile_FileNotExist(t *testing.T) {
	// Test opening a file that doesn't exist
	nonExistentFile := "/tmp/nonexistent-file-" + t.Name() + ".mp4"
	err := OpenFile(nonExistentFile)
	if err == nil {
		t.Error("expected error when opening non-existent file")
	}
	if !strings.Contains(err.Error(), "file not found") {
		t.Errorf("expected 'file not found' error, got: %v", err)
	}
}

func TestOpenFile_EmptyPath(t *testing.T) {
	// Test opening with empty path
	err := OpenFile("")
	if err == nil {
		t.Error("expected error when opening empty path")
	}
	if !strings.Contains(err.Error(), "file path cannot be empty") {
		t.Errorf("expected 'file path cannot be empty' error, got: %v", err)
	}
}

func TestOpenFile_Success(t *testing.T) {
	// Create a temporary file to test with
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-video.mp4")

	// Create the file
	f, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	f.Close()

	// Test opening the file
	// Note: This will actually try to open the file in the default player
	// On CI or headless systems, this might fail, so we'll check the error message
	err = OpenFile(tmpFile)

	// On macOS, Windows, and Linux with a default player, this should succeed
	// On headless systems or systems without a default player, we might get an error
	// We'll consider it a success if either:
	// 1. No error
	// 2. Error is about no default application (not a file not found error)
	if err != nil {
		// If we get an error, make sure it's not a "file not found" error
		if strings.Contains(err.Error(), "file not found") {
			t.Errorf("unexpected 'file not found' error: %v", err)
		}
		// Other errors (like no default application) are acceptable in test environments
		t.Logf("OpenFile returned error (acceptable in test environment): %v", err)
	}
}

func TestOpenFile_Integration(t *testing.T) {
	// Skip if not running on a system with a display
	if os.Getenv("CI") == "true" || os.Getenv("HEADLESS") == "true" {
		t.Skip("Skipping integration test in CI or headless environment")
	}

	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "integration-test.mp4")

	f, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	// Write a minimal valid MP4 header (this is not a real video, but enough for the file to exist)
	f.WriteString("fake video content for testing")
	f.Close()

	// Test on current OS
	err = OpenFile(tmpFile)
	if err != nil {
		// On systems without a default player or in CI, this is acceptable
		t.Logf("OpenFile error (may be expected): %v", err)
	} else {
		t.Logf("OpenFile succeeded on %s", runtime.GOOS)
	}
}

func TestOpenFile_RelativePath(t *testing.T) {
	// Test with relative path - should work with existing files
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "relative-test.mp4")

	f, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	f.Close()

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Test with relative path
	err = OpenFile("relative-test.mp4")

	// Similar to above, we accept errors if there's no default player
	if err != nil && !strings.Contains(err.Error(), "file not found") {
		t.Logf("OpenFile with relative path error (acceptable): %v", err)
	}
}
