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
	output += RenderHeader("=== Mode Selection ===") + "\n\n"

	// Show mode options
	modes := []string{
		"1. Rename in place",
		"2. Copy to new directory",
	}

	for i, mode := range modes {
		if i == data.SelectedMode {
			output += RenderCursor("> ") + RenderHighlight(mode)
		} else {
			output += "  " + mode
		}

		// Show output directory for copy mode
		if i == 1 {
			output += fmt.Sprintf("\n     %s %s",
				RenderMuted("Output:"),
				RenderMuted(filepath.Base(data.OutputDirectory)))
		}

		output += "\n"
	}

	output += "\n"

	// Show conflict warning if any
	if data.HasConflicts {
		output += RenderDanger("WARNING: Conflicts Detected!") + "\n"
		output += RenderWarning(fmt.Sprintf("%d file(s) would overwrite existing files:", len(data.Conflicts))) + "\n\n"

		// Show up to 5 conflicts
		maxShow := 5
		for i, conflict := range data.Conflicts {
			if i >= maxShow {
				remaining := len(data.Conflicts) - maxShow
				output += RenderWarning(fmt.Sprintf("  ... and %d more", remaining)) + "\n"
				break
			}
			output += fmt.Sprintf("  %s %s %s %s\n",
				RenderMuted("-"),
				RenderMuted(filepath.Base(conflict.OriginalPath)),
				RenderMuted("->"),
				filepath.Base(conflict.TargetPath))
		}

		output += "\n"
		output += RenderDanger("These files will be OVERWRITTEN if you proceed.") + "\n\n"
	}

	// Instructions
	output += RenderMuted("Controls:") + "\n"
	output += RenderKeyHint("  Up/Down - Select mode") + "\n"
	output += RenderKeyHint("  Enter - Execute operation") + "\n"
	if data.HasConflicts {
		output += RenderKeyHint("  Esc - Go back (abort)") + "\n"
	}
	output += RenderKeyHint("  q - Quit") + "\n"

	return output
}

// renderExecutionResult renders the execution result screen
func renderExecutionResult(result *CompletionExecutionResult) string {
	var output string

	if result.Success {
		output += RenderSuccess("=== Success! ===") + "\n\n"
		output += fmt.Sprintf("%s %s\n", RenderMuted("Operation:"), result.Mode)
		output += fmt.Sprintf("%s %s\n\n",
			RenderMuted("Files changed:"),
			RenderSuccess(fmt.Sprintf("%d", result.FilesChanged)))
		output += RenderSuccess("All files have been successfully renamed.") + "\n\n"
	} else {
		output += RenderDanger("=== Error ===") + "\n\n"
		output += RenderDanger("Failed to complete operation.") + "\n\n"
		if result.Error != nil {
			output += fmt.Sprintf("%s %v\n\n", RenderMuted("Error:"), result.Error)
		}
	}

	output += RenderKeyHint("Press any key to exit") + "\n"

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

// updateStateAfterRename updates state Classifications to use new filenames after successful rename
func updateStateAfterRename(appState *state.State, renames []renamer.Rename, mode string, outputDir string) {
	// Build mapping from old filename to new filename
	filenameMap := make(map[string]string)
	for _, r := range renames {
		oldFilename := filepath.Base(r.OriginalPath)
		newFilename := filepath.Base(r.TargetPath)
		if oldFilename != newFilename {
			filenameMap[oldFilename] = newFilename
		}
	}

	// Update all Classifications to use new filenames
	for i := range appState.Classifications {
		if newFilename, exists := filenameMap[appState.Classifications[i].File]; exists {
			appState.Classifications[i].File = newFilename
		}
	}

	// If copy to directory mode, update the directory path
	if mode == "Copy to new directory" {
		appState.Directory = outputDir
	}
}
