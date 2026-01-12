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
