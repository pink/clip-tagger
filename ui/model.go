// ui/model.go
package ui

import (
	"clip-tagger/scanner"
	"clip-tagger/state"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Screen represents the different screens in the application
type Screen int

const (
	ScreenStartup Screen = iota
	ScreenClassification
	ScreenGroupSelection
	ScreenGroupInsertion
	ScreenReview
	ScreenComplete
)

// String returns the string representation of a Screen
func (s Screen) String() string {
	switch s {
	case ScreenStartup:
		return "startup"
	case ScreenClassification:
		return "classification"
	case ScreenGroupSelection:
		return "group_selection"
	case ScreenGroupInsertion:
		return "group_insertion"
	case ScreenReview:
		return "review"
	case ScreenComplete:
		return "complete"
	default:
		return "unknown"
	}
}

// Model is the main Bubbletea model
type Model struct {
	state         *state.State
	currentScreen Screen
	err           string
	directory     string
	startupData   *StartupData
}

// NewModel creates a new Model with the given state and directory
func NewModel(appState *state.State, directory string) Model {
	return Model{
		state:         appState,
		currentScreen: ScreenStartup,
		err:           "",
		directory:     directory,
		startupData:   nil,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	// Return a command to initialize the startup screen
	return func() tea.Msg {
		// Scan directory for video files
		scan := scanner.NewScanner(m.directory)
		result, err := scan.Scan(scanner.SortBy(m.state.SortBy))
		if err != nil {
			return ErrorMsg{Err: fmt.Sprintf("Failed to scan directory: %v", err)}
		}

		// Extract filenames from scan result
		scannedFiles := make([]string, len(result.Files))
		for i, f := range result.Files {
			scannedFiles[i] = f.Name
		}

		// If state has classifications, merge with scanned files
		var mergeResult *state.MergeResult
		if len(m.state.Classifications) > 0 {
			mergeResult = state.MergeFiles(m.state, scannedFiles)
		}

		return StartupInitialized{
			ScannedFiles: scannedFiles,
			MergeResult:  mergeResult,
		}
	}
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle startup screen keys
		if m.currentScreen == ScreenStartup && m.startupData != nil {
			var keyMsg string
			switch msg.Type {
			case tea.KeyEnter:
				keyMsg = "enter"
			case tea.KeyCtrlC:
				keyMsg = "ctrl+c"
			default:
				keyMsg = msg.String()
			}

			screen := StartupUpdate(m.startupData, keyMsg)
			if screen == -1 {
				return m, tea.Quit
			} else if screen >= 0 {
				m.currentScreen = screen
				return m, nil
			}
			// screen == -2 means no action, continue
		}

		// Global quit handler
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}

	case StartupInitialized:
		m.startupData = NewStartupData(m.state, msg.ScannedFiles, msg.MergeResult)
		return m, nil

	case TransitionToScreen:
		m.currentScreen = msg.Screen
		return m, nil

	case StateUpdate:
		m.state = msg.State
		return m, nil

	case ErrorMsg:
		m.err = msg.Err
		return m, nil
	}

	return m, nil
}

// View renders the current screen
func (m Model) View() string {
	if m.err != "" {
		return fmt.Sprintf("Error: %s\n\nPress Ctrl+C to quit", m.err)
	}

	// Render screens based on current screen
	switch m.currentScreen {
	case ScreenStartup:
		if m.startupData != nil {
			return StartupView(m.startupData)
		}
		return "Loading...\n\nPress Ctrl+C to quit"
	case ScreenClassification:
		return "Classification Screen\n\nPress Ctrl+C to quit"
	case ScreenGroupSelection:
		return "Group Selection Screen\n\nPress Ctrl+C to quit"
	case ScreenGroupInsertion:
		return "Group Insertion Screen\n\nPress Ctrl+C to quit"
	case ScreenReview:
		return "Review Screen\n\nPress Ctrl+C to quit"
	case ScreenComplete:
		return "Complete Screen\n\nPress Ctrl+C to quit"
	default:
		return "Unknown Screen\n\nPress Ctrl+C to quit"
	}
}
