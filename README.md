# VGX — AI Code Security Scanner

<p align="center">
  <img src="https://img.shields.io/badge/version-2.0.0-blue" alt="Version">
  <img src="https://img.shields.io/github/stars/rohansx/vgx?style=social" alt="Stars">
  <img src="https://img.shields.io/badge/license-MIT-green" alt="License">
</p>

**VGX** is an open-source security scanner for AI-assisted development. It detects AI-generated code, scans for vulnerabilities, and integrates with your pre-commit workflow.

## Features

- 🤖 **AI Code Detection** — Identify AI-generated code (Copilot, Cursor, Claude)
- 🔒 **Security Scanning** — Vulnerability detection via Semgrep + optional OpenAI
- 🪝 **Pre-commit Hooks** — Block insecure code before it's committed
- 📊 **Reports** — HTML & JSON vulnerability reports
- 🐳 **Docker Support** — Run anywhere

## Quick Start

```bash
# Install
curl -sSL https://vgx.sh/install | bash

# Or with Go
go install github.com/rohansx/vgx@latest

# Detect AI-generated code
vgx detect --path ./src

# Security scan
vgx scan
```

## AI Code Detection

VGX uses stylometry and pattern analysis to detect AI-generated code — no API keys required.

```bash
$ vgx detect --path ./src

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
     🤖 src/components/Modal.tsx                    76%
     🤖 src/hooks/useAuth.ts                        71%
     ✓ src/index.ts                                 34%
     ✓ src/config.ts                                28%
     ...

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  🤖 4 file(s) detected as AI-generated
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

## Security Scanning

```bash
# Scan changed files (default)
vgx scan

# Scan all files
vgx scan --changes=false

# Scan specific file
vgx scan src/auth.ts
```

### Pre-commit Hook

```bash
# Add to .git/hooks/pre-commit
#!/bin/bash
vgx scan --changes=true
```

Or use the install script:

```bash
vgx install-hook
```

## Detection Methods

| Method | Accuracy | Description |
|--------|----------|-------------|
| Stylometry | 75-85% | Naming patterns, indentation, comment density |
| Pattern Matching | 80-90% | Known AI code signatures |
| Telemetry | 99% | IDE extension (coming soon) |

## Configuration

VGX works out of the box. For custom settings:

```bash
# Optional: Enhanced scanning with OpenAI
export OPENAI_API_KEY=sk-...

# Semgrep rules (auto-detected)
export SEMGREP_RULES=p/security-audit
```

## Commands

```
vgx <command> [options]

Commands:
  scan      Security scan (vulnerabilities, secrets)
  detect    Detect AI-generated code
  version   Print version
  help      Show help

Detect Options:
  --path, -p     Path to scan (default: .)
  --format, -f   Output: text, json (default: text)
  --threshold    AI detection threshold 0-100 (default: 70)

Scan Options:
  --changes      Scan only changed files (default: true)
  --report       Generate HTML/JSON report (default: true)
```

## VS Code Extension

Coming soon — real-time AI code highlighting in your editor.

## Why VGX?

- **Privacy-first**: Code never leaves your machine (unless you enable OpenAI)
- **Fast**: Rule-based analysis, no ML inference required
- **Open source**: Audit, modify, self-host

## Contributing

PRs welcome! See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT — see [LICENSE](LICENSE)

---

<p align="center">
  <a href="https://vgx.sh">Website</a> •
  <a href="https://github.com/rohansx/vgx/issues">Issues</a> •
  <a href="https://twitter.com/rohansx">Twitter</a>
</p>
