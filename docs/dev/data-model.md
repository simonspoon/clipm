# Data Model

This document is a precise reference for the data structures used by clipm. All types are defined in Go source; JSON representations are what appears in `.clipm/tasks.json` and in command output.

## Source Files

| Type | Source file |
|------|-------------|
| `Task`, `Note`, status constants | `internal/models/task.go` |
| `TaskStore`, `NextResult` | `internal/storage/storage.go` |
| `WatchEvent` | `internal/commands/watch.go` |

---

## Task

Defined in `internal/models/task.go`.

```go
type Task struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description,omitempty"`
    Parent      *string   `json:"parent"`
    Status      string    `json:"status"`
    BlockedBy   []string  `json:"blockedBy,omitempty"`
    Owner       *string   `json:"owner,omitempty"`
    Notes       []Note    `json:"notes,omitempty"`
    Created     time.Time `json:"created"`
    Updated     time.Time `json:"updated"`
}
```

| Field | Go type | JSON tag | Description |
|-------|---------|----------|-------------|
| `ID` | `string` | `"id"` | 4-character lowercase alphabetic string (e.g. `"abcd"`). Generated via `crypto/rand`. User input is normalized to lowercase via `NormalizeTaskID`. |
| `Name` | `string` | `"name"` | Task title. Required. |
| `Description` | `string` | `"description,omitempty"` | Optional free-text details. Omitted from JSON when empty. |
| `Parent` | `*string` | `"parent"` | Pointer to the parent task's ID. `null` in JSON means the task is a root task. Always present in JSON (not omitempty). |
| `Status` | `string` | `"status"` | Lifecycle state. One of `"todo"`, `"in-progress"`, `"done"`. |
| `BlockedBy` | `[]string` | `"blockedBy,omitempty"` | List of task IDs that must reach `"done"` before this task can be started. Omitted from JSON when empty. |
| `Owner` | `*string` | `"owner,omitempty"` | Agent name that has claimed this task. `null` / omitted when unclaimed. |
| `Notes` | `[]Note` | `"notes,omitempty"` | Append-only list of timestamped observations. Omitted from JSON when empty. |
| `Created` | `time.Time` | `"created"` | Creation timestamp. Serialized as RFC3339Nano. |
| `Updated` | `time.Time` | `"updated"` | Last-modified timestamp. Serialized as RFC3339Nano. |

---

## Note

Defined in `internal/models/task.go`.

```go
type Note struct {
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
}
```

| Field | Go type | JSON tag | Description |
|-------|---------|----------|-------------|
| `Content` | `string` | `"content"` | The text of the note. |
| `Timestamp` | `time.Time` | `"timestamp"` | When the note was added. Serialized as RFC3339Nano. |

Notes are append-only. Existing notes are never modified or deleted.

---

## Status Constants

Defined in `internal/models/task.go`.

```go
const (
    StatusTodo       = "todo"
    StatusInProgress = "in-progress"
    StatusDone       = "done"
)
```

| Constant | Value | Meaning |
|----------|-------|---------|
| `StatusTodo` | `"todo"` | Work has not started. |
| `StatusInProgress` | `"in-progress"` | Work is actively underway. |
| `StatusDone` | `"done"` | Work is complete. |

Valid transitions are enforced by commands. Notably: a task cannot be set to `"done"` if it has undone children, and cannot be set to `"in-progress"` if it has incomplete blockers.

---

## TaskStore

Defined in `internal/storage/storage.go`. This is the root object of `.clipm/tasks.json`.

```go
type TaskStore struct {
    Version string        `json:"version"`
    Tasks   []models.Task `json:"tasks"`
}
```

| Field | Go type | JSON tag | Description |
|-------|---------|----------|-------------|
| `Version` | `string` | `"version"` | Schema version. Always `"3.0.0"`. |
| `Tasks` | `[]models.Task` | `"tasks"` | Flat list of all tasks. Relationships (parent/child, blockers) are encoded within each Task. |

---

## NextResult

Defined in `internal/storage/storage.go`. Returned by the `next` command.

```go
type NextResult struct {
    Task         *models.Task  `json:"task,omitempty"`
    Candidates   []models.Task `json:"candidates,omitempty"`
    BlockedCount int           `json:"blockedCount,omitempty"`
}
```

| Field | Go type | JSON tag | Description |
|-------|---------|----------|-------------|
| `Task` | `*models.Task` | `"task,omitempty"` | The single recommended next task. Present when an in-progress task provides context and a specific next step is identified. |
| `Candidates` | `[]models.Task` | `"candidates,omitempty"` | List of candidate tasks when there is no in-progress context to narrow the choice. |
| `BlockedCount` | `int` | `"blockedCount,omitempty"` | Number of tasks skipped because all of their blockers are incomplete. Present when nothing is available. |

Exactly one of `Task` or `Candidates` will be populated in a successful response. `BlockedCount` supplements either field when applicable.

---

## WatchEvent

Defined in `internal/commands/watch.go`. Emitted to stdout by `clipm watch` (JSON mode) whenever the task store changes.

```go
type WatchEvent struct {
    Type      string        `json:"type"`
    Task      *models.Task  `json:"task,omitempty"`
    Tasks     []models.Task `json:"tasks,omitempty"`
    TaskID    string        `json:"taskId,omitempty"`
    Timestamp time.Time     `json:"timestamp"`
}
```

| Field | Go type | JSON tag | Description |
|-------|---------|----------|-------------|
| `Type` | `string` | `"type"` | Event kind. One of `"snapshot"`, `"added"`, `"updated"`, `"deleted"`. |
| `Task` | `*models.Task` | `"task,omitempty"` | The affected task. Present for `"added"` and `"updated"` events. |
| `Tasks` | `[]models.Task` | `"tasks,omitempty"` | Full task list at the time of the snapshot. Present for `"snapshot"` events only. |
| `TaskID` | `string` | `"taskId,omitempty"` | ID of the deleted task. Present for `"deleted"` events only. |
| `Timestamp` | `time.Time` | `"timestamp"` | When the event was emitted. Serialized as RFC3339Nano. |

### WatchEvent type values

| `type` value | When emitted | Populated fields |
|--------------|--------------|-----------------|
| `"snapshot"` | First tick after `watch` starts | `Tasks`, `Timestamp` |
| `"added"` | A task that did not exist in the previous tick now exists | `Task`, `Timestamp` |
| `"updated"` | A task existed in the previous tick and its data has changed | `Task`, `Timestamp` |
| `"deleted"` | A task that existed in the previous tick no longer exists | `TaskID`, `Timestamp` |

---

## tasks.json Example

```json
{
  "version": "3.0.0",
  "tasks": [
    {
      "id": "abcd",
      "name": "Implement authentication",
      "description": "Add JWT-based login and token refresh",
      "parent": null,
      "status": "in-progress",
      "blockedBy": [],
      "owner": "agent-1",
      "notes": [
        {
          "content": "Started with login endpoint",
          "timestamp": "2026-02-20T10:00:00.000000000Z"
        }
      ],
      "created": "2026-02-20T09:00:00.000000000Z",
      "updated": "2026-02-20T10:00:00.000000000Z"
    },
    {
      "id": "efgh",
      "name": "Write login handler",
      "parent": "abcd",
      "status": "todo",
      "created": "2026-02-20T09:01:00.000000000Z",
      "updated": "2026-02-20T09:01:00.000000000Z"
    },
    {
      "id": "ijkl",
      "name": "Write token refresh handler",
      "parent": "abcd",
      "status": "todo",
      "blockedBy": ["efgh"],
      "created": "2026-02-20T09:02:00.000000000Z",
      "updated": "2026-02-20T09:02:00.000000000Z"
    }
  ]
}
```

Notes on the example:

- `"parent": null` serializes as a JSON `null` (the field is never omitted because the struct tag has no `omitempty`).
- Fields with `omitempty` (`description`, `blockedBy`, `owner`, `notes`) are absent from the JSON when empty, as shown for `efgh`.
- `blockedBy` for `ijkl` means `efgh` must reach `"done"` before `ijkl` can be started.
- Timestamps use RFC3339Nano format.
