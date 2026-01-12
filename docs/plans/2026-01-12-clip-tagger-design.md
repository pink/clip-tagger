# clip-tagger Design Document

**Date:** 2026-01-12
**Status:** Approved

## Overview

clip-tagger is a CLI tool for interactively reviewing and renaming video files into a structured naming convention. It helps organize hundreds of video clips by grouping them semantically and numbering them chronologically.

**Target naming format:** `[XX_YY] group-name.ext`
- `XX` = group number (sequential)
- `YY` = take number within group
- Enables chronological sorting and semantic grouping

## Use Case

Users film many video clips (100+) with auto-generated filenames. They need to review each clip and organize them into semantic groups (e.g., "intro", "magic trick", "outro") while maintaining chronological order and tracking multiple takes.

## Architecture

### Core Components

1. **File Scanner**
   - Discovers video files: `.mp4`, `.mov`, `.avi`, `.mkv`, `.webm`
   - Sorts by modified time (default), create time, or name
   - Detects new files added since last session

2. **State Manager**
   - Persists work-in-progress to `.clip-tagger-state.json`
   - Auto-saves after each classification
   - Supports pause/resume workflow
   - Tracks groups, classifications, current position, skipped files

3. **Bubbletea UI**
   - Interactive TUI with multiple screens
   - Classification screen (main loop)
   - Group selection with built-in filtering
   - Group insertion with position selection
   - Final review screen
   - Uses `bubbles/list` and `bubbles/progress` components

4. **Renamer**
   - Generates target filenames from state
   - Checks for conflicts before operations
   - Supports rename-in-place or copy-to-directory
   - Tracks changes for review display

### Data Flow

1. Scan directory → build file list
2. Load existing state (if any) → merge with file list, detect new/missing files
3. Enter interactive classification mode → user classifies clips → save state after each change
4. Final review → show all changes including renumbering → conflict check
5. User chooses rename or copy → execute operations with progress bar

## State Management

### State File Format

`.clip-tagger-state.json`:
```json
{
  "directory": "/path/to/videos",
  "sort_by": "modified_time",
  "current_index": 5,
  "groups": [
    {"id": "uuid1", "name": "intro", "order": 1},
    {"id": "uuid2", "name": "magic trick", "order": 2},
    {"id": "uuid3", "name": "outro", "order": 3}
  ],
  "classifications": [
    {"file": "VID001.mp4", "group_id": "uuid1", "take_number": 1},
    {"file": "VID002.mp4", "group_id": "uuid2", "take_number": 1},
    {"file": "VID003.mp4", "group_id": "uuid2", "take_number": 2}
  ],
  "skipped": ["VID999.mp4"]
}
```

### Key Design Decisions

- **Groups have UUIDs**: Renaming a group doesn't break references
- **Separate order field**: Easy to reorder groups without changing IDs
- **Current index tracking**: Resume exactly where you left off
- **Auto-increment take numbers**: Calculated within each group
- **Saves after every action**: Crash-safe, not just on quit

### New File Detection

When resuming with existing state:
- Compare current directory scan vs state file
- New files added to classification queue in chronological position (based on sort order)
- Missing files marked as skipped with warning
- Show summary: "Found 12 new files since last session. 45 already classified."
- Adjust `current_index` if new files inserted before current position

## UI Flow

### 1. Startup Screen

- Detect existing state → prompt "Resume or start fresh?"
- If starting fresh, ask for sort order
- Scan directory and show file count
- Display new/missing file summary if resuming

### 2. Classification Screen (Main Loop)

**Display:**
- Top: Progress indicator (e.g., "Clip 5 of 127")
- Middle: Current filename and prompt
- Bottom: Available actions

**Actions:**
- `1` - Same as last group (adds as next take)
- `2` - Choose existing group → fuzzy search screen
- `3` - Create new group → insertion screen
- `←/→` - Navigate to previous/next clip
- `p` - Preview clip (opens system default player)
- `q` - Save and quit

**Behavior:**
- Can navigate backwards to review/change previous classifications
- Reclassifying a clip updates group take numbers automatically
- Auto-saves state after each change

### 3. Group Selection Screen

