# clip-tagger Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build an interactive CLI tool for reviewing and batch-renaming video files into a structured naming convention with semantic grouping.

**Architecture:** Bubbletea TUI with stateful file classification, persistent JSON state management, and safe batch rename operations. Core components: file scanner, state manager, UI screens (classification, group selection, review), and renamer with conflict detection.

**Tech Stack:** Go 1.21+, bubbletea (TUI framework), bubbles (UI components), google/uuid

---

## Task 1: Project Initialization

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `.gitignore`

**Step 1: Initialize Go module**

Run:
```bash
go mod init github.com/yourusername/clip-tagger
```

Expected: Creates `go.mod` with module declaration

**Step 2: Add dependencies**

Run:
```bash
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/bubbles/list@latest
go get github.com/charmbracelet/bubbles/progress@latest
go get github.com/google/uuid@latest
```

Expected: Dependencies added to `go.mod`

**Step 3: Create main.go skeleton**

```go
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("clip-tagger v0.1.0")
	if len(os.Args) < 2 {
		fmt.Println("Usage: clip-tagger <directory>")
		os.Exit(1)
	}
}
```

**Step 4: Create .gitignore**

```
# Binaries
clip-tagger
*.exe
*.dll
*.so
*.dylib

# Test binary
*.test

# Output
*.out

# Go workspace
go.work

# State files (for testing)
.clip-tagger-state.json
.clip-tagger.lock
```

**Step 5: Test compilation**

Run:
```bash
go build -o clip-tagger .
./clip-tagger
```

Expected: Prints usage message

**Step 6: Commit**

```bash
git add go.mod go.sum main.go .gitignore
git commit -m "chore: initialize Go project with dependencies"
```

---

## Task 2: State Data Structures

**Files:**
- Create: `state/types.go`
- Create: `state/state_test.go`

**Step 1: Write test for State types**

```go
// state/state_test.go
package state

import (
	"testing"

	"github.com/google/uuid"
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
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./state/...
```

Expected: FAIL with "no such file or directory"

**Step 3: Write State types**

```go
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
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./state/... -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add state/
git commit -m "feat: add state data structures and basic operations"
```

---

## Task 3: State Persistence

**Files:**
- Modify: `state/types.go` (add persistence methods)
- Create: `state/persistence.go`
- Modify: `state/state_test.go` (add persistence tests)

**Step 1: Write test for Save and Load**

Add to `state/state_test.go`:

```go
func TestState_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := tmpDir + "/.clip-tagger-state.json"

	// Create and save state
	original := NewState(tmpDir, SortByModifiedTime)
	original.Groups = append(original.Groups, NewGroup("intro", 1))
	original.CurrentIndex = 5

	err := original.Save(statePath)
	if err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Load state
	loaded, err := Load(statePath)
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	if loaded.Directory != original.Directory {
		t.Errorf("directory mismatch")
	}
	if loaded.CurrentIndex != 5 {
		t.Errorf("expected index 5, got %d", loaded.CurrentIndex)
	}
	if len(loaded.Groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(loaded.Groups))
	}
}

func TestLoad_NonExistent(t *testing.T) {
	_, err := Load("/nonexistent/path/.clip-tagger-state.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./state/... -v -run TestState_SaveAndLoad
```

Expected: FAIL with "undefined: State.Save"

**Step 3: Implement persistence**

```go
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
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./state/... -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add state/
git commit -m "feat: add state persistence with JSON serialization"
```

---

## Task 4: File Scanner

**Files:**
- Create: `scanner/scanner.go`
- Create: `scanner/scanner_test.go`

**Step 1: Write scanner tests**

```go
// scanner/scanner_test.go
package scanner

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScanner_ScanDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test video files
	files := []string{"vid1.mp4", "vid2.mov", "vid3.avi", "readme.txt"}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	scanner := NewScanner(tmpDir)
	result, err := scanner.Scan(SortByName)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	// Should find 3 video files, not the txt
	if len(result.Files) != 3 {
		t.Errorf("expected 3 files, got %d", len(result.Files))
	}

	// Check sorted by name
	expected := []string{"vid1.mp4", "vid2.mov", "vid3.avi"}
	for i, f := range result.Files {
		if filepath.Base(f.Path) != expected[i] {
			t.Errorf("index %d: expected %s, got %s", i, expected[i], filepath.Base(f.Path))
		}
	}
}

func TestIsVideoFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"video.mp4", true},
		{"video.MP4", true},
		{"video.mov", true},
		{"video.avi", true},
		{"video.mkv", true},
		{"video.webm", true},
		{"document.pdf", false},
		{"image.jpg", false},
		{"noext", false},
	}

	for _, tt := range tests {
		result := isVideoFile(tt.path)
		if result != tt.expected {
			t.Errorf("isVideoFile(%s) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./scanner/... -v
```

