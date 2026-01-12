// state/state_test.go
package state

import (
	"testing"
)

func TestState_NextTakeNumber(t *testing.T) {
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

func TestState_NextTakeNumber_EdgeCases(t *testing.T) {
	t.Run("empty classifications returns 1", func(t *testing.T) {
		state := &State{
			Classifications: []Classification{},
		}
		nextTake := state.NextTakeNumber("group1")
		if nextTake != 1 {
			t.Errorf("expected take 1 for empty classifications, got %d", nextTake)
		}
	})

	t.Run("group with no classifications returns 1", func(t *testing.T) {
		state := &State{
			Classifications: []Classification{
				{File: "vid1.mp4", GroupID: "group1", TakeNumber: 5},
				{File: "vid2.mp4", GroupID: "group1", TakeNumber: 3},
			},
		}
		nextTake := state.NextTakeNumber("group2")
		if nextTake != 1 {
			t.Errorf("expected take 1 for group with no classifications, got %d", nextTake)
		}
	})
}

func TestState_GetClassification(t *testing.T) {
	t.Run("finding existing file", func(t *testing.T) {
		state := &State{
			Classifications: []Classification{
				{File: "vid1.mp4", GroupID: "group1", TakeNumber: 1},
				{File: "vid2.mp4", GroupID: "group2", TakeNumber: 2},
			},
		}

		c, found := state.GetClassification("vid1.mp4")
		if !found {
			t.Fatal("expected to find classification")
		}
		if c.File != "vid1.mp4" {
			t.Errorf("expected file 'vid1.mp4', got '%s'", c.File)
		}
		if c.GroupID != "group1" {
			t.Errorf("expected groupID 'group1', got '%s'", c.GroupID)
		}
		if c.TakeNumber != 1 {
			t.Errorf("expected take 1, got %d", c.TakeNumber)
		}
	})

	t.Run("missing file returns false", func(t *testing.T) {
		state := &State{
			Classifications: []Classification{
				{File: "vid1.mp4", GroupID: "group1", TakeNumber: 1},
			},
		}

		_, found := state.GetClassification("nonexistent.mp4")
		if found {
			t.Error("expected not to find classification")
		}
	})

	t.Run("empty state returns false", func(t *testing.T) {
		state := &State{
			Classifications: []Classification{},
		}

		_, found := state.GetClassification("vid1.mp4")
		if found {
			t.Error("expected not to find classification in empty state")
		}
	})
}

func TestState_AddOrUpdateClassification(t *testing.T) {
	t.Run("adding new classification", func(t *testing.T) {
		state := &State{
			Groups: []Group{
				{ID: "group1", Name: "intro", Order: 1},
			},
			Classifications: []Classification{},
		}

		state.AddOrUpdateClassification("vid1.mp4", "group1")

		if len(state.Classifications) != 1 {
			t.Fatalf("expected 1 classification, got %d", len(state.Classifications))
		}

		c := state.Classifications[0]
		if c.File != "vid1.mp4" {
			t.Errorf("expected file 'vid1.mp4', got '%s'", c.File)
		}
		if c.GroupID != "group1" {
			t.Errorf("expected groupID 'group1', got '%s'", c.GroupID)
		}
		if c.TakeNumber != 1 {
			t.Errorf("expected take 1, got %d", c.TakeNumber)
		}
	})

	t.Run("updating existing classification removes old", func(t *testing.T) {
		state := &State{
			Groups: []Group{
				{ID: "group1", Name: "intro", Order: 1},
				{ID: "group2", Name: "outro", Order: 2},
			},
			Classifications: []Classification{
				{File: "vid1.mp4", GroupID: "group1", TakeNumber: 1},
				{File: "vid2.mp4", GroupID: "group1", TakeNumber: 2},
			},
		}

		state.AddOrUpdateClassification("vid1.mp4", "group2")

		if len(state.Classifications) != 2 {
			t.Fatalf("expected 2 classifications, got %d", len(state.Classifications))
		}

		// Verify old classification was removed
		for _, c := range state.Classifications {
			if c.File == "vid1.mp4" {
				if c.GroupID != "group2" {
					t.Errorf("expected updated groupID 'group2', got '%s'", c.GroupID)
				}
				if c.TakeNumber != 1 {
					t.Errorf("expected take 1 for group2, got %d", c.TakeNumber)
				}
			}
		}
	})

	t.Run("take numbers increment correctly", func(t *testing.T) {
		state := &State{
			Groups: []Group{
				{ID: "group1", Name: "intro", Order: 1},
			},
			Classifications: []Classification{
				{File: "vid1.mp4", GroupID: "group1", TakeNumber: 1},
				{File: "vid2.mp4", GroupID: "group1", TakeNumber: 2},
			},
		}

		state.AddOrUpdateClassification("vid3.mp4", "group1")

		c, found := state.GetClassification("vid3.mp4")
		if !found {
			t.Fatal("expected to find new classification")
		}
		if c.TakeNumber != 3 {
			t.Errorf("expected take 3, got %d", c.TakeNumber)
		}
	})
}

func TestNewState(t *testing.T) {
	directory := "/test/dir"
	sortBy := SortByModifiedTime

	state := NewState(directory, sortBy)

	if state.Directory != directory {
		t.Errorf("expected directory '%s', got '%s'", directory, state.Directory)
	}
	if state.SortBy != sortBy {
		t.Errorf("expected sortBy '%s', got '%s'", sortBy, state.SortBy)
	}
	if state.CurrentIndex != 0 {
		t.Errorf("expected CurrentIndex 0, got %d", state.CurrentIndex)
	}
	if state.Groups == nil {
		t.Error("expected Groups to be initialized, got nil")
	}
	if len(state.Groups) != 0 {
		t.Errorf("expected empty Groups, got length %d", len(state.Groups))
	}
	if state.Classifications == nil {
		t.Error("expected Classifications to be initialized, got nil")
	}
	if len(state.Classifications) != 0 {
		t.Errorf("expected empty Classifications, got length %d", len(state.Classifications))
	}
	if state.Skipped == nil {
		t.Error("expected Skipped to be initialized, got nil")
	}
	if len(state.Skipped) != 0 {
		t.Errorf("expected empty Skipped, got length %d", len(state.Skipped))
	}
}

func TestNewGroup(t *testing.T) {
	name := "intro"
	order := 1

	group := NewGroup(name, order)

	if group.Name != name {
		t.Errorf("expected name '%s', got '%s'", name, group.Name)
	}
	if group.Order != order {
		t.Errorf("expected order %d, got %d", order, group.Order)
	}
	if group.ID == "" {
		t.Error("expected non-empty ID")
	}

	// Verify UUID format (basic check - should have dashes)
	if len(group.ID) < 32 {
		t.Errorf("expected UUID format, got '%s'", group.ID)
	}

	// Verify UUIDs are unique
	group2 := NewGroup("outro", 2)
	if group.ID == group2.ID {
		t.Error("expected unique IDs for different groups")
	}
}
