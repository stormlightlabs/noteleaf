---
title: Building Noteleaf
sidebar_label: Building
sidebar_position: 1
description: Build configurations and development workflows.
---

# Building Noteleaf

Noteleaf uses [Task](https://taskfile.dev) for build automation, providing consistent workflows across development, testing, and releases.

## Prerequisites

- Go 1.21 or later
- [Task](https://taskfile.dev) (install via `brew install go-task/tap/go-task` on macOS)
- Git (for version information)

## Build Types

### Development Build

Quick build without version injection for local development:

```sh
task build
```

Output: `./tmp/noteleaf`

### Development Build with Version

Build with git commit hash and development tools enabled:

```sh
task build:dev
```

Version format: `git describe` output (e.g., `v0.1.0-15-g1234abc`)
Output: `./tmp/noteleaf`

### Release Candidate Build

Build with `-rc` tag, excludes development tools:

```sh
git tag v1.0.0-rc1
task build:rc
```

Requirements:

- Clean git tag with `-rc` suffix
- Tag format: `v1.0.0-rc1`, `v2.1.0-rc2`, etc.

### Production Build

Build for release with strict validation:

```sh
git tag v1.0.0
task build:prod
```

Requirements:

- Clean semver git tag (e.g., `v1.0.0`, `v2.1.3`)
- No uncommitted changes
- No prerelease suffix

## Build Tags

Production builds use the `prod` build tag to exclude development and seed commands:

```go
//go:build !prod
```

Commands excluded from production:

- `noteleaf dev` - Development utilities
- `noteleaf seed` - Test data generation

## Version Information

Build process injects version metadata via ldflags:

```go
// internal/version/version.go
var (
    Version   = "dev"           // Git tag or "dev"
    Commit    = "none"          // Git commit hash
    BuildDate = "unknown"       // Build timestamp
)
```

View version information:

```sh
task version:show
noteleaf version
```

## Build Artifacts

All binaries are built to `./tmp/` directory:

```
tmp/
└── noteleaf    # Binary for current platform
```

## Development Workflow

Full development cycle with linting and testing:

```sh
task dev
```

Runs:

1. Clean build artifacts
2. Run linters (vet, fmt)
3. Execute all tests
4. Build binary

## Manual Build

Build directly with Go (bypasses Task automation):

```sh
go build -o ./tmp/noteleaf ./cmd
```

With version injection:

```sh
go build -ldflags "-X github.com/stormlightlabs/noteleaf/internal/version.Version=v1.0.0" -o ./tmp/noteleaf ./cmd
```

## Cross-Platform Builds

Build for specific platforms:

```sh
# Linux
GOOS=linux GOARCH=amd64 go build -o ./tmp/noteleaf-linux ./cmd

# Windows
GOOS=windows GOARCH=amd64 go build -o ./tmp/noteleaf.exe ./cmd

# macOS (ARM)
GOOS=darwin GOARCH=arm64 go build -o ./tmp/noteleaf-darwin-arm64 ./cmd
```

## Clean Build

Remove all build artifacts:

```sh
task clean
```

Removes:

- `./tmp/` directory
- `coverage.out` and `coverage.html`