Expected: FAIL with "no such file or directory"

**Step 3: Implement scanner**

```go
// scanner/scanner.go
package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SortBy defines file sorting order
type SortBy string

const (
	SortByModifiedTime SortBy = "modified_time"
	SortByCreatedTime  SortBy = "created_time"
	SortByName         SortBy = "name"
)

var videoExtensions = map[string]bool{
	".mp4":  true,
	".mov":  true,
	".avi":  true,
	".mkv":  true,
	".webm": true,
}

// FileInfo contains metadata about a scanned file
type FileInfo struct {
	Path         string
	Name         string
	ModifiedTime time.Time
	CreatedTime  time.Time
}

// ScanResult contains the results of a directory scan
type ScanResult struct {
	Files []FileInfo
	Total int
}

// Scanner scans directories for video files
type Scanner struct {
	directory string
}

// NewScanner creates a new scanner for a directory
func NewScanner(directory string) *Scanner {
	return &Scanner{directory: directory}
}

// Scan scans the directory for video files and sorts them
func (s *Scanner) Scan(sortBy SortBy) (*ScanResult, error) {
	var files []FileInfo

	entries, err := os.ReadDir(s.directory)
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(s.directory, entry.Name())
		if !isVideoFile(path) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue // Skip files we can't stat
		}

		files = append(files, FileInfo{
			Path:         path,
			Name:         entry.Name(),
			ModifiedTime: info.ModTime(),
			CreatedTime:  info.ModTime(), // Use ModTime as fallback
		})
	}

	sortFiles(files, sortBy)

	return &ScanResult{
		Files: files,
		Total: len(files),
	}, nil
}

// isVideoFile checks if a file has a video extension
func isVideoFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return videoExtensions[ext]
}

// sortFiles sorts files by the specified order
func sortFiles(files []FileInfo, sortBy SortBy) {
	switch sortBy {
	case SortByModifiedTime:
		sort.Slice(files, func(i, j int) bool {
			return files[i].ModifiedTime.Before(files[j].ModifiedTime)
		})
	case SortByCreatedTime:
		sort.Slice(files, func(i, j int) bool {
			return files[i].CreatedTime.Before(files[j].CreatedTime)
		})
	case SortByName:
		sort.Slice(files, func(i, j int) bool {
			return files[i].Name < files[j].Name
		})
	}
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./scanner/... -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add scanner/
git commit -m "feat: add file scanner with video detection and sorting"
```

---

## Task 5: File Merger (New vs Existing State)

**Files:**
- Create: `state/merger.go`
- Create: `state/merger_test.go`

**Step 1: Write merger tests**

```go
// state/merger_test.go
package state

import (
	"testing"
)

func TestMergeFiles(t *testing.T) {
	state := &State{
		Directory:       "/test",
		Classifications: []Classification{
			{File: "existing1.mp4", GroupID: "g1", TakeNumber: 1},
			{File: "existing2.mp4", GroupID: "g2", TakeNumber: 1},
		},
	}

	scannedFiles := []string{
		"existing1.mp4",
		"existing2.mp4",
		"new1.mp4",
		"new2.mp4",
	}

	result := MergeFiles(state, scannedFiles)

	if len(result.NewFiles) != 2 {
		t.Errorf("expected 2 new files, got %d", len(result.NewFiles))
	}

	if len(result.MissingFiles) != 0 {
		t.Errorf("expected 0 missing files, got %d", len(result.MissingFiles))
	}

	if result.ExistingCount != 2 {
		t.Errorf("expected 2 existing, got %d", result.ExistingCount)
	}
}

func TestMergeFiles_WithMissing(t *testing.T) {
	state := &State{
		Directory: "/test",
		Classifications: []Classification{
			{File: "old1.mp4", GroupID: "g1", TakeNumber: 1},
			{File: "old2.mp4", GroupID: "g2", TakeNumber: 1},
		},
	}

	scannedFiles := []string{"new1.mp4"}

	result := MergeFiles(state, scannedFiles)

	if len(result.MissingFiles) != 2 {
		t.Errorf("expected 2 missing files, got %d", len(result.MissingFiles))
	}

	expectedMissing := map[string]bool{"old1.mp4": true, "old2.mp4": true}
	for _, f := range result.MissingFiles {
		if !expectedMissing[f] {
			t.Errorf("unexpected missing file: %s", f)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./state/... -v -run TestMergeFiles
```

Expected: FAIL with "undefined: MergeFiles"

**Step 3: Implement merger**

```go
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
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./state/... -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add state/
git commit -m "feat: add file merger for detecting new/missing files"
```

---

## Task 6: Renamer - Name Generation

