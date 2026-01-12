// ui/classification_logic.go
package ui

// handleClassificationSameAsLast handles the "Same as Last" classification action
// It finds the most recent classified file and assigns the current file to the same group
func (m Model) handleClassificationSameAsLast() Model {
	if m.currentFileIndex >= len(m.files) {
		// No current file to classify
		return m
	}

	currentFile := m.files[m.currentFileIndex]

	// Find the most recent classified file before current index
	var lastGroupID string
	for i := m.currentFileIndex - 1; i >= 0; i-- {
		prevFile := m.files[i]
		if classification, ok := m.state.GetClassification(prevFile); ok {
			lastGroupID = classification.GroupID
			break
		}
	}

	// If we found a previous classification, apply it to current file
	if lastGroupID != "" {
		m.state.AddOrUpdateClassification(currentFile, lastGroupID)
	}

	// Advance to next file
	m.currentFileIndex++

	// Check if we're done with all files
	if m.currentFileIndex >= len(m.files) {
		// Transition to review screen
		m.currentScreen = ScreenReview
		m.reviewData = NewReviewData(m.state, m.files)
	} else {
		// Update classification data for next file
		m.classificationData = NewClassificationData(m.state, m.files, m.currentFileIndex)
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

	// Advance to next file
	m.currentFileIndex++

	// Check if we're done with all files
	if m.currentFileIndex >= len(m.files) {
		// Transition to review screen
		m.currentScreen = ScreenReview
		m.reviewData = NewReviewData(m.state, m.files)
	} else {
		// Update classification data for next file
		m.classificationData = NewClassificationData(m.state, m.files, m.currentFileIndex)
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

	// Advance to next file
	m.currentFileIndex++

	// Check if we're done with all files
	if m.currentFileIndex >= len(m.files) {
		// Transition to review screen
		m.currentScreen = ScreenReview
		m.reviewData = NewReviewData(m.state, m.files)
	} else {
		// Update classification data for next file
		m.classificationData = NewClassificationData(m.state, m.files, m.currentFileIndex)
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

	// Advance to next file
	m.currentFileIndex++

	// Check if we're done with all files
	if m.currentFileIndex >= len(m.files) {
		// Transition to review screen
		m.currentScreen = ScreenReview
		m.reviewData = NewReviewData(m.state, m.files)
	} else {
		// Update classification data for next file
		m.classificationData = NewClassificationData(m.state, m.files, m.currentFileIndex)
	}

	return m
}
