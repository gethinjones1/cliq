# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Cliq is an AI-powered CLI assistant for Neovim and tmux users. It runs locally using Phi-3-mini for privacy-first assistance with commands, keybindings, and workflows. Written in Go, it supports multiple LLM backends (ollama, llama-server, llama-cli).

## Build Commands

```bash
make build          # Build binary (CGO disabled by default)
make build-cgo      # Build with CGO (for llama.cpp)
make test           # Run all tests
make test-cover     # Tests with HTML coverage report
make lint           # Run golangci-lint
make fmt            # Format code with gofmt and goimports
make install        # Install to /usr/local/bin (requires sudo)
make install-user   # Install to ~/.local/bin
make uninstall      # Remove from /usr/local/bin
```

## Running Locally

```bash
go run . -v              # Run with verbose output
./cliq init              # Initialize (download model, detect configs)
./cliq "your question"   # Query mode
./cliq -i                # Interactive TUI mode
./cliq config show       # View parsed configurations
```

## Architecture

```
cmd/                    # Cobra CLI commands
  root.go              # Entry point, global flags, Viper config
  query.go             # Query execution pipeline
  init.go              # LLM backend setup, model download
  interactive.go       # Bubble Tea TUI
  config.go            # Config show/reload/edit commands

internal/
  config/              # Config struct (TOML) + XDG path resolution
  llm/                 # Multi-backend LLM client + prompt building
  parser/              # Neovim (Lua/Vimscript) and tmux config parsers
  response/            # Response parsing and formatting (text/JSON/markdown)
```

## Key Patterns

- **LLM Backend Abstraction**: `internal/llm/client.go` supports ollama, llama-server, and llama-cli with auto-detection fallback
- **Config Parsing**: Regex-based extraction of keymaps from Neovim configs (Lua + Vimscript) without executing user code
- **XDG Compliance**: All paths in `internal/config/paths.go` respect `$XDG_CONFIG_HOME`, `$XDG_DATA_HOME`, `$XDG_CACHE_HOME`
- **Version Injection**: Build flags inject version/commit/date via ldflags in Makefile
- **Prompt Engineering**: `internal/llm/prompts.go` contains Vim/tmux reference material and few-shot examples to ground the LLM and prevent hallucination. Small models need explicit examples.

## Response Quality

The system prompt in `internal/llm/prompts.go` includes:
- Vim/tmux command reference (motions, operators, counts)
- Few-shot examples showing correct response format
- Low temperature (0.3) to reduce hallucination

Default model is `mistral` (7B). To use a different model: `CLIQ_OLLAMA_MODEL=llama3 cliq "query"`

## File Locations at Runtime

- Config: `~/.config/cliq/config.toml`
- Model: `~/.local/share/cliq/model/phi-3-mini-q4.gguf`
- Cache: `~/.cache/cliq/config-cache.json`