**Files:**
- Create: `renamer/generator.go`
- Create: `renamer/generator_test.go`

**Step 1: Write generator tests**

```go
// renamer/generator_test.go
package renamer

import (
	"testing"
)

func TestGenerateFilename(t *testing.T) {
	tests := []struct {
		groupOrder int
		takeNum    int
		groupName  string
		origExt    string
		expected   string
	}{
		{1, 1, "intro", ".mp4", "[01_01] intro.mp4"},
		{2, 3, "magic trick", ".mov", "[02_03] magic trick.mov"},
		{15, 2, "outro", ".avi", "[15_02] outro.avi"},
	}

	for _, tt := range tests {
		result := GenerateFilename(tt.groupOrder, tt.takeNum, tt.groupName, tt.origExt)
		if result != tt.expected {
			t.Errorf("GenerateFilename(%d, %d, %s, %s) = %s, want %s",
				tt.groupOrder, tt.takeNum, tt.groupName, tt.origExt,
				result, tt.expected)
		}
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		num      int
		expected string
	}{
		{1, "01"},
		{9, "09"},
		{10, "10"},
		{99, "99"},
		{100, "100"},
	}

	for _, tt := range tests {
		result := formatNumber(tt.num)
		if result != tt.expected {
			t.Errorf("formatNumber(%d) = %s, want %s", tt.num, result, tt.expected)
		}
	}
}
```

**Step 2: Run test to verify it fails**

Run:
```bash
go test ./renamer/... -v
```

Expected: FAIL with "no such file or directory"

**Step 3: Implement generator**

```go
// renamer/generator.go
package renamer

import (
	"fmt"
	"path/filepath"
)

// GenerateFilename creates a filename in format [XX_YY] name.ext
func GenerateFilename(groupOrder, takeNumber int, groupName, extension string) string {
	groupNum := formatNumber(groupOrder)
	takeNum := formatNumber(takeNumber)
	return fmt.Sprintf("[%s_%s] %s%s", groupNum, takeNum, groupName, extension)
}

// formatNumber formats a number with leading zero (01, 02, ..., 10, 11, ...)
func formatNumber(n int) string {
	if n < 10 {
		return fmt.Sprintf("0%d", n)
	}
	return fmt.Sprintf("%d", n)
}

// GenerateTargetPath generates the full target path for a file
func GenerateTargetPath(directory, originalPath string, groupOrder, takeNumber int, groupName string) string {
	ext := filepath.Ext(originalPath)
	newName := GenerateFilename(groupOrder, takeNumber, groupName, ext)
	return filepath.Join(directory, newName)
}
```

**Step 4: Run test to verify it passes**

Run:
```bash
go test ./renamer/... -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add renamer/
git commit -m "feat: add filename generator for structured naming"
```

---

## Task 7: Renamer - Conflict Detection and Operations

**Files:**
- Modify: `renamer/generator.go` (add conflict detection)
- Create: `renamer/operations.go`
- Modify: `renamer/generator_test.go` (add tests)
- Create: `renamer/operations_test.go`

**Step 1: Write conflict detection test**

Add to `renamer/generator_test.go`:

```go
func TestDetectConflicts(t *testing.T) {
	tmpDir := t.TempDir()

	// Create existing file
	existing := filepath.Join(tmpDir, "[01_01] intro.mp4")
	if err := os.WriteFile(existing, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	renames := []Rename{
		{
			OriginalPath: filepath.Join(tmpDir, "vid1.mp4"),
			TargetPath:   filepath.Join(tmpDir, "[01_01] intro.mp4"), // Conflicts
		},
		{
			OriginalPath: filepath.Join(tmpDir, "vid2.mp4"),
			TargetPath:   filepath.Join(tmpDir, "[02_01] outro.mp4"), // No conflict
		},
	}

	conflicts := DetectConflicts(renames)
	if len(conflicts) != 1 {
		t.Errorf("expected 1 conflict, got %d", len(conflicts))
	}

	if conflicts[0].TargetPath != existing {
		t.Errorf("wrong conflict path: %s", conflicts[0].TargetPath)
	}
}
```

**Step 2: Write operations test**

