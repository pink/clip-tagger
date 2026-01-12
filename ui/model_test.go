// ui/model_test.go
package ui

import (
	"clip-tagger/state"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	model := NewModel(appState, "/test/dir")

	if model.state != appState {
		t.Error("expected model to have correct state reference")
	}
	if model.currentScreen != ScreenStartup {
		t.Errorf("expected initial screen to be ScreenStartup, got %v", model.currentScreen)
	}
	if model.directory != "/test/dir" {
		t.Errorf("expected directory to be /test/dir, got %s", model.directory)
	}
}

func TestScreen_String(t *testing.T) {
	tests := []struct {
		screen   Screen
		expected string
	}{
		{ScreenStartup, "startup"},
		{ScreenClassification, "classification"},
		{ScreenGroupSelection, "group_selection"},
		{ScreenGroupInsertion, "group_insertion"},
		{ScreenReview, "review"},
		{ScreenComplete, "complete"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.screen.String()
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestModel_Init(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	model := NewModel(appState, "/test/dir")

	// Init should return a command for initialization
	cmd := model.Init()
	if cmd == nil {
		t.Error("expected Init to return a command")
	}
}

func TestModel_Update_ScreenTransitions(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	model := NewModel(appState, "/test/dir")

	t.Run("transition to classification screen", func(t *testing.T) {
		msg := TransitionToScreen{Screen: ScreenClassification}
		updated, _ := model.Update(msg)
		updatedModel := updated.(Model)

		if updatedModel.currentScreen != ScreenClassification {
			t.Errorf("expected screen to be ScreenClassification, got %v", updatedModel.currentScreen)
		}
	})

	t.Run("transition to group selection screen", func(t *testing.T) {
		msg := TransitionToScreen{Screen: ScreenGroupSelection}
		updated, _ := model.Update(msg)
		updatedModel := updated.(Model)

		if updatedModel.currentScreen != ScreenGroupSelection {
			t.Errorf("expected screen to be ScreenGroupSelection, got %v", updatedModel.currentScreen)
		}
	})

	t.Run("transition to complete screen", func(t *testing.T) {
		msg := TransitionToScreen{Screen: ScreenComplete}
		updated, _ := model.Update(msg)
		updatedModel := updated.(Model)

		if updatedModel.currentScreen != ScreenComplete {
			t.Errorf("expected screen to be ScreenComplete, got %v", updatedModel.currentScreen)
		}
	})
}

func TestModel_Update_QuitMessage(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	model := NewModel(appState, "/test/dir")

	t.Run("quit with ctrl+c", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyCtrlC}
		_, cmd := model.Update(msg)

		if cmd == nil {
			t.Fatal("expected quit command to be returned")
		}

		// Note: We can't easily test that it's exactly tea.Quit without exposing internals,
		// but we verify a command is returned
	})

	t.Run("no quit on regular key", func(t *testing.T) {
		msg := tea.KeyMsg{Type: tea.KeyEnter}
		_, cmd := model.Update(msg)

		// For now, enter key should not do anything in base model
		// (screen-specific handlers will be added later)
		// We can only check that cmd is nil for non-quit operations
		if cmd != nil {
			t.Error("expected regular key to return nil command")
		}
	})
}

func TestModel_View(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	model := NewModel(appState, "/test/dir")

	t.Run("view returns non-empty string", func(t *testing.T) {
		view := model.View()
		if view == "" {
			t.Error("expected View to return non-empty string")
		}
	})

	t.Run("view includes screen name", func(t *testing.T) {
		model.currentScreen = ScreenStartup
		view := model.View()
		// Should show something about the current screen
		if len(view) < 5 {
			t.Error("expected View to return meaningful content")
		}
	})

	t.Run("view changes with screen", func(t *testing.T) {
		model.currentScreen = ScreenStartup
		startupView := model.View()

		model.currentScreen = ScreenClassification
		classificationView := model.View()

		// Views should be different for different screens
		if startupView == classificationView {
			t.Error("expected different views for different screens")
		}
	})
}

func TestModel_StateUpdate(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	model := NewModel(appState, "/test/dir")

	t.Run("state update message updates state", func(t *testing.T) {
		newState := state.NewState("/new/dir", state.SortByName)
		msg := StateUpdate{State: newState}

		updated, _ := model.Update(msg)
		updatedModel := updated.(Model)

		if updatedModel.state.Directory != "/new/dir" {
			t.Errorf("expected directory to be updated to '/new/dir', got '%s'", updatedModel.state.Directory)
		}
	})
}

func TestAllScreensEnumerated(t *testing.T) {
	// Verify all 6 screens are defined
	screens := []Screen{
		ScreenStartup,
		ScreenClassification,
		ScreenGroupSelection,
		ScreenGroupInsertion,
		ScreenReview,
		ScreenComplete,
	}

	// Verify each screen has a unique value
	seen := make(map[Screen]bool)
	for _, screen := range screens {
		if seen[screen] {
			t.Errorf("duplicate screen value: %v", screen)
		}
		seen[screen] = true
	}

	if len(screens) != 6 {
		t.Errorf("expected 6 screens, found %d", len(screens))
	}
}
