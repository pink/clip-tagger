# clip-tagger

A terminal-based video file organization tool for bulk renaming video clips into semantic groups with take numbers.

## Overview

clip-tagger helps you organize hundreds of video clips by:
- Grouping clips by content/scene
- Assigning sequential take numbers within each group
- Generating structured filenames: `[XX_YY] GroupName.ext`
- Providing an interactive TUI for classification

Perfect for video editors, content creators, and anyone managing large collections of video clips.

## Features

- **Interactive TUI** - Easy-to-use terminal interface built with Bubbletea
- **Semantic Groups** - Organize clips by meaning (e.g., "Intro", "Scene 1", "B-Roll")
- **Take Numbers** - Automatically number multiple takes of the same scene
- **Resume Support** - Save progress and resume later
- **Auto-save** - State persists automatically as you work
- **Preview Files** - Open clips in your default video player
- **Flexible Sorting** - Sort by name, modified time, or created time
- **Conflict Detection** - Warns before overwriting existing files
- **Copy or Rename** - Choose to rename in place or copy to new directory

## Installation

### Prerequisites

- Go 1.21 or later
- A terminal emulator
- Default video player configured (for preview feature)

### Build from Source

```bash
git clone https://github.com/yourusername/clip-tagger.git
cd clip-tagger
go build -o clip-tagger .
```

### Install Globally

To install the binary globally so you can run `clip-tagger` from anywhere:

```bash
# From the project directory
go install

# Add to your ~/.zshrc (or ~/.bashrc on Linux) to make it permanent:
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc

# Then reload your shell:
source ~/.zshrc  # or source ~/.bashrc
```

**Verify Installation**

```bash
clip-tagger --help
```

## Usage

### Basic Usage

```bash
clip-tagger /path/to/video/directory
```

### Command-Line Flags

- `--help` - Show usage information
- `--sort-by=<mode>` - Sort files (name, modified, created)
- `--reset` - Delete existing state and start fresh
- `--clean-missing` - Remove missing files from state
- `--preview` - Show what would be renamed without executing

### Examples

```bash
# Start classifying clips in a directory
clip-tagger ./raw-clips

# Start fresh, ignoring previous state
clip-tagger --reset ./raw-clips

# Preview renames without executing
clip-tagger --preview ./raw-clips

# Remove missing files from state
clip-tagger --clean-missing ./raw-clips

# Sort by modified time instead of name
clip-tagger --sort-by=modified ./raw-clips
```

## Workflow

### 1. Startup Screen

Shows current state:
- New session or resume from previous classification
- Number of files found
- Files classified and remaining
- New/missing files since last session

**Actions:**
- `Enter` - Start/continue classification

### 2. Classification Screen

For each video file:
- Shows filename and progress (e.g., "File 5 of 23")
- Displays file path
- Shows previous group if available

