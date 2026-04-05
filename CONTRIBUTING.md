# Contributing to Wellspring

Thank you for your interest in contributing to Wellspring! This document provides guidelines and instructions for contributing.

## Development Environment Setup

### Prerequisites

- **Go:** 1.21 or later
- **Make:** For build automation
- **Git:** For version control

### Getting Started

```bash
# Fork and clone the repository
git clone https://github.com/<your-username>/Wellspring.git
cd Wellspring

# Build
make build

# Run tests
make test

# Clean build artifacts
make clean
```

## Code Style

Wellspring is written in Go. Follow these conventions:

- **Formatting:** Run `gofmt` before committing
- **Naming:** Follow Go naming conventions (`CamelCase` for exports, `camelCase` for unexported)
- **Comments:** Exported functions require GoDoc comments
- **Error handling:** Always handle errors explicitly; no ignored returns
- **Imports:** Group stdlib, external, and internal imports separately

## Testing

Run the full test suite before submitting:

```bash
# Run tests with race detection
go test -race -count=1 ./...

# Build and verify the binary starts
make build
./build/wellspring -h
```

## Pull Request Process

1. **Fork** the repository and create a feature branch:
   ```bash
   git checkout -b feat/my-feature
   ```

2. **Make your changes** with clear, focused commits.

3. **Test thoroughly** with `make test` and verify the binary builds.

4. **Push** your branch and open a Pull Request against `main`.

5. **Describe your changes** in the PR using the provided template.

6. **Respond to review feedback** promptly.

## Commit Message Format

This project follows [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<optional scope>): <description>

[optional body]

[optional footer(s)]
```

### Types

| Type       | Description                          |
| ---------- | ------------------------------------ |
| `feat`     | New feature                          |
| `fix`      | Bug fix                              |
| `docs`     | Documentation changes                |
| `style`    | Formatting, no code change           |
| `refactor` | Code restructuring, no behavior change |
| `test`     | Adding or updating tests             |
| `ci`       | CI/CD changes                        |
| `chore`    | Maintenance, dependencies            |
| `perf`     | Performance improvements             |

### Examples

```
feat(loader): add PowerShell delivery method
fix(token): correct CIDR validation for IPv6 source locks
docs: update delivery methods table with memfd examples
```

### Important

- Do **not** include AI co-author signatures in commits.
- Keep commits focused on a single logical change.

## Questions?

If you have questions about contributing, feel free to open a discussion or issue on GitHub.
