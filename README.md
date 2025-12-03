# Runbook Generator

[![CI](https://github.com/mrf/runbook-generator/actions/workflows/ci.yml/badge.svg)](https://github.com/mrf/runbook-generator/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/mrf/runbook-generator)](https://goreportcard.com/report/github.com/mrf/runbook-generator)
[![codecov](https://codecov.io/gh/mrf/runbook-generator/branch/main/graph/badge.svg)](https://codecov.io/gh/mrf/runbook-generator)

A CLI tool that transforms zsh history into actionable runbook documentation.

## Overview

Runbook Generator analyzes your zsh command history between specified command numbers and produces a structured markdown runbook. The tool removes duplicates, infers intent from command sequences, and scrubs sensitive data.

## Installation

```bash
go install github.com/mrf/runbook-generator/cmd/runbook-gen@latest
```

Or build from source:

```bash
git clone https://github.com/mrf/runbook-generator.git
cd runbook-generator
make build
```

## Usage

```bash
# Generate runbook from command numbers (use `history` to find numbers)
runbook-gen --from 100 --to 200 --title "Deployment Runbook" -o runbook.md

# Output to stdout
runbook-gen -f 1500 -t 1600
```

### Finding Command Numbers

Use the `history` command in zsh to see command numbers:

```bash
history | tail -50  # Last 50 commands with numbers
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--from` | `-f` | required | Start command number |
| `--to` | `-t` | required | End command number |
| `--output` | `-o` | stdout | Output file path |
| `--title` | | "Runbook" | Runbook title |

## Features

- **Smart deduplication**: Removes consecutive duplicates and typo corrections
- **Intent analysis**: Groups related commands and infers workflow purpose
- **Automatic prerequisites**: Detects required tools from commands
- **Sensitive data redaction**: Automatically detects and redacts 55+ secret patterns
- **AI-enhanced mode**: Optional Claude integration for smarter deduplication and explanations

## AI Features (Optional)

When `ANTHROPIC_API_KEY` is set, the tool enables AI-powered enhancements:

```bash
export ANTHROPIC_API_KEY=sk-ant-...
runbook-gen -f 100 -t 200 -o runbook.md
```

AI features include:
- Semantic deduplication (catches duplicates regex misses)
- Generated "why" explanations for each step
- Professional overview and prerequisites

Uses Claude 3.5 Haiku (~$0.01 per typical runbook).

## Security

### Important: Review Before Publishing

**This tool uses regex-based pattern matching to detect and redact sensitive data. While it covers 55+ common secret patterns, no automated detection is perfect.**

Before publishing any generated runbook:

1. **Carefully review the entire output** for any secrets that may have been missed
2. **Check for custom secret formats** specific to your organization
3. **Verify environment-specific values** like internal hostnames or IPs

The tool runs entirely locally and never transmits your data. **You are ultimately responsible** for ensuring no sensitive information is leaked.

### What Gets Redacted

- Passwords (command flags, environment variables, connection strings)
- API keys (AWS, GCP, Azure, Stripe, Twilio, SendGrid, etc.)
- Tokens (JWT, GitHub, Slack, Discord, NPM, PyPI, etc.)
- Private keys (RSA, EC, Ed25519, PGP, etc.)
- Webhook URLs (Slack, Discord)
- Database credentials
- Authorization headers

## Development

```bash
make test    # Run tests
make lint    # Run linter
make build   # Build binary
make release # Build for all platforms
```

### CI/CD

The project uses GitHub Actions for continuous integration. The CI pipeline runs on:
- Pushes to `main` branch
- Pull requests targeting `main`

**Pipeline Jobs:**

1. **Lint & Format** - Runs individual Go linters:
   - `gofmt` - Code formatting validation
   - `go vet` - Official Go static analyzer
   - `staticcheck` - Comprehensive static analysis
   - `gosec` - Security-focused linting
   - `errcheck` - Unchecked error detection
   - `ineffassign` - Ineffectual assignment detection
   - `misspell` - Spell checking

2. **Test** - Runs the test suite:
   - `go test` with race detection
   - Coverage reporting to Codecov

3. **Build** - Builds the binary:
   - Single platform build (Ubuntu)
   - Uploads artifact for verification

4. **Build Matrix** - Cross-platform builds:
   - Linux (ubuntu-latest)
   - macOS (macos-latest)
   - Windows (windows-latest)

All jobs use Go version specified in `go.mod` and leverage GitHub Actions caching for faster builds.

## License

MIT
