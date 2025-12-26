# Contributing to shape-xml

Thank you for your interest in contributing to shape-xml! This document provides guidelines and instructions for contributing.

## Code of Conduct

This project adheres to the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How to Contribute

### Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When creating a bug report, include:

- **Description**: Clear description of the bug
- **Steps to Reproduce**: Numbered list of steps
- **Expected Behavior**: What should happen
- **Actual Behavior**: What actually happens
- **Environment**: 
  - Go version (`go version`)
  - OS and version
  - shape-xml version

### Suggesting Features

Feature requests are welcome! Please:

1. Check existing feature requests first
2. Describe the problem you're trying to solve
3. Explain your proposed solution
4. Consider alternative solutions

### Pull Requests

1. **Fork and Clone**
   ```bash
   git clone https://github.com/shapestone/shape-xml.git
   cd shape-xml
   ```

2. **Create a Branch**
   ```bash
   git checkout -b feature/my-feature
   # or
   git checkout -b fix/my-bugfix
   ```

3. **Make Changes**
   - Write clean, readable code
   - Follow existing code style
   - Add tests for new functionality
   - Update documentation as needed

4. **Run Tests**
   ```bash
   # Run all tests
   make test
   
   # Check coverage
   make coverage
   
   # Run benchmarks
   make bench
   
   # Run linter (if available)
   golangci-lint run
   ```

5. **Commit Changes**
   - Use clear, descriptive commit messages
   - Follow [Conventional Commits](https://www.conventionalcommits.org/) format:
     - `feat: add new feature`
     - `fix: resolve bug in parser`
     - `docs: update README`
     - `test: add coverage for edge case`
     - `perf: optimize tokenizer performance`

6. **Push and Create PR**
   ```bash
   git push origin feature/my-feature
   ```
   - Create PR on GitHub
   - Fill out PR template completely
   - Link related issues

## Development Setup

### Prerequisites

- Go 1.23 or higher
- Make (for running Makefile targets)

### Local Development

```bash
# Clone the repository
git clone https://github.com/shapestone/shape-xml.git
cd shape-xml

# Run tests
make test

# Run with race detection
go test -race ./...

# Generate coverage report
make coverage

# Run benchmarks
make bench
```

## Code Style

- **Formatting**: Use `gofmt` (automatically applied by editors)
- **Linting**: Run `golangci-lint run` before committing
- **Comments**: Document exported functions, types, and packages
- **Error Messages**: Use clear, actionable error messages
- **Testing**: Write table-driven tests where applicable

### Example Test Pattern

```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "valid input",
            input: "<root/>",
            want:  "<root/>",
        },
        {
            name:    "invalid input",
            input:   "<root>",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Parse(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("Parse() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Testing Requirements

- All new code must include tests
- Aim for >80% coverage for new code
- Test both success and error cases
- Include edge cases and boundary conditions
- Fuzz tests for parsers welcome

## Documentation

- Update README.md if adding new features
- Add godoc comments to all exported symbols
- Update CHANGELOG.md for significant changes
- Include examples in documentation

## Project Structure

```
shape-xml/
├── internal/           # Internal implementation
│   ├── tokenizer/      # XML tokenization
│   ├── parser/         # Full parser with AST
│   └── fastparser/     # Fast validation parser
├── pkg/xml/            # Public API
└── testdata/           # Test data files
```

## Release Process

Maintainers handle releases:

1. Update CHANGELOG.md
2. Update version in relevant files
3. Create git tag: `git tag v0.9.0`
4. Push tag: `git push origin v0.9.0`
5. GitHub Actions creates release

## Questions?

- Open an issue for questions
- Check existing documentation
- Review shape-json for reference implementation

Thank you for contributing to shape-xml!
