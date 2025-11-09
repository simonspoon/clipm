# Libraries

## Dependencies

### github.com/spf13/cobra v1.10.1
- **Purpose**: CLI framework for building command-line applications
- **Usage**: Command structure, flags, argument parsing
- **Alternatives considered**: urfave/cli (cobra chosen for better documentation and larger ecosystem)

### gopkg.in/yaml.v3 (latest)
- **Purpose**: YAML parsing and marshaling
- **Usage**: Task frontmatter serialization/deserialization
- **Alternatives considered**: None - standard YAML library for Go

### github.com/fatih/color v1.18.0
- **Purpose**: Terminal color output
- **Usage**: Colorized status indicators and formatted output
- **Alternatives considered**: None - simple, lightweight, widely used

### github.com/stretchr/testify v1.11.1
- **Purpose**: Testing utilities and assertions
- **Usage**: Unit and integration tests
- **Alternatives considered**: None - de facto standard for Go testing

## Development Dependencies
- golangci-lint (linting)
- Go 1.21+ (runtime)
