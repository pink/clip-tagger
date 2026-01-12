package main

import (
	"clip-tagger/flags"
	"clip-tagger/renamer"
	"clip-tagger/state"
	"clip-tagger/ui"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Parse command-line flags
	config, err := flags.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		flags.PrintUsage()
		os.Exit(1)
	}

	directory := config.Directory

	// Validate directory exists
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: directory '%s' does not exist\n", directory)
		os.Exit(1)
	}

	// Handle --reset flag: delete state file
	if config.Reset {
		statePath := state.StateFilePath(directory)
		if err := os.Remove(statePath); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error deleting state file: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("State reset successfully")
		// If only reset was requested, exit
		if !config.CleanMissing && !config.Preview {
			return
		}
	}

	// Determine sort order
	sortBy := state.SortByModifiedTime // default
	if config.SortBy != "" {
		switch config.SortBy {
		case "name":
			sortBy = state.SortByName
		case "modified":
			sortBy = state.SortByModifiedTime
		case "created":
			sortBy = state.SortByCreatedTime
		}
	}

	// Initialize or load state
	var appState *state.State
	statePath := state.StateFilePath(directory)
	if state.StateExists(directory) && !config.Reset {
		appState, err = state.Load(statePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading state: %v\n", err)
			os.Exit(1)
		}
		// Override sort order if flag is set
		if config.SortBy != "" {
			appState.SortBy = sortBy
		}
	} else {
		appState = state.NewState(directory, sortBy)
	}

	// Handle --clean-missing flag: remove files that no longer exist
	if config.CleanMissing {
		cleanedCount := cleanMissingFiles(appState)
		fmt.Printf("Cleaned %d missing file(s) from state\n", cleanedCount)
		if cleanedCount > 0 {
			if err := appState.Save(statePath); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving state: %v\n", err)
				os.Exit(1)
			}
		}
		// If only clean-missing was requested, exit
		if !config.Preview {
			return
		}
	}

	// Handle --preview flag: show what would be renamed
	if config.Preview {
		showPreview(appState)
		return
	}

	// Create and run the Bubbletea program
	model := ui.NewModel(appState, directory)
	program := tea.NewProgram(model)

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}

// cleanMissingFiles removes classifications for files that no longer exist
func cleanMissingFiles(appState *state.State) int {
	cleanedCount := 0
	newClassifications := []state.Classification{}

	for _, c := range appState.Classifications {
		filePath := filepath.Join(appState.Directory, c.File)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			cleanedCount++
			continue
		}
		newClassifications = append(newClassifications, c)
	}

	appState.Classifications = newClassifications
	return cleanedCount
}

// showPreview displays what would be renamed without executing
func showPreview(appState *state.State) {
	if len(appState.Classifications) == 0 {
		fmt.Println("No classifications to preview")
		return
	}

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

	// Display preview
	fmt.Println("=== Rename Preview ===")
	fmt.Printf("\nTotal files to rename: %d\n\n", len(renames))

	for _, r := range renames {
		// Skip if no actual change
		if r.OriginalPath == r.TargetPath {
			continue
		}

		fmt.Printf("  %s\n", filepath.Base(r.OriginalPath))
		fmt.Printf("  -> %s\n\n", filepath.Base(r.TargetPath))
	}

	// Show conflicts if any
	if len(conflicts) > 0 {
		fmt.Println("\n=== WARNING: Conflicts Detected ===")
		fmt.Printf("%d file(s) would overwrite existing files:\n\n", len(conflicts))

		for _, conflict := range conflicts {
			fmt.Printf("  %s -> %s (CONFLICT)\n",
				filepath.Base(conflict.OriginalPath),
				filepath.Base(conflict.TargetPath))
		}
		fmt.Println()
	}

	fmt.Println("Use the UI to execute these renames")
}
