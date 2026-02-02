# VGX

**Open-source security scanner for AI-assisted development.**

<p align="center">
  <img src="https://img.shields.io/badge/version-2.0.0-blue" alt="Version">
  <img src="https://img.shields.io/github/license/rohansx/vgx" alt="License">
  <img src="https://img.shields.io/github/stars/rohansx/vgx?style=social" alt="Stars">
</p>

<p align="center">
  <a href="#installation">Installation</a> &bull;
  <a href="#cli-reference">CLI Reference</a> &bull;
  <a href="#api">API</a> &bull;
  <a href="#vs-code-extension">VS Code</a> &bull;
  <a href="#how-it-works">How It Works</a>
</p>

---

## The Problem

With AI coding assistants (Copilot, Cursor, Claude), developers ship code they didn't write and often don't fully understand. Studies show:

- **42%** of code now has AI assistance ([Sonar 2026](https://www.sonarsource.com/))
- **96%** of developers don't fully trust AI-generated code
- **45%** of AI-generated code contains security vulnerabilities

VGX helps you **know what's AI-generated** and **catch vulnerabilities** before they hit production.

---

## Features

| Feature | Description |
|---------|-------------|
| **AI Code Detection** | Identify AI-generated code using stylometry + pattern analysis |
| **Security Scanning** | Vulnerability detection via Semgrep (+ optional OpenAI) |
| **Pre-commit Hooks** | Block insecure code before it's committed |
| **VS Code Extension** | Real-time AI detection in your editor |
| **Offline-First** | No API keys required. Your code never leaves your machine. |
| **CI/CD Ready** | JSON output for pipeline integration |

---

## Installation

### Quick Install (Linux/macOS)

```bash
curl -sSL https://vgx.sh/install | bash
```

### Go Install

```bash
go install github.com/rohansx/vgx@latest
```

### Docker

```bash
docker pull ghcr.io/rohansx/vgx:latest
docker run -v $(pwd):/code ghcr.io/rohansx/vgx detect --path /code
```

### Build from Source

```bash
git clone https://github.com/rohansx/vgx.git
cd vgx
go build -o vgx ./cmd/vgx
```

---

## CLI Reference

### Commands

```
vgx <command> [options]

Commands:
  detect    Detect AI-generated code
  scan      Security vulnerability scan
  version   Print version
  help      Show help
```

### `vgx detect` — AI Code Detection

Analyze files for AI-generated code patterns.

```bash
# Scan a directory
vgx detect --path ./src

# Scan a single file
vgx detect src/api/handler.ts

# JSON output (for CI/CD)
vgx detect --path ./src --format json

# Custom threshold (default: 70)
vgx detect --path ./src --threshold 80
```

**Options:**

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--path` | `-p` | `.` | Path to scan |
| `--format` | `-f` | `text` | Output format: `text`, `json` |
| `--threshold` | | `70` | AI confidence threshold (0-100) |

**Example Output:**

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  VGX AI Code Detection
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Files scanned:     12
  AI-generated:      4
  Human-written:     8
  AI percentage:     33.2%
  Max AI confidence: 89%

  FILES
     🤖 src/api/handlers.ts                         89%
     🤖 src/utils/fetch.ts                          82%
     ✓  src/config.ts                               28%

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  🤖 4 file(s) detected as AI-generated
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### `vgx scan` — Security Scanning

Scan for vulnerabilities using Semgrep rules.

```bash
# Scan changed files (default)
vgx scan

# Scan all files
vgx scan --changes=false

# Scan specific files
vgx scan src/auth.ts src/api/users.ts
```

**Options:**

| Flag | Default | Description |
|------|---------|-------------|
| `--changes` | `true` | Scan only git-changed files |
| `--report` | `true` | Generate markdown report |
| `--update-context` | `true` | Update file context cache |

**Detects:**
- Hardcoded secrets & API keys
- SQL injection vulnerabilities
- XSS (Cross-Site Scripting)
- Insecure cryptography
- Path traversal
- Command injection
- And more via [Semgrep rules](https://semgrep.dev/r)

---

## API

### JSON Output Schema

#### Detection Result

```json
{
  "files_scanned": 12,
  "ai_detected": 4,
  "human_written": 8,
  "max_ai_confidence": 0.89,
  "ai_percentage": 33.2,
  "results": [
    {
      "file_path": "src/api/handler.ts",
      "ai_confidence": 0.89,
      "confidence_level": "very_high",
      "style_score": 0.82,
      "pattern_score": 0.94,
      "is_ai_generated": true,
      "lines_of_code": 145,
      "patterns": [
        {
          "name": "async_await_fetch",
          "confidence": 0.10,
          "line_start": 23,
          "line_end": 35
        }
      ]
    }
  ]
}
```

#### Vulnerability Result

```json
{
  "file": "src/auth.ts",
  "description": "Hardcoded API key detected",
  "rule": "generic.secrets.security.detected-api-key",
  "severity": "high",
  "line": 15,
  "source": "semgrep",
  "recommendation": "Use environment variables for secrets"
}
```

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success, no issues found |
| `1` | Issues found (AI code or vulnerabilities) |
| `2` | Error (invalid path, etc.) |

---

## VS Code Extension

Real-time AI code detection in your editor.

### Features

- **Status Bar Indicator** — Shows AI confidence for current file
- **Inline Highlighting** — Highlights AI-generated code patterns
- **Workspace Scanning** — Scan entire workspace for AI code

### Commands

| Command | Description |
|---------|-------------|
| `VGX: Detect AI Code in File` | Analyze current file |
| `VGX: Detect AI Code in Workspace` | Scan all files |
| `VGX: Security Scan File` | Run security scan |

### Settings

```json
{
  "vgx.aiThreshold": 70,
  "vgx.highlightAICode": true,
  "vgx.showInlineConfidence": true
}
```

### Installation

```bash
cd vscode-extension
npm install
npm run compile
# Then install via "Install from VSIX" in VS Code
```

---

## How It Works

### AI Detection Algorithm

VGX uses a combination of **stylometry** (code style analysis) and **pattern matching** to detect AI-generated code.

```
┌─────────────────────────────────────────────────────────────┐
│                    VGX DETECTION ENGINE                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  INPUT: Source code file                                     │
│                                                              │
│  ┌─────────────────────┐    ┌─────────────────────┐         │
│  │   STYLOMETRY (45%)  │    │  PATTERNS (55%)     │         │
│  ├─────────────────────┤    ├─────────────────────┤         │
│  │ • Naming consistency│    │ • try/catch style   │         │
│  │ • Indentation       │    │ • async/await fetch │         │
│  │ • Comment density   │    │ • React hooks       │         │
│  │ • Line length var.  │    │ • Go error handling │         │
│  │ • Boilerplate ratio │    │ • Python docstrings │         │
│  │ • Empty line ratio  │    │ • JSDoc comments    │         │
│  └─────────────────────┘    └─────────────────────┘         │
│            │                          │                      │
│            └──────────┬───────────────┘                      │
│                       ▼                                      │
│              Combined Score (0-100%)                         │
│                       │                                      │
│                       ▼                                      │
│         ┌─────────────────────────┐                         │
│         │  Threshold Check (70%)  │                         │
│         └─────────────────────────┘                         │
│                       │                                      │
│            ┌──────────┴──────────┐                          │
│            ▼                     ▼                          │
│     🤖 AI-Generated        ✓ Human-Written                  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Stylometry Features

| Feature | AI Signal | Weight |
|---------|-----------|--------|
| Naming Consistency | High consistency = AI | 20% |
| Indentation | Perfect consistency = AI | 20% |
| Boilerplate Ratio | High boilerplate = AI | 20% |
| Line Length Variance | Low variance = AI | 15% |
| Comment Density | ~15% density = AI | 10% |
| Empty Line Ratio | Consistent spacing = AI | 10% |
| Average Line Length | ~45 chars = AI | 5% |

### Pattern Detection

VGX recognizes 16+ language-specific patterns commonly generated by AI:

**JavaScript/TypeScript:**
- `copilot_try_catch` — Standard try/catch with console.error
- `async_await_fetch` — Async function with await fetch
- `arrow_with_types` — Typed arrow functions
- `use_effect_deps` — React useEffect with dependency array
- `use_state_destructure` — React useState destructuring

**Go:**
- `go_error_check` — `if err != nil { return }` pattern
- `go_defer` — `defer file.Close()` pattern
- `go_struct_init` — Multi-line struct initialization

**Python:**
- `python_docstring` — Docstrings with Args/Returns/Raises
- `python_type_hints` — Function type annotations

---

## Pre-commit Hook

Block AI code or vulnerabilities before commit.

### Setup

```bash
# Create hook
cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
vgx scan --changes=true
if [ $? -ne 0 ]; then
  echo "VGX: Commit blocked due to security issues"
  exit 1
fi
EOF

chmod +x .git/hooks/pre-commit
```

### With pre-commit framework

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: vgx-scan
        name: VGX Security Scan
        entry: vgx scan --changes=true
        language: system
        pass_filenames: false
```

---

## Configuration

VGX works out of the box with zero configuration.

### Environment Variables

| Variable | Description |
|----------|-------------|
| `OPENAI_API_KEY` | Enable AI-enhanced scanning (optional) |
| `SEMGREP_RULES` | Custom Semgrep rules (default: `p/security-audit`) |

### Example `.env`

```bash
# Optional: Enhanced scanning with OpenAI
OPENAI_API_KEY=sk-...

# Optional: Custom Semgrep rules
SEMGREP_RULES=p/security-audit,p/secrets
```

---

## Project Structure

```
vgx/
├── cmd/
│   └── vgx/
│       └── main.go           # CLI entrypoint
├── pkg/
│   ├── detection/
│   │   ├── detector.go       # Main detector
│   │   ├── stylometry.go     # Style analysis
│   │   └── patterns.go       # Pattern matching
│   ├── scanner/
│   │   ├── scanner.go        # Vulnerability scanner
│   │   ├── semgrep.go        # Semgrep integration
│   │   └── openai.go         # OpenAI integration
│   ├── context/
│   │   └── manager.go        # File context & caching
│   ├── cache/
│   │   └── cache.go          # Scan result cache
│   └── types/
│       └── types.go          # Shared types
├── vscode-extension/
│   └── src/
│       └── extension.ts      # VS Code extension
├── website/
│   └── index.html            # Landing page
├── scripts/
│   └── install.sh            # Installer script
└── .github/
    └── workflows/
        └── release.yml       # CI/CD
```

---

## Requirements

- **Go 1.22+** (for building)
- **Semgrep** (optional, for security scanning)

```bash
# Install Semgrep (optional)
pip install semgrep
# or
brew install semgrep
```

---

## Contributing

Contributions welcome! Please read our contributing guidelines.

```bash
# Clone
git clone https://github.com/rohansx/vgx.git
cd vgx

# Build
go build -o vgx ./cmd/vgx

# Test
go test ./...

# Run
./vgx detect --path ./pkg
```

---

## License

MIT License — see [LICENSE](LICENSE)

---

## Links

- **Website:** [vgx.sh](https://vgx.sh)
- **GitHub:** [github.com/rohansx/vgx](https://github.com/rohansx/vgx)
- **Author:** [@rohansx](https://twitter.com/rohansx)

---

<p align="center">
  <sub>Built for developers who ship AI-assisted code responsibly.</sub>
</p>