```go
// renamer/operations_test.go
package renamer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRenameInPlace(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source file
	src := filepath.Join(tmpDir, "source.mp4")
	if err := os.WriteFile(src, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(tmpDir, "[01_01] intro.mp4")
	rename := Rename{OriginalPath: src, TargetPath: dst}

	err := RenameInPlace([]Rename{rename})
	if err != nil {
		t.Fatalf("rename failed: %v", err)
	}

	// Check source gone, target exists
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("source still exists")
	}

	if _, err := os.Stat(dst); err != nil {
		t.Error("target doesn't exist")
	}

	// Verify content
	content, _ := os.ReadFile(dst)
	if string(content) != "content" {
		t.Error("content mismatch")
	}
}

func TestCopyToDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	outputDir := filepath.Join(tmpDir, "output")

	// Create source file
	src := filepath.Join(tmpDir, "source.mp4")
	if err := os.WriteFile(src, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(outputDir, "[01_01] intro.mp4")
	rename := Rename{OriginalPath: src, TargetPath: dst}

	err := CopyToDirectory([]Rename{rename}, outputDir)
	if err != nil {
		t.Fatalf("copy failed: %v", err)
	}

	// Check source still exists
	if _, err := os.Stat(src); err != nil {
		t.Error("source was removed")
	}

	// Check target exists
	if _, err := os.Stat(dst); err != nil {
		t.Error("target doesn't exist")
	}

	// Verify content
	content, _ := os.ReadFile(dst)
	if string(content) != "content" {
		t.Error("content mismatch")
	}
}
```

**Step 3: Run tests to verify they fail**

Run:
```bash
go test ./renamer/... -v
```

Expected: FAIL with undefined types/functions

**Step 4: Implement conflict detection and operations**

Add to `renamer/generator.go`:

```go
import "os"

// Rename represents a file rename operation
type Rename struct {
	OriginalPath string
	TargetPath   string
	ChangeType   string // "new", "updated", "moved", or ""
}

// DetectConflicts checks if any target paths already exist
func DetectConflicts(renames []Rename) []Rename {
	var conflicts []Rename
	for _, r := range renames {
		// Skip if target is same as source (no actual rename)
		if r.OriginalPath == r.TargetPath {
			continue
		}

		if _, err := os.Stat(r.TargetPath); err == nil {
			conflicts = append(conflicts, r)
		}
	}
	return conflicts
}
```

Create `renamer/operations.go`:

```go
// renamer/operations.go
package renamer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// RenameInPlace renames files in their current directory
func RenameInPlace(renames []Rename) error {
	for _, r := range renames {
		// Skip if no actual change
		if r.OriginalPath == r.TargetPath {
			continue
		}

		if err := os.Rename(r.OriginalPath, r.TargetPath); err != nil {
			return fmt.Errorf("rename %s -> %s: %w",
				filepath.Base(r.OriginalPath),
				filepath.Base(r.TargetPath),
				err)
		}
	}
	return nil
}

// CopyToDirectory copies files to a new directory
func CopyToDirectory(renames []Rename, outputDir string) error {
	// Create output directory if needed
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	for _, r := range renames {
		targetPath := filepath.Join(outputDir, filepath.Base(r.TargetPath))

		if err := copyFile(r.OriginalPath, targetPath); err != nil {
			return fmt.Errorf("copy %s -> %s: %w",
				filepath.Base(r.OriginalPath),
				filepath.Base(targetPath),
				err)
		}
	}
	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return dstFile.Sync()
}
```

**Step 5: Run tests to verify they pass**

Run:
```bash
go test ./renamer/... -v
```

Expected: PASS

**Step 6: Commit**

```bash
git add renamer/
git commit -m "feat: add conflict detection and rename operations"
```

---

## Task 8: Basic Bubbletea UI Structure

**Files:**
- Create: `ui/model.go`
- Create: `ui/messages.go`
- Modify: `main.go` (integrate bubbletea)

**Step 1: Create UI model skeleton**

```go
// ui/model.go
package ui

import (
	"github.com/charmbracelet/bubbletea"
	tea "github.com/charmbracelet/bubbletea"
)

// Screen represents different UI screens
type Screen int

const (
	ScreenStartup Screen = iota
	ScreenClassification
	ScreenGroupSelect
	ScreenGroupInsert
	ScreenReview
)

// Model is the main UI model
type Model struct {
	screen       Screen
	directory    string
	files        []string
	currentIndex int
	width        int
	height       int
	err          error
}

// NewModel creates a new UI model
func NewModel(directory string, files []string) Model {
	return Model{
		screen:       ScreenStartup,
		directory:    directory,
		files:        files,
		currentIndex: 0,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error() + "\n"
	}

	switch m.screen {
	case ScreenStartup:
		return "Starting up...\n"
	default:
		return "clip-tagger\n"
	}
}
```

**Step 2: Create message types**

```go
// ui/messages.go
package ui

// StartClassificationMsg signals to start classification
type StartClassificationMsg struct{}

// FileClassifiedMsg signals a file was classified
type FileClassifiedMsg struct {
	Filename string
	GroupID  string
}

// ErrorMsg signals an error occurred
type ErrorMsg struct {
	Err error
}
```

**Step 3: Update main.go to use bubbletea**

