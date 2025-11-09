# Architectural Decisions

## Project Structure
- **Decision**: Use `internal/` for private packages (models, storage, commands)
- **Rationale**: Prevents external imports, keeps API surface small
- **Date**: 2025-11-08

## Module Name
- **Decision**: Use simple module name `clipm` instead of full GitHub path
- **Rationale**: Can be changed later, keeps it simple for now
- **Date**: 2025-11-08
