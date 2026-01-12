// ui/model.go
package ui

import (
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
}

// NewModel creates a new Model with the given state
func NewModel(appState *state.State) Model {
	return Model{
		state:         appState,
		currentScreen: ScreenStartup,
		err:           "",
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	// Return a command to transition to startup screen
	// For now, we just return nil as screens aren't implemented yet
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}

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

	// Basic placeholder views for each screen
	// Actual screen implementations will come in later tasks
	switch m.currentScreen {
	case ScreenStartup:
		return "Startup Screen\n\nPress Ctrl+C to quit"
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
