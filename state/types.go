// state/types.go
package state

import (
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
)

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
// Returns the classification and true if found, or an empty classification and false if not found
func (s *State) GetClassification(filename string) (Classification, bool) {
	for _, c := range s.Classifications {
		if c.File == filename {
			return c, true
		}
	}
	return Classification{}, false
}

// AddOrUpdateClassification adds or updates a classification
func (s *State) AddOrUpdateClassification(filename, groupID string) {
	// Remove existing classification if present by building new slice
	newClassifications := make([]Classification, 0, len(s.Classifications))
	for _, c := range s.Classifications {
		if c.File != filename {
			newClassifications = append(newClassifications, c)
		}
	}
	s.Classifications = newClassifications

	// Add new classification
	takeNum := s.NextTakeNumber(groupID)
	s.Classifications = append(s.Classifications, Classification{
		File:       filename,
		GroupID:    groupID,
		TakeNumber: takeNum,
	})
}

// RepairRenamedFiles attempts to fix Classifications that reference old filenames
// by matching them to renamed files that follow the [XX_YY] Name pattern
func (s *State) RepairRenamedFiles(scannedFiles []string) int {
	// Build map of scanned files for quick lookup
	scannedMap := make(map[string]bool)
	for _, f := range scannedFiles {
		scannedMap[f] = true
	}

	repairedCount := 0

	// Check each Classification
	for i := range s.Classifications {
		classification := &s.Classifications[i]

		// Skip if file still exists (not missing)
		if scannedMap[classification.File] {
			continue
		}

		// File is missing - try to find renamed version
		group := s.FindGroupByID(classification.GroupID)
		if group == nil {
			continue
		}

		// Generate expected renamed filename
		ext := filepath.Ext(classification.File)
		groupNum := formatNumber(group.Order)
		takeNum := formatNumber(classification.TakeNumber)
		expectedName := fmt.Sprintf("[%s_%s] %s%s", groupNum, takeNum, group.Name, ext)

		// Check if expected renamed file exists
		if scannedMap[expectedName] {
			// Update classification to use new filename
			classification.File = expectedName
			repairedCount++
		}
	}

	return repairedCount
}

// formatNumber formats a number with leading zero (01, 02, ..., 10, 11, ...)
func formatNumber(n int) string {
	if n < 10 {
		return fmt.Sprintf("0%d", n)
	}
	return fmt.Sprintf("%d", n)
}
