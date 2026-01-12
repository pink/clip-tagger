// ui/messages.go
package ui

import "clip-tagger/state"

// TransitionToScreen is a message to transition to a different screen
type TransitionToScreen struct {
	Screen Screen
}

// StateUpdate is a message to update the application state
type StateUpdate struct {
	State *state.State
}

// ErrorMsg is a message to display an error
type ErrorMsg struct {
	Err string
}

// StartupInitialized is sent when startup screen is initialized
type StartupInitialized struct {
	ScannedFiles []string
	MergeResult  *state.MergeResult
}
