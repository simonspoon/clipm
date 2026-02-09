# clipm

A CLI-based task manager designed for use by LLMs and AI agents.

clipm uses a single JSON file (`.clipm/tasks.json`) for storage and outputs JSON by default for easy parsing. It supports hierarchical task structures with progressive decomposition workflows.

## Installation

```bash
go install github.com/simonspoon/clipm/cmd/clipm@latest
```

Or build from source:

```bash
git clone https://github.com/simonspoon/clipm.git
cd clipm
go build -o clipm ./cmd/clipm
```

## Quick Start

```bash
# Initialize clipm in your project
clipm init

# Add tasks
clipm add "Implement user authentication"
clipm add "Add login endpoint" --parent <task-id> -d "REST endpoint for user login"

# View tasks
clipm list                    # JSON output
clipm list --pretty           # Human-readable output
clipm tree                    # Hierarchical view (pretty by default)

# Update task status
clipm status <task-id> in-progress
clipm status <task-id> done

# Get next task (depth-first traversal)
clipm next

# Watch for changes
clipm watch --pretty
```

## Command Reference

| Command | Description |
|---------|-------------|
| `init` | Initialize clipm in the current directory |
| `add <name>` | Add a new task (`--parent`, `--description`/`-d`) |
| `list` | List all tasks |
| `tree` | Display tasks in a tree structure |
| `show <id>` | Show details for a specific task |
| `status <id> <status>` | Update task status (`todo`, `in-progress`, `done`) |
| `next` | Get the next task to work on |
| `parent <id> <parent-id>` | Set a task's parent |
| `unparent <id>` | Remove a task's parent |
| `delete <id>` | Delete a task |
| `prune` | Remove all completed tasks |
| `watch` | Watch tasks for live updates |
| `block <blocker> <blocked>` | Add dependency (blocked waits for blocker) |
| `unblock <blocker> <blocked>` | Remove dependency |
| `note <id> "message"` | Add a timestamped note to a task |
| `claim <id> <agent>` | Claim task ownership |
| `unclaim <id>` | Release task ownership |

All commands output JSON by default. Use `--pretty` for human-readable output with colors.

### Filtering

The `list` command supports filtering:
- `--status <status>` - Filter by status
- `--owner <name>` - Filter by owner
- `--unclaimed` - Show only unowned tasks
- `--blocked` / `--unblocked` - Filter by blocked state

The `next` command supports:
- `--unclaimed` - Skip tasks that have an owner

## Usage with AI Agents

clipm is designed for integration with LLMs and AI agents like Claude Code. The JSON output makes it easy to parse task information programmatically.

Example workflow:
```bash
# Agent checks for next task
clipm next

# Returns JSON like:
# {"task": {"id": "abcd", "name": "Implement feature X", ...}}

# Agent claims and starts task
clipm claim abcd agent-1
clipm status abcd in-progress

# Agent adds progress notes
clipm note abcd "Started implementation"
clipm note abcd "Found edge case, handling it"

# Agent completes work, marks done
clipm status abcd done
```

### Multi-Agent Coordination

clipm supports multiple agents working on the same task queue:

```bash
# Agent claims an unclaimed task
clipm next --unclaimed
clipm claim <id> agent-1

# Other agents skip claimed tasks
clipm next --unclaimed  # won't return agent-1's task

# Set up task dependencies
clipm block <prereq-id> <dependent-id>
# dependent task won't appear in `next` until prereq is done

# When prereq completes, dependent is auto-unblocked
clipm status <prereq-id> done
```

### Progressive Decomposition

The `next` command uses depth-first traversal to support progressive decomposition:
- Finds the deepest in-progress task
- Returns its todo children (or siblings if none)
- Walks up the hierarchy when no todos exist at the current level

This allows agents to break down large tasks into smaller subtasks just-in-time.

### Watch Mode

The `watch` command monitors tasks.json for changes and outputs updates in real-time:

```bash
# JSON mode (default) - outputs newline-delimited events
clipm watch

# Pretty mode - clears screen and redraws task list
clipm watch --pretty

# Filter by status
clipm watch --status in-progress --pretty

# Custom polling interval
clipm watch --interval 1s
```

JSON mode outputs events:
- `snapshot` - Initial task list on startup
- `added` - New task created
- `updated` - Task modified
- `deleted` - Task removed

## Storage

Tasks are stored in `.clipm/tasks.json` in your project directory. The storage system walks up directories to find the `.clipm/` folder (similar to how git finds `.git/`).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.