```go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourusername/clip-tagger/ui"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: clip-tagger <directory>")
		os.Exit(1)
	}

	directory := os.Args[1]

	// Placeholder: will integrate scanner later
	files := []string{}

	model := ui.NewModel(directory, files)
	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 4: Test compilation and run**

Run:
```bash
go build -o clip-tagger .
./clip-tagger .
```

Expected: Shows "Starting up..." and responds to 'q' to quit

**Step 5: Commit**

```bash
git add ui/ main.go
git commit -m "feat: add basic bubbletea UI structure"
```

---

## Task 9: Startup Screen with State Detection

**Files:**
- Create: `ui/startup.go`
- Modify: `ui/model.go` (integrate startup)
- Modify: `main.go` (integrate scanner and state)

**Step 1: Create startup screen**

```go
// ui/startup.go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// StartupModel handles the startup screen
type StartupModel struct {
	hasExistingState bool
	newFileCount     int
	existingCount    int
	missingCount     int
	resumeMode       bool // true = resume, false = fresh
	width            int
	height           int
}

// StartupKeyMap defines keyboard shortcuts
type StartupKeyMap struct {
	Resume key.Binding
	Fresh  key.Binding
	Quit   key.Binding
}

var startupKeys = StartupKeyMap{
	Resume: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "resume"),
	),
	Fresh: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "fresh start"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

func NewStartupModel(hasState bool, newFiles, existing, missing int) StartupModel {
	return StartupModel{
		hasExistingState: hasState,
		newFileCount:     newFiles,
		existingCount:    existing,
		missingCount:     missing,
		resumeMode:       true, // Default to resume
	}
}

func (m StartupModel) Update(msg tea.Msg) (StartupModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			m.resumeMode = true
			return m, func() tea.Msg { return StartClassificationMsg{} }
		case "f":
			m.resumeMode = false
			return m, func() tea.Msg { return StartClassificationMsg{} }
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m StartupModel) View() string {
	var b strings.Builder

	b.WriteString("\n  clip-tagger\n\n")

	if m.hasExistingState {
		b.WriteString(fmt.Sprintf("  Found existing session:\n"))
		b.WriteString(fmt.Sprintf("    • %d already classified\n", m.existingCount))
		if m.newFileCount > 0 {
			b.WriteString(fmt.Sprintf("    • %d new files detected\n", m.newFileCount))
		}
		if m.missingCount > 0 {
			b.WriteString(fmt.Sprintf("    • %d files missing\n", m.missingCount))
		}
		b.WriteString("\n")
		b.WriteString("  [r] Resume    [f] Start fresh    [q] Quit\n")
	} else {
		b.WriteString("  No existing session found.\n\n")
		b.WriteString("  Press any key to start...\n")
	}

	return b.String()
}
```

**Step 2: Integrate startup into main model**

Update `ui/model.go`:

```go
// Add to Model struct
type Model struct {
	screen          Screen
	directory       string
	files           []string
	currentIndex    int
	width           int
	height          int
	err             error
	startupModel    StartupModel    // Add this
}

// Update NewModel
func NewModel(directory string, files []string, hasState bool, newFiles, existing, missing int) Model {
	return Model{
		screen:       ScreenStartup,
		directory:    directory,
		files:        files,
		currentIndex: 0,
		startupModel: NewStartupModel(hasState, newFiles, existing, missing),
	}
}

// Update Update method to delegate to startup
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case StartClassificationMsg:
		m.screen = ScreenClassification
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Delegate to screen-specific models
	if m.screen == ScreenStartup {
		var cmd tea.Cmd
		m.startupModel, cmd = m.startupModel.Update(msg)
		return m, cmd
	}

	return m, nil
}

