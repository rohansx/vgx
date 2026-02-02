# VGX — AI Code Security Scanner

<p align="center">
  <img src="https://img.shields.io/visual-studio-marketplace/v/rohansx.vgx" alt="Version">
  <img src="https://img.shields.io/visual-studio-marketplace/i/rohansx.vgx" alt="Installs">
  <img src="https://img.shields.io/visual-studio-marketplace/r/rohansx.vgx" alt="Rating">
</p>

**Detect AI-generated code and security vulnerabilities in real-time.**

## Features

### 🤖 AI Code Detection
VGX highlights code that's likely AI-generated (Copilot, Cursor, Claude, etc.):

- Status bar shows AI confidence percentage
- Inline highlighting of AI patterns
- Workspace-wide scanning

### 🔒 Security Scanning (Coming Soon)
- Hardcoded secrets detection
- SQL injection patterns
- XSS vulnerabilities

## Usage

1. Open any code file
2. Check the status bar for AI confidence: `🤖 87% AI` or `✓ 12% AI`
3. Run `VGX: Detect AI Code in Workspace` to scan all files

## Commands

| Command | Description |
|---------|-------------|
| `VGX: Detect AI Code in File` | Analyze current file |
| `VGX: Detect AI Code in Workspace` | Scan all files |
| `VGX: Security Scan File` | Security analysis |

## Settings

| Setting | Default | Description |
|---------|---------|-------------|
| `vgx.aiThreshold` | 70 | Confidence threshold (0-100) |
| `vgx.highlightAICode` | true | Highlight AI patterns |
| `vgx.showInlineConfidence` | true | Show inline indicators |

## How It Works

VGX uses stylometry and pattern analysis to detect AI-generated code:

- **Naming consistency** — AI uses very consistent naming conventions
- **Comment patterns** — Formulaic JSDoc, docstrings
- **Error handling** — Standard try-catch patterns
- **Code structure** — Perfect indentation, boilerplate

No data leaves your machine. 100% offline.

## Links

- [GitHub](https://github.com/rohansx/vgx)
- [CLI Tool](https://vgx.sh)
- [Issues](https://github.com/rohansx/vgx/issues)

## License

MIT
