# Cliq

AI-powered CLI assistant for Neovim and tmux. Get instant help with commands, keybindings, and workflows — all running locally with complete privacy.

## Features

- **Privacy-first**: Runs entirely locally using a small language model (Phi-3-mini). No data leaves your machine.
- **Configuration-aware**: Parses your Neovim and tmux configs to provide personalized responses including your custom keymaps.
- **Fast**: Optimized for quick responses. Get help without breaking your flow.
- **Interactive mode**: Full TUI for exploring commands and keybindings.
- **Cross-platform**: Works on macOS (Intel & Apple Silicon) and Linux.

## Quick Start

### Installation

**Using the install script (recommended):**
```bash
curl -sSL https://raw.githubusercontent.com/cliq-cli/cliq/main/scripts/install.sh | bash
```

**Using Homebrew (macOS/Linux):**
```bash
brew install cliq-cli/tap/cliq
```

**From source:**
```bash
git clone https://github.com/cliq-cli/cliq.git
cd cliq
make install
```

### Setup

Initialize Cliq to download the language model and detect your configurations:

```bash
cliq init
```

This will:
1. Download the Phi-3-mini model (~2.3GB)
2. Detect your Neovim and tmux configuration files
3. Create the initial configuration

### Usage

**Ask a question:**
```bash
cliq "how do I delete a line in vim"
cliq "split tmux window vertically"
cliq "search and replace in visual mode"
```

**Interactive mode:**
```bash
cliq -i
```

**View your parsed configuration:**
```bash
cliq config show
cliq config show nvim
cliq config show tmux
```

## Example Output

```
$ cliq "how do I delete a line"

💡 Command

  dd

Delete the current line in normal mode. The deleted content is stored in the
default register, so you can paste it with 'p'.

Alternatives:
  • D: Delete from cursor to end of line
  • d$: Same as D
  • dj: Delete current line and line below

📍 In your setup:
  <leader>d -> dd (Quick delete line)

🔗 Related:
  • yy: Yank (copy) current line
  • cc: Change (delete and enter insert mode) current line

💬 Tip: Use a count prefix like '3dd' to delete 3 lines at once.
```

## Commands

| Command | Description |
|---------|-------------|
| `cliq init` | Initialize Cliq (download model, detect configs) |
| `cliq [query]` | Ask a question about Neovim or tmux |
| `cliq -i` | Launch interactive TUI mode |
| `cliq config show` | Show parsed configuration |
| `cliq config reload` | Reload and re-parse configs |
| `cliq config edit` | Open config file in editor |
| `cliq version` | Show version information |

## Configuration

Cliq stores its configuration in `~/.config/cliq/config.toml`:

```toml
[general]
response_style = "concise"  # concise, detailed, minimal

[model]
path = "~/.local/share/cliq/model/phi-3-mini-q4.gguf"
temperature = 0.7
max_tokens = 512

[nvim]
config_path = "~/.config/nvim"
auto_detect = true
parse_plugins = true

[tmux]
config_path = "~/.tmux.conf"
auto_detect = true

[cache]
enabled = true
ttl_hours = 24
```

## How It Works

1. **Local LLM**: Cliq uses Phi-3-mini (3.8B parameters, Q4 quantization) for inference. The model runs entirely on your machine.

2. **Config Parsing**: Cliq parses your Neovim Lua/Vimscript configs and tmux.conf to understand your custom keymaps and plugins.

3. **Context-Aware Responses**: When you ask a question, Cliq includes relevant information about your setup in its response.

## File Locations

| Path | Description |
|------|-------------|
| `~/.config/cliq/config.toml` | User configuration |
| `~/.local/share/cliq/model/` | Downloaded language model |
| `~/.cache/cliq/` | Parsed config cache |

## Privacy

Cliq is designed with privacy as a core principle:

- **No telemetry**: Zero analytics or tracking
- **Local-only**: All processing happens on your machine
- **No network calls**: After initial model download, Cliq works entirely offline
- **Open source**: Full code transparency

## Requirements

- macOS 10.15+ or Linux
- ~3GB disk space for the model
- ~8GB RAM recommended for inference

## Building from Source

```bash
# Clone the repository
git clone https://github.com/cliq-cli/cliq.git
cd cliq

# Build
make build

# Run tests
make test

# Install locally
make install
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- [Phi-3](https://huggingface.co/microsoft/Phi-3-mini-4k-instruct) by Microsoft
- [llama.cpp](https://github.com/ggerganov/llama.cpp) for efficient inference
- [Cobra](https://github.com/spf13/cobra) and [Viper](https://github.com/spf13/viper) for CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the TUI
