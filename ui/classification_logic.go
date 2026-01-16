// ui/classification_logic.go
package ui

// findNextUnclassifiedFile advances currentFileIndex to the next unclassified file
// Returns true if found, false if all remaining files are classified
func (m *Model) findNextUnclassifiedFile() bool {
	for m.currentFileIndex < len(m.files) {
		currentFile := m.files[m.currentFileIndex]
		_, classified := m.state.GetClassification(currentFile)
		if !classified {
			return true
		}
		// This file is already classified, skip to next
		m.currentFileIndex++
	}
	return false
}

// handleClassificationSameAsLast handles the "Same as Last" classification action
// It uses the most recently classified group from the current session
func (m Model) handleClassificationSameAsLast() Model {
	if m.currentFileIndex >= len(m.files) {
		// No current file to classify
		return m
	}

	currentFile := m.files[m.currentFileIndex]

	// Use the last classified group ID from the current session
	// If no group has been classified yet, fall back to searching backwards
	lastGroupID := m.lastClassifiedGroupID
	if lastGroupID == "" {
		// Fallback: Find the most recent classified file before current index
		for i := m.currentFileIndex - 1; i >= 0; i-- {
			prevFile := m.files[i]
			if classification, ok := m.state.GetClassification(prevFile); ok {
				lastGroupID = classification.GroupID
				break
			}
		}
	}

	// If we found a previous classification, apply it to current file
	if lastGroupID != "" {
		m.state.AddOrUpdateClassification(currentFile, lastGroupID)
		// Update lastClassifiedGroupID for next "Same as Last"
		m.lastClassifiedGroupID = lastGroupID
	}

	// Advance to next unclassified file
	m.currentFileIndex++
	hasNext := m.findNextUnclassifiedFile()
	m.state.CurrentIndex = m.currentFileIndex

	// Check if we're done with all files
	if !hasNext {
		// Transition to review screen
		m.currentScreen = ScreenReview
		m.reviewData = NewReviewData(m.state, m.files)
	} else {
		// Update classification data for next file
		m.classificationData = NewClassificationData(m.state, m.files, m.currentFileIndex, m.lastClassifiedGroupID)
	}

	return m
}

// handleGroupSelected handles when a user selects an existing group
func (m Model) handleGroupSelected(groupID string) Model {
	if m.currentFileIndex >= len(m.files) {
		// No current file to classify
		return m
	}

	currentFile := m.files[m.currentFileIndex]

	// Classify the current file with the selected group
	m.state.AddOrUpdateClassification(currentFile, groupID)
	// Track for "Same as Last"
	m.lastClassifiedGroupID = groupID

	// Advance to next unclassified file
	m.currentFileIndex++
	hasNext := m.findNextUnclassifiedFile()
	m.state.CurrentIndex = m.currentFileIndex

	// Check if we're done with all files
	if !hasNext {
		// Transition to review screen
		m.currentScreen = ScreenReview
		m.reviewData = NewReviewData(m.state, m.files)
	} else {
		// Update classification data for next file
		m.classificationData = NewClassificationData(m.state, m.files, m.currentFileIndex, m.lastClassifiedGroupID)
		// Transition back to classification screen
		m.currentScreen = ScreenClassification
	}

	return m
}

// handleGroupInserted handles when a user creates a new group
func (m Model) handleGroupInserted(groupID, groupName string, order int) Model {
	if m.currentFileIndex >= len(m.files) {
		// No current file to classify
		return m
	}

	currentFile := m.files[m.currentFileIndex]

	// The group has already been added to state by the GroupInserted message handler
	// Now classify the current file with the new group
	m.state.AddOrUpdateClassification(currentFile, groupID)
	// Track for "Same as Last"
	m.lastClassifiedGroupID = groupID

	// Advance to next unclassified file
	m.currentFileIndex++
	hasNext := m.findNextUnclassifiedFile()
	m.state.CurrentIndex = m.currentFileIndex

	// Check if we're done with all files
	if !hasNext {
		// Transition to review screen
		m.currentScreen = ScreenReview
		m.reviewData = NewReviewData(m.state, m.files)
	} else {
		// Update classification data for next file
		m.classificationData = NewClassificationData(m.state, m.files, m.currentFileIndex, m.lastClassifiedGroupID)
		// Transition back to classification screen
		m.currentScreen = ScreenClassification
	}

	return m
}

// handleClassificationSkip handles when a user skips a file
func (m Model) handleClassificationSkip() Model {
	if m.currentFileIndex >= len(m.files) {
		// No current file to skip
		return m
	}

	currentFile := m.files[m.currentFileIndex]

	// Add file to skipped list
	m.state.Skipped = append(m.state.Skipped, currentFile)

	// Advance to next unclassified file
	m.currentFileIndex++
	hasNext := m.findNextUnclassifiedFile()
	m.state.CurrentIndex = m.currentFileIndex

	// Check if we're done with all files
	if !hasNext {
		// Transition to review screen
		m.currentScreen = ScreenReview
		m.reviewData = NewReviewData(m.state, m.files)
	} else {
		// Update classification data for next file
		m.classificationData = NewClassificationData(m.state, m.files, m.currentFileIndex, m.lastClassifiedGroupID)
	}

	return m
}
