# GitHub Actions Workflows

## Overview

This directory contains GitHub Actions workflows for the runbook-generator project.

## Workflows

### ci.yml - Continuous Integration

**Triggers:**
- Push to `main` branch
- Pull requests targeting `main`

**Concurrency Control:**
- Cancels in-progress runs when new commits are pushed to the same branch/PR
- Saves CI minutes and provides faster feedback

**Jobs:**

#### 1. lint-format
Runs individual Go linters (not golangci-lint) for granular control and visibility:

| Linter | Purpose | Exit on Error |
|--------|---------|---------------|
| gofmt | Code formatting | Yes |
| go vet | Official static analysis | Yes |
| staticcheck | Comprehensive static analysis | Yes |
| gosec | Security vulnerabilities (G104 excluded) | Yes |
| errcheck | Unchecked errors | Yes |
| ineffassign | Ineffectual assignments | Yes |
| misspell | Spelling errors | Yes |

#### 2. test
Runs the test suite with:
- Race detection enabled (`-race`)
- Coverage reporting (`-coverprofile`)
- Uploads to Codecov (non-blocking)

#### 3. build
Builds the binary for Linux and uploads as artifact:
- Depends on: `lint-format`, `test`
- Artifact retention: 7 days
- Named: `runbook-gen-{sha}`

#### 4. build-matrix
Cross-platform build verification:
- Depends on: `lint-format`, `test`
- Platforms: Ubuntu, macOS, Windows
- Non-blocking (`fail-fast: false`)

## Why Individual Linters?

Instead of using `golangci-lint` as an aggregated linter, we run each linter separately for:

1. **Better Visibility** - Know exactly which linter failed
2. **Granular Control** - Configure each linter independently
3. **Clearer Logs** - Easier to debug specific issues
4. **Flexible Execution** - Run different linters at different stages if needed
5. **Maintainability** - Update individual linters without affecting others

## Caching Strategy

- **Go Modules**: Cached via `actions/setup-go@v5` with `cache: true`
- **Cache Key**: Based on `go.mod` and `go.sum` files
- **Automatic**: No manual cache configuration needed

## Making Changes

When modifying workflows:

1. Edit the workflow file
2. Run `yamllint .github/workflows/ci.yml`
3. Run `actionlint .github/workflows/ci.yml`
4. Fix any errors reported by either tool
5. Test changes in a feature branch PR
6. Verify all jobs pass before merging

## Future Enhancements

Potential additions (not implemented yet):

- **Release workflow** - Automated releases with GoReleaser
- **Dependency updates** - Dependabot or Renovate
- **Security scanning** - Additional vulnerability checks
- **Performance tests** - Benchmark tracking over time
- **Documentation** - Auto-generate docs on changes
