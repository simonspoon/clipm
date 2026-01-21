# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

clipm is a CLI-based task manager designed for use by LLMs and agents. It uses a single JSON file (`.clipm/tasks.json`) for storage and outputs JSON by default for easy parsing.

## Build & Development Commands

```bash
# Build
go build -o clipm ./cmd/clipm

# Run tests
go test ./...

# Run single test
go test ./internal/commands -run TestAddCommand

# Run tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Lint
golangci-lint run
```

## Architecture

### Package Structure

- `cmd/clipm/main.go` - Entry point, calls `commands.Execute()`
- `internal/commands/` - Cobra command implementations (one file per command)
- `internal/models/` - Task model and status constants
- `internal/storage/` - JSON file storage operations

### Storage Design

Tasks are stored in `.clipm/tasks.json` with this structure:
```json
{
  "version": "2.0.0",
  "tasks": [...]
}
```

The `Storage` type walks up directories to find `.clipm/` (like git finds `.git/`). Use `NewStorage()` for auto-discovery or `NewStorageAt(dir)` for tests.

### Task Model

Tasks have: ID (Unix milliseconds), Name, Description, Parent (nullable pointer), Status, Created, Updated.

Valid statuses: `todo`, `in-progress`, `done`

### Output Convention

All commands default to JSON output. Use `--pretty` flag for human-readable output with colors.

### Key Behaviors

- `next` uses depth-first traversal for progressive decomposition workflows:
  - Finds deepest in-progress task, returns its todo children (then siblings)
  - Walks up hierarchy when no todos at current level
  - Returns `{"task": ...}` when context exists, `{"candidates": [...]}` when no in-progress tasks
- Tasks cannot be marked `done` if they have undone children
- Cannot add children to `done` tasks
- `delete` orphans children (sets their Parent to nil)
- `prune` removes all `done` tasks
