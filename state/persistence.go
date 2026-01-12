// state/persistence.go
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const StateFileName = ".clip-tagger-state.json"
const LockFileName = ".clip-tagger.lock"

// Save writes state to JSON file
func (s *State) Save(path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write state file: %w", err)
	}

	return nil
}

// Load reads state from JSON file
func Load(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read state file: %w", err)
	}

	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("unmarshal state: %w", err)
	}

	return &state, nil
}

// StateFilePath returns the state file path for a directory
func StateFilePath(dir string) string {
	return filepath.Join(dir, StateFileName)
}

// StateExists checks if a state file exists
func StateExists(dir string) bool {
	_, err := os.Stat(StateFilePath(dir))
	return err == nil
}

// BackupState creates a backup of the state file
func BackupState(dir string) error {
	statePath := StateFilePath(dir)
	backupPath := statePath + ".bak"

	data, err := os.ReadFile(statePath)
	if err != nil {
		return fmt.Errorf("read state: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("write backup: %w", err)
	}

	return nil
}
