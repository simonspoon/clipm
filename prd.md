# Product Requirements Document: clipm

## Overview

**Product Name**: clipm (CLI Project Manager)  
**Version**: 1.0.0  
**Date**: November 8, 2025  
**Language**: Go

### Purpose
A lightweight, CLI-based task manager that uses human-readable markdown files for storage. Designed for developers who want to track tasks alongside their code without heavy tooling or cloud dependencies.

### Key Principles
- **Simple**: Plain text storage, no database
- **Portable**: Single binary, no runtime dependencies
- **Git-friendly**: Markdown files can be committed and merged
- **Local-first**: No network calls, works offline
- **Developer-focused**: Terminal-native workflow

---

## User Stories

### Primary User: Software Developer
- "I want to track tasks in my project directory so they live alongside my code"
- "I want to view and edit tasks without leaving the terminal"
- "I want task files I can commit to version control and share with teammates"
- "I want to organize tasks hierarchically with parent/child relationships"
- "I want to archive completed tasks so my active list stays clean"

---

## Technical Architecture

### Storage System

#### Directory Structure
```
/project-root/
  .clipm/
    index.json           # Fast lookup index
    task-1699459200000.md
    task-1699459300000.md
    task-1699459400000.md
    archive/
      task-1699450000000.md
      task-1699451000000.md
```

#### Task File Format
Each task is a markdown file with YAML frontmatter:

```markdown
---
id: 1699459200000
name: Implement user authentication
description: Add JWT-based auth system
parent: null
status: in-progress
priority: high
created: 2025-11-08T14:30:00Z
updated: 2025-11-08T15:45:00Z
tags:
  - backend
  - security
---

## Implementation Notes

Use jsonwebtoken library for token generation and validation.

## Requirements
- Store tokens in httpOnly cookies
- Add refresh token rotation
- Implement rate limiting

## Subtasks
- [ ] Set up JWT middleware
- [ ] Create auth endpoints
- [ ] Add token refresh logic
- [ ] Write integration tests
```

#### Index File Format
```json
{
  "version": "1.0.0",
  "tasks": {
    "1699459200000": {
      "id": 1699459200000,
      "name": "Implement user authentication",
      "status": "in-progress",
      "priority": "high",
      "parent": null,
      "created": "2025-11-08T14:30:00Z",
      "updated": "2025-11-08T15:45:00Z",
      "archived": false
    },
    "1699459300000": {
      "id": 1699459300000,
      "name": "Set up JWT middleware",
      "status": "todo",
      "priority": "medium",
      "parent": 1699459200000,
      "created": "2025-11-08T14:35:00Z",
      "updated": "2025-11-08T14:35:00Z",
      "archived": false
    }
  }
}
```

**Index Purpose**: Fast queries without parsing all markdown files. Rebuilt from files if corrupted/deleted.

---

## Feature Requirements

### 1. Initialization
**Command**: `clipm init`

**Behavior**:
- Creates `.clipm/` directory in current working directory
- Creates empty `index.json`
- Creates `archive/` subdirectory
- Fails gracefully if `.clipm/` already exists

**Output**:
```
✓ Initialized clipm in /path/to/project
```

---

### 2. Task Creation
**Command**: `clipm add <name> [flags]`

**Flags**:
- `-d, --description <text>` - Task description
- `-p, --priority <low|medium|high>` - Priority level (default: medium)
- `-t, --tags <tag1,tag2>` - Comma-separated tags
- `--parent <id>` - Parent task ID

**Behavior**:
- Generates timestamp-based ID
- Creates markdown file with frontmatter
- Updates index.json
- Sets status to "todo" by default
- Opens task in $EDITOR if body content needed

**Output**:
```
✓ Created task 1699459200000: Implement user authentication
```

---

### 3. Task Listing
**Command**: `clipm list [flags]`