// Update View method
func (m Model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error() + "\n"
	}

	switch m.screen {
	case ScreenStartup:
		return m.startupModel.View()
	case ScreenClassification:
		return "Classification screen (TODO)\n"
	default:
		return "Unknown screen\n"
	}
}
```

**Step 3: Update main.go to pass state info**

```go
package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/yourusername/clip-tagger/scanner"
	"github.com/yourusername/clip-tagger/state"
	"github.com/yourusername/clip-tagger/ui"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: clip-tagger <directory>")
		os.Exit(1)
	}

	directory := os.Args[1]

	// Check for existing state
	hasState := state.StateExists(directory)
	var existingState *state.State
	var err error

	if hasState {
		existingState, err = state.Load(state.StateFilePath(directory))
		if err != nil {
			fmt.Printf("Error loading state: %v\n", err)
			os.Exit(1)
		}
	} else {
		existingState = state.NewState(directory, state.SortByModifiedTime)
	}

	// Scan directory
	sc := scanner.NewScanner(directory)
	scanResult, err := sc.Scan(scanner.SortBy(existingState.SortBy))
	if err != nil {
		fmt.Printf("Error scanning: %v\n", err)
		os.Exit(1)
	}

	// Get filenames
	var filenames []string
	for _, f := range scanResult.Files {
		filenames = append(filenames, f.Name)
	}

	// Merge with existing state
	mergeResult := state.MergeFiles(existingState, filenames)

	// Create UI
	model := ui.NewModel(
		directory,
		filenames,
		hasState,
		len(mergeResult.NewFiles),
		mergeResult.ExistingCount,
		len(mergeResult.MissingFiles),
	)

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
```

**Step 4: Test**

Run:
```bash
go build -o clip-tagger .
./clip-tagger .
```

Expected: Shows startup screen with file counts

**Step 5: Commit**

```bash
git add ui/ main.go
git commit -m "feat: add startup screen with state detection"
```

---

## Task 10: Classification Screen - Basic Structure

**Files:**
- Create: `ui/classification.go`
- Modify: `ui/model.go` (integrate classification screen)

**Step 1: Create classification screen structure**

```go
// ui/classification.go
package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ClassificationModel handles the classification screen
type ClassificationModel struct {
	files        []string
	currentIndex int
	totalFiles   int
	lastGroupID  string
	width        int
	height       int
}

func NewClassificationModel(files []string, startIndex int) ClassificationModel {
	return ClassificationModel{
		files:        files,
		currentIndex: startIndex,
		totalFiles:   len(files),
	}
}

func (m ClassificationModel) CurrentFile() string {
	if m.currentIndex >= 0 && m.currentIndex < len(m.files) {
		return m.files[m.currentIndex]
	}
	return ""
}

func (m ClassificationModel) Update(msg tea.Msg) (ClassificationModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			// Same as last group (TODO)
			return m, nil
		case "2":
			// Choose existing group (TODO)
			return m, nil
		case "3":
			// Create new group (TODO)
			return m, nil
		case "p":
			// Preview file (TODO)
			return m, nil
		case "left":
			if m.currentIndex > 0 {
				m.currentIndex--
			}
			return m, nil
		case "right":
			if m.currentIndex < m.totalFiles-1 {
				m.currentIndex++
			}
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m ClassificationModel) View() string {
	var b strings.Builder

	// Progress
	b.WriteString(fmt.Sprintf("\n  Clip %d of %d\n\n", m.currentIndex+1, m.totalFiles))

	// Current file
	currentFile := m.CurrentFile()
	if currentFile != "" {
		b.WriteString(fmt.Sprintf("  File: %s\n\n", currentFile))
	}

	// Actions
	b.WriteString("  Actions:\n")
	b.WriteString("    [1] Same as last group\n")
	b.WriteString("    [2] Choose existing group\n")
	b.WriteString("    [3] Create new group\n")
	b.WriteString("    [p] Preview file\n")
	b.WriteString("    [←/→] Navigate\n")
	b.WriteString("    [q] Save and quit\n")

	return b.String()
}
```

**Step 2: Integrate into main model**

Update `ui/model.go`:

```go
// Add to Model struct
type Model struct {
	screen             Screen
	directory          string
	files              []string
	currentIndex       int
	width              int
	height             int
	err                error
	startupModel       StartupModel
	classificationModel ClassificationModel // Add this
}

// Update NewModel
func NewModel(directory string, files []string, hasState bool, newFiles, existing, missing int) Model {
	return Model{
		screen:              ScreenStartup,
		directory:           directory,
		files:               files,
		currentIndex:        0,
		startupModel:        NewStartupModel(hasState, newFiles, existing, missing),
		classificationModel: NewClassificationModel(files, 0),
	}
}

// Update Update method
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || (msg.String() == "q" && m.screen != ScreenClassification) {
			return m, tea.Quit
		}
	case StartClassificationMsg:
		m.screen = ScreenClassification
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Delegate to screen-specific models
	switch m.screen {
	case ScreenStartup:
		var cmd tea.Cmd
		m.startupModel, cmd = m.startupModel.Update(msg)
		return m, cmd
	case ScreenClassification:
		var cmd tea.Cmd
		m.classificationModel, cmd = m.classificationModel.Update(msg)
		return m, cmd
	}

	return m, nil
}

// Update View method
func (m Model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error() + "\n"
	}

	switch m.screen {
	case ScreenStartup:
		return m.startupModel.View()
	case ScreenClassification:
		return m.classificationModel.View()
	default:
		return "Unknown screen\n"
	}
}
```

**Step 3: Test navigation**

Run:
```bash
go build -o clip-tagger .
./clip-tagger .
```

Expected: After startup, shows classification screen with navigation working

**Step 4: Commit**

```bash
git add ui/
git commit -m "feat: add classification screen with basic navigation"
```

---

## Task 11: File Preview Integration

**Files:**
- Create: `ui/preview.go`
- Modify: `ui/classification.go` (integrate preview)

**Step 1: Create preview functionality**

```go
// ui/preview.go
package ui

