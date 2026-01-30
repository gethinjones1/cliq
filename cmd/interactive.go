package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/cliq-cli/cliq/internal/config"
	"github.com/cliq-cli/cliq/internal/llm"
	"github.com/cliq-cli/cliq/internal/parser"
	"github.com/cliq-cli/cliq/internal/response"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("42")).
			Background(lipgloss.Color("235")).
			Padding(0, 1)

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	responseStyle = lipgloss.NewStyle().
			Padding(1, 2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
)

// Model represents the TUI application state
type model struct {
	textarea    textarea.Model
	viewport    viewport.Model
	spinner     spinner.Model
	history     []queryResult
	loading     bool
	err         error
	width       int
	height      int
	llmClient   *llm.Client
	nvimConfig  *parser.NvimConfig
	tmuxConfig  *parser.TmuxConfig
	ready       bool
}

type queryResult struct {
	Query    string
	Response string
}

// Messages
type responseMsg struct {
	response string
	err      error
}

type initMsg struct {
	client     *llm.Client
	nvimConfig *parser.NvimConfig
	tmuxConfig *parser.TmuxConfig
	err        error
}

func runInteractive() error {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Ask about Neovim or tmux commands..."
	ta.Focus()
	ta.CharLimit = 500
	ta.SetWidth(80)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))

	return model{
		textarea: ta,
		spinner:  s,
		history:  []queryResult{},
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		initLLM,
	)
}

func initLLM() tea.Msg {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}

	modelPath := cfg.GetModelPath()
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return initMsg{err: fmt.Errorf("model not found. Run 'cliq init' first")}
	}

	client, err := llm.NewClient(modelPath, cfg.Model.OllamaModel, cfg.Model.Temperature, cfg.Model.MaxTokens)
	if err != nil {
		return initMsg{err: fmt.Errorf("failed to load model: %w", err)}
	}

	// Parse configs
	var nvimConfig *parser.NvimConfig
	var tmuxConfig *parser.TmuxConfig

	if cfg.Nvim.ConfigPath != "" {
		nvimConfig, _ = parser.ParseNvimConfig(cfg.Nvim.ConfigPath)
	}
	if cfg.Tmux.ConfigPath != "" {
		tmuxConfig, _ = parser.ParseTmuxConfig(cfg.Tmux.ConfigPath)
	}

	return initMsg{
		client:     client,
		nvimConfig: nvimConfig,
		tmuxConfig: tmuxConfig,
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.llmClient != nil {
				m.llmClient.Close()
			}
			return m, tea.Quit

		case tea.KeyEnter:
			if !m.loading && m.ready {
				query := strings.TrimSpace(m.textarea.Value())
				if query != "" {
					m.loading = true
					m.textarea.Reset()
					return m, tea.Batch(
						m.spinner.Tick,
						m.queryLLM(query),
					)
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.textarea.SetWidth(msg.Width - 4)

		headerHeight := 3
		inputHeight := 5
		helpHeight := 2
		viewportHeight := msg.Height - headerHeight - inputHeight - helpHeight - 2

		if !m.ready {
			m.viewport = viewport.New(msg.Width-4, viewportHeight)
			m.viewport.SetContent(m.renderHistory())
			m.ready = true
		} else {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = viewportHeight
		}

	case initMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.llmClient = msg.client
			m.nvimConfig = msg.nvimConfig
			m.tmuxConfig = msg.tmuxConfig
			m.ready = true
		}

	case responseMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.history = append(m.history, queryResult{
				Query:    m.history[len(m.history)-1].Query,
				Response: msg.response,
			})
			m.viewport.SetContent(m.renderHistory())
			m.viewport.GotoBottom()
		}

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	// Update textarea
	if !m.loading {
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Update viewport
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) queryLLM(query string) tea.Cmd {
	// Add query to history first (response will be added when complete)
	m.history = append(m.history, queryResult{Query: query})

	return func() tea.Msg {
		prompt := llm.BuildPrompt(query, m.nvimConfig, m.tmuxConfig)
		resp, err := m.llmClient.Query(prompt)
		if err != nil {
			return responseMsg{err: err}
		}

		// Format response
		parsed := response.Parse(resp)
		return responseMsg{response: parsed.ToText()}
	}
}

func (m model) View() string {
	if m.err != nil {
		return errorStyle.Render(fmt.Sprintf("Error: %v\n\nPress Ctrl+C to exit.", m.err))
	}

	var b strings.Builder

	// Title
	title := titleStyle.Render(" Cliq - Interactive Mode ")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Response area
	if m.ready {
		b.WriteString(m.viewport.View())
	} else {
		b.WriteString("Loading model...")
	}
	b.WriteString("\n\n")

	// Loading indicator
	if m.loading {
		b.WriteString(m.spinner.View())
		b.WriteString(" Thinking...")
		b.WriteString("\n")
	}

	// Input area
	b.WriteString(promptStyle.Render("❯ "))
	b.WriteString(m.textarea.View())
	b.WriteString("\n")

	// Help
	help := helpStyle.Render("Enter: submit • Ctrl+C: quit • ↑↓: scroll")
	b.WriteString(help)

	return b.String()
}

func (m model) renderHistory() string {
	if len(m.history) == 0 {
		return helpStyle.Render("Welcome to Cliq! Ask me anything about Neovim or tmux.\n\nExamples:\n  • How do I delete a line?\n  • Split tmux window vertically\n  • Search and replace in vim")
	}

	var b strings.Builder
	for _, h := range m.history {
		b.WriteString(promptStyle.Render("❯ "))
		b.WriteString(h.Query)
		b.WriteString("\n\n")
		if h.Response != "" {
			b.WriteString(responseStyle.Render(h.Response))
			b.WriteString("\n\n")
		}
	}

	return b.String()
}
