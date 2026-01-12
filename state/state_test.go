// state/state_test.go
package state

import (
	"testing"
)

func TestGroup_NextTakeNumber(t *testing.T) {
	state := &State{
		Groups: []Group{
			{ID: "group1", Name: "intro", Order: 1},
		},
		Classifications: []Classification{
			{File: "vid1.mp4", GroupID: "group1", TakeNumber: 1},
			{File: "vid2.mp4", GroupID: "group1", TakeNumber: 2},
		},
	}

	nextTake := state.NextTakeNumber("group1")
	if nextTake != 3 {
		t.Errorf("expected take 3, got %d", nextTake)
	}
}

func TestState_FindGroupByID(t *testing.T) {
	group1 := Group{ID: "id1", Name: "intro", Order: 1}
	state := &State{
		Groups: []Group{group1},
	}

	found := state.FindGroupByID("id1")
	if found == nil {
		t.Fatal("expected to find group")
	}
	if found.Name != "intro" {
		t.Errorf("expected 'intro', got '%s'", found.Name)
	}

	notFound := state.FindGroupByID("nonexistent")
	if notFound != nil {
		t.Error("expected nil for nonexistent group")
	}
}
