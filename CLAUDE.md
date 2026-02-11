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

# Lint (REQUIRED before committing)
golangci-lint run
```

**Important:** Always run `golangci-lint run` before committing. CI will fail if there are linter errors.

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
  "version": "3.0.0",
  "tasks": [...]
}
```

The `Storage` type walks up directories to find `.clipm/` (like git finds `.git/`). Use `NewStorage()` for auto-discovery or `NewStorageAt(dir)` for tests.

### Task Model

Tasks have:
- `ID` - 4-character lowercase alphabetic string (e.g., `abcd`, `wxyz`)
- `Name` - Task title
- `Description` - Optional details
- `Parent` - Nullable pointer to parent task ID
- `Status` - `todo`, `in-progress`, or `done`
- `BlockedBy` - List of task IDs that must complete first
- `Owner` - Optional agent name claiming the task
- `Notes` - Append-only list of timestamped observations
- `Created`, `Updated` - Timestamps

### Output Convention

All commands default to JSON output. Use `--pretty` flag for human-readable output with colors.

### Key Behaviors

- `next` uses depth-first traversal for progressive decomposition workflows:
  - Finds deepest in-progress task, returns its todo children (then siblings)
  - Walks up hierarchy when no todos at current level
  - Returns `{"task": ...}` when context exists, `{"candidates": [...]}` when no in-progress tasks
  - Always skips blocked tasks; use `--unclaimed` to also skip owned tasks
- Completed tasks are hidden by default in `list`, `tree`, and `watch`. A done task is only shown if its parent exists and is not done (i.e., it's a completed subtask of active work). Use `--show-all` to see all tasks including completed.
- `watch --pretty` renders as a hierarchical tree (same as `tree`)
- Tasks cannot be marked `done` if they have undone children
- Tasks cannot be set to `in-progress` if they are blocked
- Cannot add children to `done` tasks
- `delete` orphans children (sets their Parent to nil)
- `prune` removes all `done` tasks

### Dependencies

- `block <blocker> <blocked>` adds blocker to blocked's BlockedBy list
- Cycle detection prevents A→B→A dependency chains
- Cannot block on completed tasks
- When a task is marked `done`, it's auto-removed from all BlockedBy lists
- `next` skips tasks with incomplete blockers

### Ownership

- `claim <id> <agent>` sets Owner; fails if already owned (use `--force` to override)
- `unclaim <id>` clears Owner
- `list --owner <name>` filters by owner; `--unclaimed` shows unowned tasks
- `next --unclaimed` skips owned tasks

### Notes

- `note <id> "message"` appends a timestamped note
- Notes are append-only and displayed in `show --pretty`
