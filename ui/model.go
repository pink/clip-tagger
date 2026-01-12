// ui/model.go
package ui

import (
	"clip-tagger/preview"
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
	state              *state.State
	currentScreen      Screen
	err                string
	directory          string
	startupData        *StartupData
	classificationData *ClassificationData
	groupSelectionData *GroupSelectionData
	groupInsertionData *GroupInsertionData
	reviewData         *ReviewData
	completionData     *CompletionData
	files              []string // List of files being classified
	currentFileIndex   int      // Current file index in files list
	actionCounter      int      // Counter for periodic auto-saves
	actionsPerSave     int      // Number of actions before auto-save (default: 5)
}

// NewModel creates a new Model with the given state and directory
func NewModel(appState *state.State, directory string) Model {
	return Model{
		state:              appState,
		currentScreen:      ScreenStartup,
		err:                "",
		directory:          directory,
		startupData:        nil,
		classificationData: nil,
		groupSelectionData: nil,
		groupInsertionData: nil,
		reviewData:         nil,
		completionData:     nil,
		files:              []string{},
		currentFileIndex:   0,
		actionCounter:      0,
		actionsPerSave:     5, // Default: save every 5 actions
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
				// If transitioning to classification screen, initialize it
				if screen == ScreenClassification {
					return m, func() tea.Msg {
						return ClassificationInitialized{
							Files:     m.files,
							FileIndex: m.currentFileIndex,
						}
					}
				}
				return m, nil
			}
			// screen == -2 means no action, continue
		}

		// Handle classification screen keys
		if m.currentScreen == ScreenClassification && m.classificationData != nil {
			var keyMsg string
			switch msg.Type {
			case tea.KeyCtrlC:
				keyMsg = "ctrl+c"
			default:
				keyMsg = msg.String()
			}

			result := ClassificationUpdate(m.classificationData, keyMsg)
			if result.Screen == -1 {
				return m, tea.Quit
			} else if result.Screen >= 0 {
				// Save state when leaving classification screen
				if result.Screen != ScreenClassification {
					m = m.autoSaveState()
				}
				m.currentScreen = result.Screen
				// If transitioning to group selection, initialize it
				if result.Screen == ScreenGroupSelection {
					return m, func() tea.Msg {
						return GroupSelectionInitialized{
							CurrentFile: m.classificationData.CurrentFile,
						}
					}
				}
				// If transitioning to group insertion, initialize it
				if result.Screen == ScreenGroupInsertion {
					return m, func() tea.Msg {
						return GroupInsertionInitialized{
							CurrentFile: m.classificationData.CurrentFile,
						}
					}
				}
				return m, nil
			}
			// result.Screen == -2 means no screen change
			// Handle actions that don't change screens
			if result.Action == ClassificationActionPreview {
				// Handle preview action
				err := preview.OpenFile(m.classificationData.FilePath)
				if err != nil {
					m.err = fmt.Sprintf("Failed to preview file: %v", err)
				}
				return m, nil
			}
			// Handle "Same as Last" action
			if result.Action == ClassificationActionSameAsLast {
				m = m.handleClassificationSameAsLast()
				m = m.incrementActionAndMaybeSave()
				return m, nil
			}
			// Handle "Skip" action
			if result.Action == ClassificationActionSkip {
				m = m.handleClassificationSkip()
				m = m.incrementActionAndMaybeSave()
				return m, nil
			}
		}

		// Handle group selection screen keys
		if m.currentScreen == ScreenGroupSelection && m.groupSelectionData != nil {
			var keyMsg string
			switch msg.Type {
			case tea.KeyCtrlC:
				keyMsg = "ctrl+c"
			case tea.KeyEnter:
				keyMsg = "enter"
			case tea.KeyEsc:
				keyMsg = "esc"
			case tea.KeyUp:
				keyMsg = "up"
			case tea.KeyDown:
				keyMsg = "down"
			case tea.KeyBackspace:
				keyMsg = "backspace"
			case tea.KeySpace:
				keyMsg = " "
			default:
				keyMsg = msg.String()
			}

			result := GroupSelectionUpdate(m.groupSelectionData, keyMsg)
			if result.Screen == -1 {
				return m, tea.Quit
			} else if result.Screen >= 0 {
				m.currentScreen = result.Screen
				// If a group was selected, send GroupSelected message
				if result.SelectedGroupID != "" {
					return m, func() tea.Msg {
						return GroupSelected{
							GroupID:   result.SelectedGroupID,
							GroupName: result.SelectedGroupName,
						}
					}
				}
				return m, nil
			}
			// result.Screen == -2 means no screen change, continue
		}

		// Handle group insertion screen keys
		if m.currentScreen == ScreenGroupInsertion && m.groupInsertionData != nil {
			var keyMsg string
			switch msg.Type {
			case tea.KeyCtrlC:
				keyMsg = "ctrl+c"
			case tea.KeyEnter:
				keyMsg = "enter"
			case tea.KeyEsc:
				keyMsg = "esc"
			case tea.KeyUp:
				keyMsg = "up"
			case tea.KeyDown:
				keyMsg = "down"
			case tea.KeyBackspace:
				keyMsg = "backspace"
			case tea.KeySpace:
				keyMsg = " "
			default:
				keyMsg = msg.String()
			}

			result := GroupInsertionUpdate(m.groupInsertionData, keyMsg)
			if result.Screen == -1 {
				return m, tea.Quit
			} else if result.Screen >= 0 {
				m.currentScreen = result.Screen
				// If a group was inserted, send GroupInserted message
				if result.InsertedGroupID != "" {
					return m, func() tea.Msg {
						return GroupInserted{
							GroupID:   result.InsertedGroupID,
							GroupName: result.InsertedGroupName,
							Order:     result.InsertedOrder,
						}
					}
				}
				return m, nil
			}
			// result.Screen == -2 means no screen change, continue
		}

		// Handle review screen keys
		if m.currentScreen == ScreenReview && m.reviewData != nil {
			var keyMsg string
			switch msg.Type {
			case tea.KeyCtrlC:
				keyMsg = "ctrl+c"
			case tea.KeyEnter:
				keyMsg = "enter"
			case tea.KeyEsc:
				keyMsg = "esc"
			case tea.KeyUp:
				keyMsg = "up"
			case tea.KeyDown:
				keyMsg = "down"
			default:
				keyMsg = msg.String()
			}

			result := ReviewUpdate(m.reviewData, keyMsg)
			if result.Screen == -1 {
				return m, tea.Quit
			} else if result.Screen >= 0 {
				m.currentScreen = result.Screen
				// If going back to classification, need to reinitialize classification data
				if result.Screen == ScreenClassification {
					// Reset to first file to allow re-classification
					m.currentFileIndex = 0
					return m, func() tea.Msg {
						return ClassificationInitialized{
							Files:     m.files,
							FileIndex: 0,
						}
					}
				}
				// If transitioning to completion screen, initialize it
				if result.Screen == ScreenComplete {
					return m, func() tea.Msg {
						return CompletionInitialized{}
					}
				}
				return m, nil
			}
			// result.Screen == -2 means no screen change, continue
		}

		// Handle completion screen keys
		if m.currentScreen == ScreenComplete && m.completionData != nil {
			var keyMsg string
			switch msg.Type {
			case tea.KeyCtrlC:
				keyMsg = "ctrl+c"
			case tea.KeyEnter:
				keyMsg = "enter"
			case tea.KeyEsc:
				keyMsg = "esc"
			case tea.KeyUp:
				keyMsg = "up"
			case tea.KeyDown:
				keyMsg = "down"
			default:
				keyMsg = msg.String()
			}

			result := CompletionUpdate(m.completionData, keyMsg)
			if result.Screen == -1 {
				return m, tea.Quit
			} else if result.Screen >= 0 {
				m.currentScreen = result.Screen
				return m, nil
			}
			// result.Screen == -2 means no screen change, continue
		}

		// Global quit handler
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}

	case StartupInitialized:
		m.startupData = NewStartupData(m.state, msg.ScannedFiles, msg.MergeResult)
		// Store files for classification
		m.files = msg.ScannedFiles
		m.currentFileIndex = 0
		return m, nil

	case ClassificationInitialized:
		m.classificationData = NewClassificationData(m.state, msg.Files, msg.FileIndex)
		return m, nil

	case GroupSelectionInitialized:
		m.groupSelectionData = NewGroupSelectionData(m.state, msg.CurrentFile)
		return m, nil

	case GroupSelected:
		// Group was selected, handle classification
		m = m.handleGroupSelected(msg.GroupID)
		// Auto-save state after group selection (immediate save)
		m = m.autoSaveState()
		return m, nil

	case GroupInsertionInitialized:
		m.groupInsertionData = NewGroupInsertionData(m.state, msg.CurrentFile)
		return m, nil

	case GroupInserted:
		// Group was inserted, add to state and handle classification
		// Create the group with the specified order
		newGroup := state.Group{
			ID:    msg.GroupID,
			Name:  msg.GroupName,
			Order: msg.Order,
		}

		// Insert group at the correct position and renumber
		insertGroupAtPosition(&m.state.Groups, newGroup, msg.Order)

		// Handle classification with the new group
		m = m.handleGroupInserted(msg.GroupID, msg.GroupName, msg.Order)

		// Auto-save state after group insertion (immediate save)
		m = m.autoSaveState()

		return m, nil

	case CompletionInitialized:
		m.completionData = NewCompletionData(m.state)
		return m, nil

	case TransitionToScreen:
		// Save state when leaving classification screen
		if m.currentScreen == ScreenClassification && msg.Screen != ScreenClassification {
			m = m.autoSaveState()
		}
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

