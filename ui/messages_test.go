// ui/messages_test.go
package ui

import (
	"clip-tagger/state"
	"testing"
)

func TestTransitionToScreen(t *testing.T) {
	msg := TransitionToScreen{Screen: ScreenClassification}
	if msg.Screen != ScreenClassification {
		t.Errorf("expected screen to be ScreenClassification, got %v", msg.Screen)
	}
}

func TestStateUpdate(t *testing.T) {
	appState := state.NewState("/test/dir", state.SortByModifiedTime)
	msg := StateUpdate{State: appState}

	if msg.State != appState {
		t.Error("expected StateUpdate to contain correct state reference")
	}
}

func TestErrorMsg(t *testing.T) {
	errMsg := ErrorMsg{Err: "test error"}
	if errMsg.Err != "test error" {
		t.Errorf("expected error message 'test error', got '%s'", errMsg.Err)
	}
}
