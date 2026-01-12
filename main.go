package main

import (
	"clip-tagger/state"
	"clip-tagger/ui"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: clip-tagger <directory>")
		os.Exit(1)
	}

	directory := os.Args[1]

	// Validate directory exists
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		fmt.Printf("Error: directory '%s' does not exist\n", directory)
		os.Exit(1)
	}

	// Initialize state
	appState := state.NewState(directory, state.SortByModifiedTime)

	// Create and run the Bubbletea program
	model := ui.NewModel(appState)
	program := tea.NewProgram(model)

	if _, err := program.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
