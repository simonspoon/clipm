# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

clipm is a CLI-based task manager designed for use by LLMs and agents. It uses a single JSON file (`.clipm/tasks.json`) for storage and outputs JSON by default for easy parsing.

For detailed architecture, contributor guides, and references, see `docs/INDEX.md` and load the relevant topic file before working on unfamiliar subsystems.

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

### Key Rules

**Storage:**
- Tasks stored in `.clipm/tasks.json`, version `3.0.0`
- Storage walks up directories to find `.clipm/` (like git finds `.git/`)

**Task model:**
- Tasks have ID (4-char alpha), Name, Description, Parent, Status, BlockedBy, Owner, Notes, Created, Updated

**Output:**
- All commands output JSON by default. Use `--pretty` for human-readable output.

**Behaviors:**
- `next` uses depth-first traversal; skips blocked tasks; `--unclaimed` skips owned
- Done tasks hidden by default in `list`/`tree`/`watch`; use `--show-all`
- Cannot mark done if undone children exist
- Cannot set in-progress if blocked
- Cannot add children to done tasks
- `delete` orphans children; `prune` removes all done tasks

**Dependencies:**
- `block <blocker> <blocked>` -- cycle detection prevents circular deps
- Done tasks auto-removed from all BlockedBy lists

**Ownership:**
- `claim`/`unclaim` for agent ownership; `--force` overrides; `next --unclaimed` skips owned

**Notes:**
- `note <id> "msg"` appends timestamped, append-only note
