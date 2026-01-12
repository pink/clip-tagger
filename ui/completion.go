// ui/completion.go
package ui

import (
	"clip-tagger/renamer"
	"clip-tagger/state"
	"fmt"
	"path/filepath"
	"time"
)

// CompletionMode represents the rename mode
type CompletionMode int

const (
	CompletionModeRenameInPlace CompletionMode = iota
	CompletionModeCopyToDirectory
)

// CompletionData contains the data needed to render the completion screen
type CompletionData struct {
	Renames         []renamer.Rename
	Conflicts       []renamer.Rename
	HasConflicts    bool
	SelectedMode    int
	OutputDirectory string
	ExecutionResult *CompletionExecutionResult
}

// CompletionExecutionResult contains the result of executing rename operations
type CompletionExecutionResult struct {
	Success      bool
	FilesChanged int
	Mode         string
	Error        error
}

// CompletionUpdateResult contains the result of a completion update
type CompletionUpdateResult struct {
	Screen Screen // -1 for quit, -2 for no screen change, >= 0 for screen transition
}

// NewCompletionData creates completion data from state
func NewCompletionData(appState *state.State) *CompletionData {
	// Build list of rename operations from classifications
	var renames []renamer.Rename
	for _, classification := range appState.Classifications {
		group := appState.FindGroupByID(classification.GroupID)
		if group == nil {
			// Skip if group not found (shouldn't happen)
			continue
		}

		originalPath := filepath.Join(appState.Directory, classification.File)
		targetPath := renamer.GenerateTargetPath(
			appState.Directory,
			originalPath,
			group.Order,
			classification.TakeNumber,
			group.Name,
		)

		renames = append(renames, renamer.Rename{
			OriginalPath: originalPath,
			TargetPath:   targetPath,
		})
	}

	// Detect conflicts
	conflicts := renamer.DetectConflicts(renames)

	// Generate output directory name with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	outputDir := filepath.Join(appState.Directory, fmt.Sprintf("renamed_%s", timestamp))

	return &CompletionData{
		Renames:         renames,
		Conflicts:       conflicts,
		HasConflicts:    len(conflicts) > 0,
		SelectedMode:    0, // Default to rename in place
		OutputDirectory: outputDir,
		ExecutionResult: nil,
	}
}

// CompletionView renders the completion screen
func CompletionView(data *CompletionData) string {
	var output string

	// If execution has completed, show result
	if data.ExecutionResult != nil {
		return renderExecutionResult(data.ExecutionResult)
	}

	// Header
	output += "=== Mode Selection ===\n\n"

	// Show mode options
	modes := []string{
		"1. Rename in place",
		"2. Copy to new directory",
	}

	for i, mode := range modes {
		indicator := "  "
		if i == data.SelectedMode {
			indicator = "> "
		}
		output += indicator + mode

		// Show output directory for copy mode
		if i == 1 {
			output += fmt.Sprintf("\n     Output: %s", filepath.Base(data.OutputDirectory))
		}

		output += "\n"
	}

	output += "\n"

	// Show conflict warning if any
	if data.HasConflicts {
		output += "WARNING: Conflicts Detected!\n"
		output += fmt.Sprintf("%d file(s) would overwrite existing files:\n\n", len(data.Conflicts))

		// Show up to 5 conflicts
		maxShow := 5
		for i, conflict := range data.Conflicts {
			if i >= maxShow {
				remaining := len(data.Conflicts) - maxShow
				output += fmt.Sprintf("  ... and %d more\n", remaining)
				break
			}
			output += fmt.Sprintf("  - %s -> %s\n",
				filepath.Base(conflict.OriginalPath),
				filepath.Base(conflict.TargetPath))
		}

		output += "\n"
		output += "These files will be OVERWRITTEN if you proceed.\n\n"
	}

	// Instructions
	output += "Controls:\n"
	output += "  Up/Down - Select mode\n"
	output += "  Enter - Execute operation\n"
	if data.HasConflicts {
		output += "  Esc - Go back (abort)\n"
	}
	output += "  q - Quit\n"

	return output
}

// renderExecutionResult renders the execution result screen
func renderExecutionResult(result *CompletionExecutionResult) string {
	var output string

	if result.Success {
		output += "=== Success! ===\n\n"
		output += fmt.Sprintf("Operation: %s\n", result.Mode)
		output += fmt.Sprintf("Files changed: %d\n\n", result.FilesChanged)
		output += "All files have been successfully renamed.\n\n"
	} else {
		output += "=== Error ===\n\n"
		output += "Failed to complete operation.\n\n"
		if result.Error != nil {
			output += fmt.Sprintf("Error: %v\n\n", result.Error)
		}
	}

	output += "Press any key to exit\n"

	return output
}

// CompletionUpdate handles input for the completion screen
func CompletionUpdate(data *CompletionData, msg string) CompletionUpdateResult {
	// If execution has completed, any key press quits
	if data.ExecutionResult != nil {
		return CompletionUpdateResult{Screen: -1}
	}

	switch msg {
	case "up":
		// Move selection up
		if data.SelectedMode > 0 {
			data.SelectedMode--
		}
		return CompletionUpdateResult{Screen: -2}

	case "down":
		// Move selection down
		if data.SelectedMode < 1 {
			data.SelectedMode++
		}
		return CompletionUpdateResult{Screen: -2}

	case "enter":
		// Execute the selected operation
		executeOperation(data)
		return CompletionUpdateResult{Screen: -2}

	case "esc":
		// Go back to review screen
		return CompletionUpdateResult{Screen: ScreenReview}

	case "q", "ctrl+c":
		// Quit
		return CompletionUpdateResult{Screen: -1}

	default:
		// No action for unrecognized keys
		return CompletionUpdateResult{Screen: -2}
	}
}

// executeOperation executes the selected rename operation
func executeOperation(data *CompletionData) {
	var err error
	var mode string
	filesChanged := 0

	// Count actual files that will be changed (exclude no-ops)
	for _, r := range data.Renames {
		if r.OriginalPath != r.TargetPath {
			filesChanged++
		}
	}

	if data.SelectedMode == int(CompletionModeRenameInPlace) {
		mode = "Rename in place"
		err = renamer.RenameInPlace(data.Renames)
	} else {
		mode = "Copy to new directory"
		err = renamer.CopyToDirectory(data.Renames, data.OutputDirectory)
	}

	data.ExecutionResult = &CompletionExecutionResult{
		Success:      err == nil,
		FilesChanged: filesChanged,
		Mode:         mode,
		Error:        err,
	}
}
