package response

import (
	"encoding/json"
	"regexp"
	"strings"
)

// Response represents a parsed LLM response
type Response struct {
	Query        string   `json:"query,omitempty"`
	Command      string   `json:"command"`
	Explanation  string   `json:"explanation"`
	Alternatives []string `json:"alternatives,omitempty"`
	UserKeymaps  []string `json:"user_keymaps,omitempty"`
	Related      []string `json:"related,omitempty"`
	Tips         []string `json:"tips,omitempty"`
	TmuxPrefix   string   `json:"tmux_prefix,omitempty"`
	Raw          string   `json:"-"`
}

// Parse parses the LLM output into a structured Response
func Parse(llmOutput string) *Response {
	resp := &Response{
		Raw: llmOutput,
	}

	// Try to extract structured sections
	sections := extractSections(llmOutput)

	if cmd, ok := sections["command"]; ok {
		resp.Command = strings.TrimSpace(cmd)
	}

	if exp, ok := sections["explanation"]; ok {
		resp.Explanation = strings.TrimSpace(exp)
	}

	if alt, ok := sections["alternatives"]; ok {
		resp.Alternatives = parseList(alt)
	}

	if rel, ok := sections["related"]; ok {
		resp.Related = parseList(rel)
	}

	if tip, ok := sections["tip"]; ok {
		resp.Tips = []string{strings.TrimSpace(tip)}
	}
	if tips, ok := sections["tips"]; ok {
		resp.Tips = parseList(tips)
	}

	// If we couldn't parse structured sections, use the raw output
	if resp.Command == "" && resp.Explanation == "" {
		resp.Explanation = strings.TrimSpace(llmOutput)
	}

	return resp
}

// extractSections extracts labeled sections from the LLM output
func extractSections(text string) map[string]string {
	sections := make(map[string]string)

	// Patterns to look for
	patterns := []string{
		"command", "explanation", "alternatives", "alternative",
		"related", "tip", "tips", "example", "examples",
		"main commands", "navigation", "usage",
	}

	// Build regex pattern
	patternStr := `(?i)(?:^|\n)[\s]*(?:` + strings.Join(patterns, "|") + `)[:\s]*\n?`
	re := regexp.MustCompile(patternStr)

	// Find all matches
	matches := re.FindAllStringIndex(text, -1)

	if len(matches) == 0 {
		return sections
	}

	// Extract sections
	for i, match := range matches {
		start := match[0]
		end := len(text)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}

		// Get the section name
		headerEnd := match[1]
		header := strings.ToLower(strings.TrimSpace(text[start:headerEnd]))
		header = regexp.MustCompile(`[:\s]+$`).ReplaceAllString(header, "")
		header = strings.TrimSpace(header)

		// Normalize header names
		switch {
		case strings.Contains(header, "command"):
			header = "command"
		case strings.Contains(header, "explanation"):
			header = "explanation"
		case strings.Contains(header, "alternative"):
			header = "alternatives"
		case strings.Contains(header, "related"):
			header = "related"
		case strings.Contains(header, "tip"):
			header = "tip"
		}

		// Get the content
		content := text[headerEnd:end]
		sections[header] = strings.TrimSpace(content)
	}

	return sections
}

// parseList parses a section into a list of items
func parseList(text string) []string {
	var items []string

	// Split by newlines
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Remove common list prefixes
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "* ")
		line = strings.TrimPrefix(line, "• ")
		line = regexp.MustCompile(`^\d+\.\s*`).ReplaceAllString(line, "")

		if line != "" {
			items = append(items, line)
		}
	}

	return items
}

// ToJSON returns the response as JSON
func (r *Response) ToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToMarkdown returns the response as markdown
func (r *Response) ToMarkdown() string {
	var sb strings.Builder

	if r.Command != "" {
		sb.WriteString("## Command\n\n")
		sb.WriteString("```\n")
		sb.WriteString(r.Command)
		sb.WriteString("\n```\n\n")
	}

	if r.Explanation != "" {
		sb.WriteString("## Explanation\n\n")
		sb.WriteString(r.Explanation)
		sb.WriteString("\n\n")
	}

	if len(r.Alternatives) > 0 {
		sb.WriteString("## Alternatives\n\n")
		for _, alt := range r.Alternatives {
			sb.WriteString("- ")
			sb.WriteString(alt)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	if len(r.UserKeymaps) > 0 {
		sb.WriteString("## Your Keymaps\n\n")
		for _, km := range r.UserKeymaps {
			sb.WriteString("- ")
			sb.WriteString(km)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	if len(r.Related) > 0 {
		sb.WriteString("## Related\n\n")
		for _, rel := range r.Related {
			sb.WriteString("- ")
			sb.WriteString(rel)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	if len(r.Tips) > 0 {
		sb.WriteString("## Tips\n\n")
		for _, tip := range r.Tips {
			sb.WriteString("> ")
			sb.WriteString(tip)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// ToText returns the response as formatted plain text
func (r *Response) ToText() string {
	// If we have the raw output and couldn't parse it well, return it directly
	if r.Command == "" && r.Explanation == "" && r.Raw != "" {
		return r.Raw
	}

	// Otherwise format nicely
	return r.Raw
}
