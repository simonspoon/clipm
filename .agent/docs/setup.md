# Setup

## Requirements
- Go 1.21+
- golangci-lint (for linting)

## Project Structure
```
clipm/
├── cmd/clipm/           # Main application entry point
├── internal/
│   ├── models/          # Data structures (Task, Index)
│   ├── storage/         # File I/O and index management
│   └── commands/        # Cobra command implementations
├── .agent/              # Documentation
└── prd.md              # Product requirements
```

## Building
```bash
go build -o clipm ./cmd/clipm
```

## Running Tests
```bash
go test ./...
```