import (
	"fmt"
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"
)

// PreviewCompleteMsg signals preview window was opened
type PreviewCompleteMsg struct {
	Err error
}

// OpenPreview opens a file in the system default application
func OpenPreview(filepath string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd

		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", filepath)
		case "linux":
			cmd = exec.Command("xdg-open", filepath)
		case "windows":
			cmd = exec.Command("cmd", "/c", "start", filepath)
		default:
			return PreviewCompleteMsg{
				Err: fmt.Errorf("unsupported platform: %s", runtime.GOOS),
			}
		}

		err := cmd.Start()
		return PreviewCompleteMsg{Err: err}
	}
}
```

**Step 2: Integrate preview into classification**

Update `ui/classification.go`:

```go
import (
	"path/filepath"
)

// Add to ClassificationModel
type ClassificationModel struct {
	files        []string
	directory    string // Add this
	currentIndex int
	totalFiles   int
	lastGroupID  string
	width        int
	height       int
	previewError string // Add this
}

func NewClassificationModel(directory string, files []string, startIndex int) ClassificationModel {
	return ClassificationModel{
		files:        files,
		directory:    directory,
		currentIndex: startIndex,
		totalFiles:   len(files),
	}
}

// Update Update method
func (m ClassificationModel) Update(msg tea.Msg) (ClassificationModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "p":
			currentFile := m.CurrentFile()
			if currentFile != "" {
				fullPath := filepath.Join(m.directory, currentFile)
				return m, OpenPreview(fullPath)
			}
			return m, nil
		// ... rest of cases
		}
	case PreviewCompleteMsg:
		if msg.Err != nil {
			m.previewError = msg.Err.Error()
		} else {
			m.previewError = ""
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// Update View method to show preview errors
func (m ClassificationModel) View() string {
	var b strings.Builder

	// Progress
	b.WriteString(fmt.Sprintf("\n  Clip %d of %d\n\n", m.currentIndex+1, m.totalFiles))

	// Current file
	currentFile := m.CurrentFile()
	if currentFile != "" {
		b.WriteString(fmt.Sprintf("  File: %s\n\n", currentFile))
	}

	// Preview error
	if m.previewError != "" {
		b.WriteString(fmt.Sprintf("  ⚠ Preview error: %s\n\n", m.previewError))
	}

	// ... rest of view
}
```

**Step 3: Update model.go to pass directory**

Update `ui/model.go`:

```go
func NewModel(directory string, files []string, hasState bool, newFiles, existing, missing int) Model {
	return Model{
		screen:              ScreenStartup,
		directory:           directory,
		files:               files,
		currentIndex:        0,
		startupModel:        NewStartupModel(hasState, newFiles, existing, missing),
		classificationModel: NewClassificationModel(directory, files, 0), // Pass directory
	}
}
```

**Step 4: Test preview**

Create a test video file and run:
```bash
touch test.mp4
go build -o clip-tagger .
./clip-tagger .
```

Press 'p' to preview

Expected: Opens system video player

**Step 5: Commit**

```bash
git add ui/
git commit -m "feat: add file preview with system default player"
```

---

## Task 12: Group Selection Screen with Filtering

**Files:**
- Create: `ui/groupselect.go`
- Modify: `ui/model.go` (integrate group select)
- Modify: `ui/messages.go` (add messages)

**Step 1: Add messages**

Update `ui/messages.go`:

```go
// ShowGroupSelectMsg signals to show group selection
type ShowGroupSelectMsg struct{}

// GroupSelectedMsg signals a group was selected
type GroupSelectedMsg struct {
	GroupID string
}

// CancelSelectionMsg signals selection was cancelled
type CancelSelectionMsg struct{}
```

**Step 2: Create group selection screen**

```go
// ui/groupselect.go
package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// GroupItem represents a group in the list
type GroupItem struct {
	id        string
	name      string
	order     int
	takeCount int
}

func (g GroupItem) FilterValue() string { return g.name }
func (g GroupItem) Title() string {
	return fmt.Sprintf("[%02d] %s (%d takes)", g.order, g.name, g.takeCount)
}
func (g GroupItem) Description() string { return "" }

// GroupSelectModel handles group selection
type GroupSelectModel struct {
	list   list.Model
	width  int
	height int
}

func NewGroupSelectModel(groups []GroupItem) GroupSelectModel {
	items := make([]list.Item, len(groups))
	for i, g := range groups {
		items[i] = g
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select Group"
	l.SetFilteringEnabled(true)

	return GroupSelectModel{
		list: l,
	}
}

func (m GroupSelectModel) Update(msg tea.Msg) (GroupSelectModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			selected := m.list.SelectedItem()
			if selected != nil {
				group := selected.(GroupItem)
				return m, func() tea.Msg {
					return GroupSelectedMsg{GroupID: group.id}
				}
			}
			return m, nil
		case "esc":
			return m, func() tea.Msg { return CancelSelectionMsg{} }
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-4)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m GroupSelectModel) View() string {
	return "\n" + m.list.View() + "\n\n  [enter] select  [esc] cancel  [/] filter\n"
}
```

**Step 3: Integrate into main model**

Update `ui/model.go`:

```go
// Add to Model struct
type Model struct {
	screen              Screen
	directory           string
	files               []string
	currentIndex        int
	width               int
	height              int
	err                 error
	startupModel        StartupModel
	classificationModel ClassificationModel
	groupSelectModel    GroupSelectModel // Add this
	previousScreen      Screen           // Add this for navigation back
}

// Update Update method
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case StartClassificationMsg:
		m.screen = ScreenClassification
		return m, nil
	case ShowGroupSelectMsg:
		// Create group select with dummy data for now
		groups := []GroupItem{
			{id: "1", name: "intro", order: 1, takeCount: 2},
			{id: "2", name: "magic trick", order: 2, takeCount: 3},
		}
		m.previousScreen = m.screen
		m.screen = ScreenGroupSelect
		m.groupSelectModel = NewGroupSelectModel(groups)
		return m, nil
	case GroupSelectedMsg:
		// TODO: Handle group selection
		m.screen = m.previousScreen
		return m, nil
	case CancelSelectionMsg:
		m.screen = m.previousScreen
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	// Delegate to screen-specific models
	switch m.screen {
	case ScreenStartup:
		var cmd tea.Cmd
		m.startupModel, cmd = m.startupModel.Update(msg)
		return m, cmd
	case ScreenClassification:
		var cmd tea.Cmd
		m.classificationModel, cmd = m.classificationModel.Update(msg)
		return m, cmd
	case ScreenGroupSelect:
		var cmd tea.Cmd
		m.groupSelectModel, cmd = m.groupSelectModel.Update(msg)
		return m, cmd
	}

	return m, nil
}

