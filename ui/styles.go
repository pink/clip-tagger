package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	colorPrimary   = lipgloss.Color("#7C3AED")
	colorSecondary = lipgloss.Color("#EC4899")
	colorSuccess   = lipgloss.Color("#10B981")
	colorWarning   = lipgloss.Color("#F59E0B")
	colorDanger    = lipgloss.Color("#EF4444")
	colorMuted     = lipgloss.Color("#6B7280")

	// Styles
	styleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1)

	styleSubheader = lipgloss.NewStyle().
			Foreground(colorSecondary).
			MarginBottom(1)

	styleKeyHint = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true)

	styleSuccess = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true)

	styleWarning = lipgloss.NewStyle().
			Foreground(colorWarning).
			Bold(true)

	styleDanger = lipgloss.NewStyle().
			Foreground(colorDanger).
			Bold(true)

	styleHighlight = lipgloss.NewStyle().
				Foreground(colorPrimary).
				Bold(true)

	styleCursor = lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true)

	styleTag = lipgloss.NewStyle().
			Foreground(colorMuted).
			Background(lipgloss.Color("#1F2937")).
			Padding(0, 1)

	styleMuted = lipgloss.NewStyle().
			Foreground(colorMuted)
)

// Helper functions
func RenderHeader(text string) string {
	return styleHeader.Render(text)
}

func RenderSubheader(text string) string {
	return styleSubheader.Render(text)
}

func RenderKeyHint(text string) string {
	return styleKeyHint.Render(text)
}

func RenderSuccess(text string) string {
	return styleSuccess.Render(text)
}

func RenderWarning(text string) string {
	return styleWarning.Render(text)
}

func RenderDanger(text string) string {
	return styleDanger.Render(text)
}

func RenderHighlight(text string) string {
	return styleHighlight.Render(text)
}

func RenderCursor(text string) string {
	return styleCursor.Render(text)
}

func RenderMuted(text string) string {
	return styleMuted.Render(text)
}

func RenderTag(text string, tagType string) string {
	switch tagType {
	case "new":
		return styleSuccess.Render("[" + text + "]")
	case "moved":
		return styleWarning.Render("[" + text + "]")
	case "updated":
		return styleHighlight.Render("[" + text + "]")
	default:
		return styleTag.Render("[" + text + "]")
	}
}

func RenderProgress(current, total int) string {
	percentage := float64(current) / float64(total) * 100
	return styleHighlight.Render(fmt.Sprintf("Progress: %d/%d (%.0f%%)", current, total, percentage))
}
