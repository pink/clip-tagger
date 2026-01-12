// state/merger.go
package state

// MergeResult contains the result of merging scanned files with state
type MergeResult struct {
	NewFiles      []string
	MissingFiles  []string
	ExistingCount int
}

// MergeFiles compares scanned files with state and identifies new/missing files
func MergeFiles(state *State, scannedFiles []string) *MergeResult {
	result := &MergeResult{
		NewFiles:     []string{},
		MissingFiles: []string{},
	}

	// Build map of classified files
	classified := make(map[string]bool)
	for _, c := range state.Classifications {
		classified[c.File] = true
	}

	// Build map of scanned files
	scanned := make(map[string]bool)
	for _, f := range scannedFiles {
		scanned[f] = true
	}

	// Find new files
	for _, f := range scannedFiles {
		if !classified[f] {
			result.NewFiles = append(result.NewFiles, f)
		} else {
			result.ExistingCount++
		}
	}

	// Find missing files
	for _, c := range state.Classifications {
		if !scanned[c.File] {
			result.MissingFiles = append(result.MissingFiles, c.File)
		}
	}

	return result
}