// Update View method
func (m Model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error() + "\n"
	}

	switch m.screen {
	case ScreenStartup:
		return m.startupModel.View()
	case ScreenClassification:
		return m.classificationModel.View()
	case ScreenGroupSelect:
		return m.groupSelectModel.View()
	default:
		return "Unknown screen\n"
	}
}
```

**Step 4: Trigger group select from classification**

Update `ui/classification.go`:

```go
func (m ClassificationModel) Update(msg tea.Msg) (ClassificationModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "2":
			// Choose existing group
			return m, func() tea.Msg { return ShowGroupSelectMsg{} }
		// ... rest
		}
	}
	// ...
}
```

**Step 5: Test group selection**

Run:
```bash
go build -o clip-tagger .
./clip-tagger .
```

Navigate to classification, press '2'

Expected: Shows group selection screen with filtering

**Step 6: Commit**

```bash
git add ui/
git commit -m "feat: add group selection screen with filtering"
```

---

Due to length constraints, I'll summarize the remaining tasks:

## Task 13: Group Insertion Screen
- Create `ui/groupinsert.go` with list selection and name input
- Allow choosing position "after X" with filtering
- Integrate into model

## Task 14: State Integration with UI
- Thread state through all UI components
- Save state after each classification
- Load existing classifications on startup
- Handle missing/new files

## Task 15: Classification Logic
- Implement "same as last group" action
- Implement group selection handler
- Implement new group creation
- Update take numbers automatically
- Save state after each action

## Task 16: Review Screen
- Create `ui/review.go`
- Generate rename list from state
- Show changes with tags ([new], [updated], [moved])
- Display skipped files
- Allow scrolling through list

## Task 17: Rename Confirmation
- Add operation mode selection (rename/copy)
- Run conflict detection
- Show conflicts if any
- Final confirmation prompt
- Execute operations with progress

## Task 18: CLI Flags
- Add flag parsing with stdlib `flag` package
- Support --sort-by, --reset, --clean-missing, --preview
- Update main.go to handle flags

## Task 19: Integration Testing
- Create test scenarios with sample videos
- Test full flow: scan -> classify -> rename
- Test resume functionality
- Test conflict detection

## Task 20: Documentation
- Create README.md with usage examples
- Document keyboard shortcuts
- Add installation instructions

---

Would you like me to continue with the detailed tasks 13-20, or would you prefer to start implementing from Task 1?
