# clip-tagger
Whenever I shoot projects, I always end up with dozens to hundreds of clips. The default naming convention for most cameras is time-based (e.g. `iPhone001_12161034_C083.mov`). This has its use cases, but isn't useful to me when editing.

I created this CLI to help speed up a workflow I previously did manually. It essentially boils down to this naming convention:
```
[SequenceNumber_TakeNumber] GroupName.mov
```

Imagine you had 5 video files. The first one might be an intro, the next 3 might be multiple takes of you doing a magic trick, and the last an outro.

With my naming convention, I would end up with a list of files that looked like this
```
[01_01] intro.mov
[02_01] magic trick.mov
[02_02] magic trick.mov
[02_03] magic trick.mov
[03_01] outro.mov
```

Why is this useful for me?
1. The number prefix lets me sort alphabetically and have the list in my preferred shot order
2. The `XX_YY` structure lets me denote multiple takes of the same shot
3. The group name lets me know at a glance what the file contains, or at least in a way that's relevant for me

I would do this for each video. Watching each one, and then renaming the file in this format. It was exhausting, and even more tedious if I ever wanted to insert a new group in the middle of the list!! (all sequence numbers +1...)

`clip-tagger` makes this so much easier.

I'm sure this workflow might evolve over time. With more complex projects, or with more subjects maybe, I could see this needing more granularity. But for now - it works for me 🙂.

## Getting Started
A couple of notes:
1. The CLI won't rename any files until you give it confirmation at the end, so it's relatively safe.
2. After every clip, the CLI auto-saves progress locally, so you can quit out/leave and resume later.

### 1) Start the CLI
In your terminal of choice, navigate to a directory that contains your video files and start the CLI:
```bash
clip-tagger .
```
<p align="center">
   <img width="572" height="358" alt="image" src="https://github.com/user-attachments/assets/f2b41c12-e6ef-44d4-9222-30632a1ca28e" />
</p>



### 2) Preview the current file
Pressing `p` will open up the current file in your default video player.

### 3) Categorize the clip
You have 4 choices:

**1 - Same as last**

`(1)` will assign the clip to the last used group name.

**2 - Select from existing groups**

`(2)` will let you choose from one of the other previously used groups:
<p align="center">
   <img width="572" height="356" alt="image" src="https://github.com/user-attachments/assets/ab6686b9-d2ed-4b00-9618-63dcd42ec2a5" />
</p>

**3 - Create new group**

`(3)` will let you create a new group and choose where in the list it gets added:
<p align="center">
   <img width="572" height="357" alt="image" src="https://github.com/user-attachments/assets/b091d55b-2a3f-41f2-830d-e24b78bbbd6d" />
</p>

**s - Skip this file**

`(s)` will mark the file as skipped. Useful in case you want to defer until the end or delete altogether.

### 4) Rinse and repeat
Do this until all of the files in your directory have been reviewed.

### 5) Finalize
The last step is executing the rename. You can either rename files in-place in the current directory, or have copies made in a new directory.

### 6) Adding new video files 
If you end up adding more files to the current project, you can just drop the files into the same directory and call the CLI again.

The CLI keeps track of what's already been classified, so you will be able to resume the workflow just focusing on the new files.

## Installation

### Homebrew (macOS - Recommended)

```bash
brew install pink/clip-tagger/clip-tagger
```

### Manual Installation

#### Download Pre-built Binary

1. Go to [Releases](https://github.com/pink/clip-tagger/releases)
2. Download the binary for your platform:
   - macOS (Apple Silicon): `clip-tagger_*_Darwin_arm64.tar.gz`
   - macOS (Intel): `clip-tagger_*_Darwin_x86_64.tar.gz`
   - Linux (64-bit): `clip-tagger_*_Linux_x86_64.tar.gz`
   - Windows (64-bit): `clip-tagger_*_Windows_x86_64.zip`
3. Extract and move to your PATH

#### Build from Source

If you have Go installed:

```bash
git clone https://github.com/pink/clip-tagger.git
cd clip-tagger
go install
```

### Verify Installation

```bash
clip-tagger --help
```

### Prerequisites

- A terminal emulator
- Default video player configured (for preview feature)

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

## State File

clip-tagger saves progress to `.clip-tagger-state.json` in the working directory.

The state file contains:
- Classification assignments
- Group definitions and order
- Current position in workflow
- Skipped files
- Sort preferences

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

## Acknowledgments

Built with:
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [google/uuid](https://github.com/google/uuid) - UUID generation
