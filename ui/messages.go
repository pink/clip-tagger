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

// ClassificationInitialized is sent when classification screen is initialized
type ClassificationInitialized struct {
	Files     []string
	FileIndex int
}

// GroupSelectionInitialized is sent when group selection screen is initialized
type GroupSelectionInitialized struct {
	CurrentFile string
}

// GroupSelected is sent when a group is selected from the group selection screen
type GroupSelected struct {
	GroupID   string
	GroupName string
}
