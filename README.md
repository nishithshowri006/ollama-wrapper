# Ollama Wrapper

Ollama Wrapper is a command-line interface (CLI) application that provides a user-friendly terminal UI for interacting with Ollama, a local LLM inference server.

## Features

- Interactive terminal UI built with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- Browse and select from available models in Ollama
- Chat with selected models via a clean, intuitive interface
- Automatic model loading and downloading
- Markdown rendering of responses
- Stream responses in real-time

## Installation

### Pre-built Binaries

Download the latest release for your platform from the [releases page](https://github.com/nishithshowri006/ollama-wrapper/releases).

### Build from Source

```bash
git clone https://github.com/nishithshowri006/ollama-wrapper.git
cd ollama-wrapper
go build -o ollama_wrapper
```

## Usage

1. Make sure Ollama is running on your system
2. Launch the application:

```bash
./ollama_wrapper
```

### Controls

- Navigation:
  - In model selection: Arrow keys to navigate, Enter to select a model
  - In chat view: Arrow keys/Page Up/Page Down to scroll through chat history

- Commands in chat:
  - `/clear` or `/clear()` - Clear the current conversation
  - `/back()` or `q` - Return to model selection
  - `/exit` or `/exit()` - Quit the application
  - `ESC` - Toggle focus between input and chat history

## Requirements

- [Ollama](https://ollama.ai/) running locally (default port: 11434)
- For building: Go 1.24 or later

## License

This project is licensed under the GPL-2.0 License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - The TUI framework
- [Ollama](https://ollama.ai/) - The local LLM server

## Contributing