Uses `bubbles/list` with built-in filtering:
- Shows groups: `[01] intro (2 takes)`, `[02] magic trick (3 takes)`
- Press `/` or start typing to filter
- Arrow keys to navigate, Enter to confirm, Esc to cancel
- Groups implement `FilterValue()` for searchability

### 4. Group Insertion Screen

For creating new groups:
- Shows ordered list of existing groups with filtering
- Select position: "After [group name]"
- Default option: "Append to end"
- Then prompts for new group name
- Validates no duplicate names (case-insensitive)

### 5. Final Review Screen

**Display format:**
```
[01_01] intro.mp4          (was: VID001.mp4) [new]
[02_01] magic trick.mp4    (was: VID002.mp4)
[02_02] magic trick.mp4    (was: VID003.mp4)
[02_03] magic trick.mp4    (was: VID004.mp4)
[03_01] outro.mp4          (was: [02_01] outro.mp4) [updated]

Skipped files (1):
- VID999.mp4 (failed to open)
```

**Change tags:**
- `[new]` - Newly classified file
- `[updated]` - Group number changed due to reordering
- `[moved]` - File reclassified to different group
- No tag - Already correctly named

**Operations:**
- Arrow keys to scroll through preview
- Highlights any classification gaps
- Choose operation mode:
  - Rename in place (default)
  - Copy to output directory (prompts for path, default: `./tagged`)

**Conflict Detection:**
- Check if any target filenames already exist
- If conflicts found: list them and abort (no changes made)
- User must resolve manually and re-run

**Final Confirmation:**
- Prompt: "Proceed with [rename/copy] of 125 files? (y/n)"
- Show progress bar during operations
- On completion: summary of changes

## Error Handling

### File Operation Errors

- **Missing file on preview**: Show warning, mark as skipped, continue
- **File deleted between scan and rename**: Mark skipped, continue
- **Rename/copy fails mid-operation**: Stop immediately, log successes, preserve state
- **Preview fails to open**: Mark as skipped, show warning, continue

### State Management Errors

- **Directory changed since last session**: Warn and offer to start fresh
- **Corrupted state file**: Backup to `.bak`, start fresh
- **Lock file**: `.clip-tagger.lock` prevents multiple instances on same directory

### Navigation Edge Cases

- First clip: left arrow does nothing
- Last clip: right arrow does nothing
- All clips classified: show "All done!", jump to review screen

### Group Management

- Deleting last clip from a group removes the group
- Renaming a group updates all references via UUID
- Duplicate group names prevented (case-insensitive)

## CLI Interface

### Command Structure

```bash
clip-tagger [directory] [flags]
```

### Flags

- `--sort-by` - Sort order: `modified` (default), `created`, `name`
- `--reset` - Ignore existing state, start fresh
- `--clean-missing` - Remove references to missing files from state
- `--preview` - Show what would be renamed without interactive mode
- `--help` - Show usage information

### Example Usage

```bash
# Start new session in current directory
clip-tagger .

# Resume previous session with explicit path
clip-tagger ~/Videos/project

# Preview changes without interactive mode
clip-tagger ~/Videos/project --preview

# Clean up references to deleted files
clip-tagger . --clean-missing

# Start fresh, ignoring previous state
clip-tagger . --reset --sort-by created
```

## Dependencies

- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles/list` - List component with filtering
- `github.com/charmbracelet/bubbles/progress` - Progress bar
- `github.com/google/uuid` - Group ID generation
- Standard library for file operations

## Project Structure

```
clip-tagger/
├── main.go              # Entry point, CLI parsing
├── scanner/             # File discovery and sorting
│   └── scanner.go
├── state/               # State persistence and management
│   ├── state.go
│   └── lock.go
├── ui/                  # Bubbletea models and screens
│   ├── classification.go
│   ├── review.go
│   ├── groupselect.go
│   └── components.go
├── renamer/             # Filename generation and operations
│   ├── generator.go
│   └── operations.go
├── docs/
│   └── plans/
│       └── 2026-01-12-clip-tagger-design.md
└── go.mod
```

## Future Considerations (Out of Scope)

- Batch editing groups (rename, merge, split)
- Undo/redo functionality
- Export to different naming conventions
- Integration with video editing software
- Cloud storage support
- Collaborative tagging
