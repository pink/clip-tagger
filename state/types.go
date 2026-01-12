// state/types.go
package state

import "github.com/google/uuid"

// SortBy defines how files should be sorted
type SortBy string

const (
	SortByModifiedTime SortBy = "modified_time"
	SortByCreatedTime  SortBy = "created_time"
	SortByName         SortBy = "name"
)

// State represents the complete session state
type State struct {
	Directory       string           `json:"directory"`
	SortBy          SortBy           `json:"sort_by"`
	CurrentIndex    int              `json:"current_index"`
	Groups          []Group          `json:"groups"`
	Classifications []Classification `json:"classifications"`
	Skipped         []string         `json:"skipped"`
}

// Group represents a semantic group of clips
type Group struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Order int    `json:"order"`
}

// Classification links a file to a group with take number
type Classification struct {
	File       string `json:"file"`
	GroupID    string `json:"group_id"`
	TakeNumber int    `json:"take_number"`
}

// NewState creates a new empty state
func NewState(directory string, sortBy SortBy) *State {
	return &State{
		Directory:       directory,
		SortBy:          sortBy,
		CurrentIndex:    0,
		Groups:          []Group{},
		Classifications: []Classification{},
		Skipped:         []string{},
	}
}

// NewGroup creates a new group with generated UUID
func NewGroup(name string, order int) Group {
	return Group{
		ID:    uuid.New().String(),
		Name:  name,
		Order: order,
	}
}

// FindGroupByID finds a group by its ID
func (s *State) FindGroupByID(id string) *Group {
	for i := range s.Groups {
		if s.Groups[i].ID == id {
			return &s.Groups[i]
		}
	}
	return nil
}

// NextTakeNumber calculates the next take number for a group
func (s *State) NextTakeNumber(groupID string) int {
	maxTake := 0
	for _, c := range s.Classifications {
		if c.GroupID == groupID && c.TakeNumber > maxTake {
			maxTake = c.TakeNumber
		}
	}
	return maxTake + 1
}

// GetClassification finds classification for a file
func (s *State) GetClassification(filename string) *Classification {
	for i := range s.Classifications {
		if s.Classifications[i].File == filename {
			return &s.Classifications[i]
		}
	}
	return nil
}

// AddOrUpdateClassification adds or updates a classification
func (s *State) AddOrUpdateClassification(filename, groupID string) {
	// Remove existing classification if present
	for i := range s.Classifications {
		if s.Classifications[i].File == filename {
			s.Classifications = append(s.Classifications[:i], s.Classifications[i+1:]...)
			break
		}
	}

	// Add new classification
	takeNum := s.NextTakeNumber(groupID)
	s.Classifications = append(s.Classifications, Classification{
		File:       filename,
		GroupID:    groupID,
		TakeNumber: takeNum,
	})
}