**Flags**:
- `-s, --status <status>` - Filter by status
- `-p, --priority <priority>` - Filter by priority
- `-t, --tag <tag>` - Filter by tag
- `--parent <id>` - Show children of parent
- `--no-parent` - Show only top-level tasks
- `-a, --all` - Include archived tasks

**Output Format**:
```
TODO (3)
  1699459200000  [HIGH]  Implement user authentication  #backend #security
  1699459300000  [MED]   Set up database migrations     #database
  1699459400000  [LOW]   Update documentation           #docs

IN PROGRESS (1)
  1699459500000  [HIGH]  Fix login bug                  #bugfix

BLOCKED (1)
  1699459600000  [MED]   Deploy to staging              #devops
    → Blocked by: 1699459500000
```

**Tree View** (if parent/child tasks exist):
```
1699459200000  Implement user authentication
  ├─ 1699459300000  Set up JWT middleware
  ├─ 1699459310000  Create auth endpoints
  └─ 1699459320000  Add token refresh logic
```

---

### 4. Task Details
**Command**: `clipm show <id>`

**Behavior**:
- Reads task markdown file
- Displays formatted output with metadata and body

**Output**:
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Task: 1699459200000
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Name:        Implement user authentication
Description: Add JWT-based auth system
Status:      in-progress
Priority:    high
Parent:      none
Tags:        backend, security
Created:     2025-11-08 14:30:00
Updated:     2025-11-08 15:45:00

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

## Implementation Notes

Use jsonwebtoken library for token generation and validation.

## Requirements
- Store tokens in httpOnly cookies
- Add refresh token rotation
...
```

---

### 5. Task Editing
**Command**: `clipm edit <id>`

**Behavior**:
- Opens task markdown file in $EDITOR (falls back to vi/vim/nano)
- Validates YAML frontmatter after save
- Updates index.json with new metadata
- Updates `updated` timestamp

**Error Handling**:
- Invalid YAML: Show error, don't update index
- $EDITOR not set: Use sensible defaults

---

### 6. Status Updates
**Command**: `clipm status <id> <status>`

**Valid Statuses**: `todo`, `in-progress`, `done`, `blocked`

**Behavior**:
- Updates status in frontmatter
- Updates index.json
- Updates `updated` timestamp
- Does NOT auto-archive (use `clipm done` for that)

**Output**:
```
✓ Updated task 1699459200000 status: in-progress
```

---

### 7. Task Completion
**Command**: `clipm done <id>`

**Behavior**:
- Sets status to "done"
- Moves task file to `archive/` directory
- Updates index.json (marks as archived)
- Updates `updated` timestamp

**Output**:
```
✓ Completed and archived task 1699459200000
```

---

### 8. Task Deletion
**Command**: `clipm delete <id>`

**Behavior**:
- Prompts for confirmation
- Deletes markdown file (or moves to archive if in active dir)
- Removes from index.json
- If task has children, prompts to delete children or orphan them

**Confirmation**:
```
Delete task 1699459200000: "Implement user authentication"?
This task has 3 child tasks. Delete children too? [y/N/orphan]
```

---

### 9. Parent Management
**Command**: `clipm parent <id> <parent-id>`

**Behavior**:
- Sets parent field in frontmatter
- Updates index.json
- Validates parent exists and isn't archived
- Prevents circular dependencies

**Command**: `clipm unparent <id>`

**Behavior**:
- Sets parent to null
- Updates index.json

**Output**:
```
✓ Task 1699459300000 is now a child of 1699459200000
✓ Task 1699459300000 is now a top-level task
```

---

## Data Models

### Task Metadata (YAML Frontmatter)
```go
type Task struct {
    ID          int64     `yaml:"id"`
    Name        string    `yaml:"name"`
    Description string    `yaml:"description"`
    Parent      *int64    `yaml:"parent"` // null if top-level
    Status      string    `yaml:"status"` // todo, in-progress, done, blocked
    Priority    string    `yaml:"priority"` // low, medium, high
    Created     time.Time `yaml:"created"`
    Updated     time.Time `yaml:"updated"`
    Tags        []string  `yaml:"tags"`
}
```

### Index Entry
```go
type IndexEntry struct {
    ID       int64     `json:"id"`
    Name     string    `json:"name"`
    Status   string    `json:"status"`
    Priority string    `json:"priority"`
    Parent   *int64    `json:"parent"`
    Created  time.Time `json:"created"`
    Updated  time.Time `json:"updated"`
    Archived bool      `json:"archived"`
}

