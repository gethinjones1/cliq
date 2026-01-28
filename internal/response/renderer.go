package response

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Styles for terminal rendering
var (
	// CommandStyle for displaying commands
	CommandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	// SectionStyle for section headers
	SectionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("99")).
			Bold(true)

	// TipStyle for tips and hints
	TipStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Italic(true)

	// ExplanationStyle for explanations
	ExplanationStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	// KeymapStyle for user keymaps
	KeymapStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("141"))

	// DimStyle for less important text
	DimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	// IconCommand is the icon for command sections
	IconCommand = "💡"
	// IconTip is the icon for tips
	IconTip = "💬"
	// IconRelated is the icon for related commands
	IconRelated = "🔗"
	// IconUser is the icon for user-specific info
	IconUser = "📍"
)

// RenderResponse renders a response with terminal styling
func RenderResponse(resp *Response) string {
	var sb strings.Builder

	// Command section
	if resp.Command != "" {
		sb.WriteString(IconCommand)
		sb.WriteString(" ")
		sb.WriteString(SectionStyle.Render("Command"))
		sb.WriteString("\n\n")
		sb.WriteString("  ")
		sb.WriteString(CommandStyle.Render(resp.Command))
		sb.WriteString("\n\n")
	}

	// Explanation section
	if resp.Explanation != "" {
		sb.WriteString(ExplanationStyle.Render(resp.Explanation))
		sb.WriteString("\n\n")
	}

	// Alternatives section
	if len(resp.Alternatives) > 0 {
		sb.WriteString(SectionStyle.Render("Alternatives:"))
		sb.WriteString("\n")
		for _, alt := range resp.Alternatives {
			sb.WriteString("  ")
			sb.WriteString(DimStyle.Render("•"))
			sb.WriteString(" ")
			sb.WriteString(alt)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// User keymaps section
	if len(resp.UserKeymaps) > 0 {
		sb.WriteString(IconUser)
		sb.WriteString(" ")
		sb.WriteString(SectionStyle.Render("In your setup:"))
		sb.WriteString("\n")
		for _, km := range resp.UserKeymaps {
			sb.WriteString("  ")
			sb.WriteString(KeymapStyle.Render(km))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// Tmux prefix note
	if resp.TmuxPrefix != "" {
		sb.WriteString(DimStyle.Render("(Your tmux prefix: "))
		sb.WriteString(KeymapStyle.Render(resp.TmuxPrefix))
		sb.WriteString(DimStyle.Render(")"))
		sb.WriteString("\n\n")
	}

	// Related commands section
	if len(resp.Related) > 0 {
		sb.WriteString(IconRelated)
		sb.WriteString(" ")
		sb.WriteString(SectionStyle.Render("Related:"))
		sb.WriteString("\n")
		for _, rel := range resp.Related {
			sb.WriteString("  ")
			sb.WriteString(DimStyle.Render("•"))
			sb.WriteString(" ")
			sb.WriteString(rel)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// Tips section
	if len(resp.Tips) > 0 {
		sb.WriteString(IconTip)
		sb.WriteString(" ")
		sb.WriteString(SectionStyle.Render("Tip:"))
		sb.WriteString(" ")
		sb.WriteString(TipStyle.Render(resp.Tips[0]))
		sb.WriteString("\n")
	}

	return sb.String()
}

// RenderSimple renders a simple, non-styled response
func RenderSimple(resp *Response) string {
	var sb strings.Builder

	if resp.Command != "" {
		sb.WriteString("Command: ")
		sb.WriteString(resp.Command)
		sb.WriteString("\n\n")
	}

	if resp.Explanation != "" {
		sb.WriteString(resp.Explanation)
		sb.WriteString("\n\n")
	}

	if len(resp.Alternatives) > 0 {
		sb.WriteString("Alternatives:\n")
		for _, alt := range resp.Alternatives {
			sb.WriteString("  - ")
			sb.WriteString(alt)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	if len(resp.UserKeymaps) > 0 {
		sb.WriteString("In your setup:\n")
		for _, km := range resp.UserKeymaps {
			sb.WriteString("  - ")
			sb.WriteString(km)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	if len(resp.Related) > 0 {
		sb.WriteString("Related:\n")
		for _, rel := range resp.Related {
			sb.WriteString("  - ")
			sb.WriteString(rel)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	if len(resp.Tips) > 0 {
		sb.WriteString("Tip: ")
		sb.WriteString(resp.Tips[0])
		sb.WriteString("\n")
	}

	return sb.String()
}
