// flags/flags.go
package flags

import (
	"flag"
	"fmt"
	"os"
)

// Config holds parsed flag values
type Config struct {
	SortBy       string
	Reset        bool
	CleanMissing bool
	Preview      bool
	Help         bool
	Directory    string
}

// Parse parses command-line flags and returns config
func Parse() (*Config, error) {
	config := &Config{}

	// Define flags
	flag.StringVar(&config.SortBy, "sort-by", "", "Override default sort order (name, modified, created)")
	flag.BoolVar(&config.Reset, "reset", false, "Delete existing state and start fresh")
	flag.BoolVar(&config.CleanMissing, "clean-missing", false, "Remove missing files from state")
	flag.BoolVar(&config.Preview, "preview", false, "Show what would be renamed without executing")
	flag.BoolVar(&config.Help, "help", false, "Show usage information")

	// Custom usage function
	flag.Usage = PrintUsage

	// Parse flags
	flag.Parse()

	// Check if help flag is set
	if config.Help {
		PrintUsage()
		os.Exit(0)
	}

	// Get directory (required if not --help)
	args := flag.Args()
	if len(args) < 1 {
		return nil, fmt.Errorf("directory argument is required")
	}
	config.Directory = args[0]

	// Validate sort-by if specified
	if config.SortBy != "" {
		valid := config.SortBy == "name" || config.SortBy == "modified" || config.SortBy == "created"
		if !valid {
			return nil, fmt.Errorf("invalid sort-by value: %s (must be name, modified, or created)", config.SortBy)
		}
	}

	return config, nil
}

// PrintUsage prints usage information
func PrintUsage() {
	fmt.Fprintf(os.Stderr, `clip-tagger - Interactive video file classifier and renamer

Usage:
  clip-tagger [OPTIONS] <directory>

Arguments:
  <directory>    Path to directory containing video files

Options:
  --sort-by=<mode>     Override default sort order
                       Values: name, modified, created
                       Default: modified

  --reset              Delete existing state and start fresh
                       WARNING: This removes all previous classifications

  --clean-missing      Remove missing files from state
                       Useful if files were deleted since last session

  --preview            Show what would be renamed without executing
                       Displays the rename plan and exits

  --help               Show this help message

Examples:
  # Start tagging videos in current directory
  clip-tagger .

  # Sort by name instead of modified time
  clip-tagger --sort-by=name ./videos

  # Start fresh, deleting previous session
  clip-tagger --reset ./videos

  # Clean up missing files from previous session
  clip-tagger --clean-missing ./videos

  # Preview rename operations without executing
  clip-tagger --preview ./videos

For more information, see the documentation.
`)
}