type Index struct {
    Version string                `json:"version"`
    Tasks   map[int64]IndexEntry `json:"tasks"`
}
```

---

## Error Handling

### Common Errors
1. **No .clipm directory**: `Error: Not in a clipm project. Run 'clipm init' first.`
2. **Task not found**: `Error: Task 1699459200000 not found.`
3. **Invalid status**: `Error: Invalid status "foo". Must be: todo, in-progress, done, blocked`
4. **Circular parent**: `Error: Cannot set parent - would create circular dependency.`
5. **Corrupted file**: `Error: Failed to parse task file. YAML frontmatter may be invalid.`

### Index Recovery
If `index.json` is missing or corrupted:
- Rebuild from all markdown files in `.clipm/`
- Show warning: `Warning: Rebuilt index from task files.`

---

## Non-Functional Requirements

### Performance
- List command: < 100ms for 1000 tasks
- Add/edit commands: < 50ms
- Index rebuild: < 500ms for 1000 tasks

### Compatibility
- Go 1.21+
- Unix-like systems (Linux, macOS, BSD)
- Windows support (best effort)

### Binary Size
- Target: < 10MB uncompressed
- Single binary, statically linked

---

## Future Enhancements (Out of Scope for v1.0)

### v1.1 Candidates
- `clipm tree` - ASCII tree view of task hierarchy
- `clipm tag <id> <tag>` - Add/remove tags
- `clipm search <query>` - Full-text search in task bodies
- `clipm export` - Export to JSON/CSV

### v2.0 Candidates
- Due dates and reminders
- Time tracking
- Multiple projects (project switching)
- Custom fields
- Git integration (auto-commit on changes)
- TUI (text user interface) mode

---

## Dependencies

### Go Libraries
- **cobra** - CLI framework
- **yaml.v3** - YAML parsing
- **color** - Terminal colors
- Standard library for file operations

### Development Tools
- Go 1.21+
- golangci-lint
- goreleaser (for releases)

---

## Success Metrics

### v1.0 Launch Criteria
- All core commands implemented and tested
- Binary builds for Linux, macOS, Windows
- README with usage examples
- 90%+ test coverage
- No critical bugs

### User Success
- Can manage 500+ tasks without performance issues
- Index corruption auto-recovery works
- Task files are readable and editable outside CLI

---

## Open Questions

1. **Should we support config file** (~/.clipm.yaml) for default flags?
2. **Color scheme**: Configurable or sensible defaults?
3. **Interactive mode**: Add prompts for task creation vs pure CLI args?
4. **Bulk operations**: `clipm done --tag=backend` to complete multiple tasks?
5. **Task templates**: Pre-defined task structures?

---

## Appendix: Example Workflows

### Daily Development Workflow
```bash
# Start of day
cd ~/projects/myapp
clipm list --status=in-progress

# Add new task
clipm add "Fix login bug" -p high -t bugfix

# Work on task, add notes
clipm edit 1699459200000

# Mark complete
clipm done 1699459200000

# Review all tasks
clipm list
```

### Managing Subtasks
```bash
# Create parent task
clipm add "Implement authentication" -p high

# Create subtasks
clipm add "Set up JWT middleware" --parent 1699459200000
clipm add "Create auth endpoints" --parent 1699459200000
clipm add "Add tests" --parent 1699459200000

# View hierarchy
clipm list --parent 1699459200000
```

---

**Document Status**: Draft  
**Next Review**: After initial prototype