// insertGroupAtPosition inserts a group at the specified order position and renumbers all groups
func insertGroupAtPosition(groups *[]state.Group, newGroup state.Group, order int) {
	// Find the insertion index based on order
	insertIndex := 0
	for i, g := range *groups {
		if g.Order >= order {
			insertIndex = i
			break
		}
		insertIndex = i + 1
	}

	// Insert the group at the correct position
	*groups = append(*groups, state.Group{})
	copy((*groups)[insertIndex+1:], (*groups)[insertIndex:])
	(*groups)[insertIndex] = newGroup

	// Renumber all groups to maintain sequential order
	for i := range *groups {
		(*groups)[i].Order = i + 1
	}
}

// autoSaveState saves the current state to disk and handles errors gracefully
// This is called after key state-changing actions:
// - GroupSelected: After a group is selected
// - GroupInserted: After a new group is inserted
// - ClassificationActionSameAsLast: After classifying file with same group as last
// - ClassificationActionSkip: After skipping a file
// - Screen transitions: When leaving classification screen
func (m Model) autoSaveState() Model {
	statePath := state.StateFilePath(m.directory)
	err := m.state.Save(statePath)
	if err != nil {
		m.err = fmt.Sprintf("Failed to save state: %v", err)
	}
	return m
}

// incrementActionAndMaybeSave increments the action counter and performs periodic auto-save
// This implements the periodic save feature that saves state every N actions
func (m Model) incrementActionAndMaybeSave() Model {
	m.actionCounter++
	// Check if we've reached the periodic save threshold
	if m.actionCounter >= m.actionsPerSave {
		m.actionCounter = 0 // Reset counter
		m = m.autoSaveState()
	}
	return m
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
		if m.classificationData != nil {
			return ClassificationView(m.classificationData)
		}
		return "Loading classification...\n\nPress Ctrl+C to quit"
	case ScreenGroupSelection:
		if m.groupSelectionData != nil {
			return GroupSelectionView(m.groupSelectionData)
		}
		return "Loading group selection...\n\nPress Ctrl+C to quit"
	case ScreenGroupInsertion:
		if m.groupInsertionData != nil {
			return GroupInsertionView(m.groupInsertionData)
		}
		return "Loading group insertion...\n\nPress Ctrl+C to quit"
	case ScreenReview:
		if m.reviewData != nil {
			return ReviewView(m.reviewData)
		}
		return "Loading review...\n\nPress Ctrl+C to quit"
	case ScreenComplete:
		if m.completionData != nil {
			return CompletionView(m.completionData)
		}
		return "Loading completion...\n\nPress Ctrl+C to quit"
	default:
		return "Unknown Screen\n\nPress Ctrl+C to quit"
	}
}
