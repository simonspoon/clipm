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
clipm add "Add login endpoint" --parent <task-id>

# View tasks
clipm list                    # JSON output
clipm list --pretty           # Human-readable output
clipm tree --pretty           # Hierarchical view

# Update task status
clipm status <task-id> in-progress
clipm status <task-id> done

# Get next task (depth-first traversal)
clipm next
```

## Command Reference

| Command | Description |
|---------|-------------|
| `init` | Initialize clipm in the current directory |
| `add <name>` | Add a new task (use `--parent` for subtasks) |
| `list` | List all tasks |
| `tree` | Display tasks in a tree structure |
| `show <id>` | Show details for a specific task |
| `status <id> <status>` | Update task status (`todo`, `in-progress`, `done`) |
| `next` | Get the next task to work on |
| `parent <id> <parent-id>` | Set a task's parent |
| `unparent <id>` | Remove a task's parent |
| `delete <id>` | Delete a task |
| `prune` | Remove all completed tasks |

All commands output JSON by default. Use `--pretty` for human-readable output with colors.

## Usage with AI Agents

clipm is designed for integration with LLMs and AI agents like Claude Code. The JSON output makes it easy to parse task information programmatically.

Example workflow:
```bash
# Agent checks for next task
clipm next

# Returns JSON like:
# {"task": {"id": 1737500000000, "name": "Implement feature X", ...}}

# Agent marks task in progress
clipm status 1737500000000 in-progress

# Agent completes work, marks done
clipm status 1737500000000 done
```

### Progressive Decomposition

The `next` command uses depth-first traversal to support progressive decomposition:
- Finds the deepest in-progress task
- Returns its todo children (or siblings if none)
- Walks up the hierarchy when no todos exist at the current level

This allows agents to break down large tasks into smaller subtasks just-in-time.

## Storage

Tasks are stored in `.clipm/tasks.json` in your project directory. The storage system walks up directories to find the `.clipm/` folder (similar to how git finds `.git/`).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.
