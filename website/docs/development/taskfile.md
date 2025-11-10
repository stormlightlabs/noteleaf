---
title: Task Automation
sidebar_label: Taskfile
sidebar_position: 3
description: Using Taskfile for development workflows.
---

# Task Automation

Noteleaf uses [Task](https://taskfile.dev) to automate common development workflows.

## Installation

### macOS

```sh
brew install go-task/tap/go-task
```

### Linux

```sh
sh -c "$(curl -fsSL https://taskfile.dev/install.sh)"
```

### Go Install

```sh
go install github.com/go-task/task/v3/cmd/task@latest
```

## Available Tasks

View all tasks:

```sh
task
# or
task --list
```

## Common Tasks

### Build Commands

**task build** - Quick development build

```sh
task build
```

Output: `./tmp/noteleaf`

**task build:dev** - Build with version information

```sh
task build:dev
```

Includes git commit hash and build date.

**task build:rc** - Release candidate build

```sh
git tag v1.0.0-rc1
task build:rc
```

Requires git tag with `-rc` suffix.

**task build:prod** - Production build

```sh
git tag v1.0.0
task build:prod
```

Requires clean semver tag and no uncommitted changes.

### Testing Commands

**task test** - Run all tests

```sh
task test
```

**task coverage** - Generate HTML coverage report

```sh
task coverage
open coverage.html  # View report
```

**task cov** - Terminal coverage summary

```sh
task cov
```

**task check** - Lint and coverage

```sh
task check
```

Runs linters and generates coverage report.

### Development Commands

**task dev** - Full development workflow

```sh
task dev
```

Runs:

1. `task clean`
2. `task lint`
3. `task test`
4. `task build`

**task lint** - Run linters

```sh
task lint
```

Runs `go vet` and `go fmt`.

**task run** - Build and run

```sh
task run
```

Builds then executes the binary.

### Maintenance Commands

**task clean** - Remove build artifacts

```sh
task clean
```

Removes:

- `./tmp/` directory
- Coverage files

**task deps** - Download and tidy dependencies

```sh
task deps
```

Runs `go mod download` and `go mod tidy`.

### Documentation Commands

**task docs:generate** - Generate all documentation

```sh
task docs:generate
```

Generates:

- Docusaurus docs (website/docs/manual)
- Man pages (docs/manual)

**task docs:man** - Generate man pages

```sh
task docs:man
```

**task docs:serve** - Start documentation server

```sh
task docs:serve
```

Starts Docusaurus development server at <http://localhost:3000>.

### Version Commands

**task version:show** - Display version info

```sh
task version:show
```

Shows:

- Git tag
- Git commit
- Git describe output
- Build date

**task version:validate** - Validate git tag

```sh
task version:validate
```

Checks tag format for releases.

### Utility Commands

**task status** - Show Go environment

```sh
task status
```

Displays:

- Go version
- Module information
- Dependencies

## Taskfile Variables

Variables injected during build:

```yaml
BINARY_NAME: noteleaf
BUILD_DIR: ./tmp
VERSION_PKG: github.com/stormlightlabs/noteleaf/internal/version
GIT_COMMIT: $(git rev-parse --short HEAD)
GIT_TAG: $(git describe --tags --exact-match)
BUILD_DATE: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
```

## Task Dependencies

Some tasks automatically trigger others:

```sh
task run
# Automatically runs: task build
```

```sh
task dev
# Runs in sequence:
# 1. task clean
# 2. task lint
# 3. task test
# 4. task build
```

## Custom Workflows

### Pre-commit Workflow

```sh
task lint && task test
```

### Release Preparation

```sh
task check && \
git tag v1.0.0 && \
task build:prod && \
./tmp/noteleaf version
```

### Documentation Preview

```sh
task docs:generate
task docs:serve
```

### Full CI Simulation

```sh
task clean && \
task deps && \
task lint && \
task test && \
task coverage && \
task build:dev
```

## Taskfile Structure

Location: `Taskfile.yml` (project root)

Key sections:

- **vars**: Build variables and git information
- **tasks**: Command definitions with descriptions
- **deps**: Task dependencies
- **preconditions**: Validation before execution

## Configuration

Customize via `Taskfile.yml` or environment variables:

```yaml
vars:
  BINARY_NAME: noteleaf
  BUILD_DIR: ./tmp
```

Override at runtime:

```sh
BINARY_NAME=custom-noteleaf task build
```

## Why Task Over Make?

- Cross-platform (Windows, macOS, Linux)
- YAML syntax (more readable than Makefile)
- Built-in variable interpolation
- Better dependency management
- Precondition validation
- Native Go integration

## Further Reading

- [Task Documentation](https://taskfile.dev)
- [Taskfile Schema](https://taskfile.dev/api/)
- Project Taskfile: `Taskfile.yml`