**Actions:**
- `p` - Preview file in default player
- `1` - Same as last (assign to previous file's group)
- `2` - Select from existing groups
- `3` - Create new group
- `s` - Skip this file
- `q` or `Ctrl+C` - Quit

### 3. Group Selection Screen (Option 2)

Filter and select from existing groups:
- Type to filter groups (case-insensitive)
- Shows group order numbers

**Actions:**
- Type characters to filter
- `Up/Down` - Navigate groups
- `Enter` - Select group
- `Backspace` - Remove filter character
- `Esc` - Cancel and return

### 4. Group Insertion Screen (Option 3)

Create a new group:
1. **Name Entry Mode:**
   - Type the group name
   - `Enter` - Proceed to position selection
   - `Esc` - Cancel

2. **Position Selection Mode:**
   - Choose where to insert the group in the order
   - `Up/Down` - Navigate insertion positions
   - `Enter` - Confirm position
   - `Esc` - Back to name entry

### 5. Review Screen

Shows all renames before executing:
- Summary of classified and skipped files
- List of rename operations
- Change tags: `[new]`, `[moved]`, `[updated]`

**Actions:**
- `Up/Down` - Navigate list
- `Enter` - Proceed to rename
- `Esc` - Back to classification for more edits
- `q` or `Ctrl+C` - Quit

### 6. Completion Screen

Choose operation mode:
1. **Rename in place** - Rename files in current directory
2. **Copy to new directory** - Copy renamed files to timestamped folder

Shows conflict warnings if any files would be overwritten.

**Actions:**
- `Up/Down` - Select mode
- `Enter` - Execute operation
- `Esc` - Cancel and return to review
- `q` or `Ctrl+C` - Quit

## File Naming Convention

Generated filenames follow the pattern:

```
[XX_YY] GroupName.ext
```

Where:
- `XX` = Group order number (01, 02, 03...)
- `YY` = Take number within group (01, 02, 03...)
- `GroupName` = The semantic group name
- `.ext` = Original file extension

### Examples

```
[01_01] Intro.mp4      # First take of Intro group
[01_02] Intro.mp4      # Second take of Intro group
[02_01] Scene 1.mp4    # First take of Scene 1 group
[03_01] Outro.mov      # First take of Outro group
```

This format allows:
- Chronological sorting by group order
- Multiple takes per scene
- Clear semantic meaning
- Preservation of file extensions

## State File

clip-tagger saves progress to `.clip-tagger-state.json` in the working directory.

The state file contains:
- Classification assignments
- Group definitions and order
- Current position in workflow
- Skipped files
- Sort preferences

**Auto-save triggers:**
- After group selection/creation
- Every 5 classification actions
- When leaving classification screen

## Keyboard Shortcuts Summary

### Global
- `Ctrl+C` - Quit application
- `q` - Quit (most screens)

### Startup
- `Enter` - Start/continue classification

### Classification
- `p` - Preview file
- `1` - Same as last group
- `2` - Select existing group
- `3` - Create new group
- `s` - Skip file

### Group Selection
- Type - Filter groups
- `Up/Down` - Navigate
- `Enter` - Select
- `Backspace` - Delete filter character
- `Esc` - Cancel

### Group Insertion
- Type - Enter group name
- `Enter` - Proceed/confirm
- `Up/Down` - Navigate positions
- `Backspace` - Delete character
- `Esc` - Cancel/back

### Review
- `Up/Down` - Navigate list
- `Enter` - Proceed to completion
- `Esc` - Back to classification

### Completion
- `Up/Down` - Select mode
- `Enter` - Execute
- `Esc` - Cancel

## Supported File Formats

Video files with these extensions:
- `.mp4`
- `.mov`
- `.avi`
- `.mkv`
- `.webm`

## Troubleshooting

### Files not detected
- Ensure files have supported extensions
- Check file permissions
- Try `--reset` to start fresh

### State file corrupted
- Delete `.clip-tagger-state.json` manually
- Use `--reset` flag to start over

### Preview not working
- Verify default video player is configured
- Check file exists and is readable
- Platform-specific commands:
  - macOS: uses `open`
  - Linux: uses `xdg-open`
  - Windows: uses `start`

### Conflicts detected
- Review files that would be overwritten
- Use "Copy to new directory" mode
- Manually resolve conflicts before running

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./ui -v
```

### Project Structure

```
clip-tagger/
├── main.go              # Application entry point
├── flags/               # CLI flag parsing
├── preview/             # File preview functionality
├── renamer/             # Filename generation and operations
├── scanner/             # Directory scanning
├── state/               # State management and persistence
└── ui/                  # Terminal user interface
    ├── model.go         # Main Bubbletea model
    ├── startup.go       # Startup screen
    ├── classification.go # Classification screen
    ├── group_selection.go # Group selection
    ├── group_insertion.go # Group insertion
    ├── review.go        # Review screen
    └── completion.go    # Completion screen
```

## Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

[Your chosen license]

## Acknowledgments

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [google/uuid](https://github.com/google/uuid) - UUID generation
