# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
go build ./...          # Build all packages
go test ./...           # Run all tests
go test -v ./...        # Run tests with verbose output
go test -run TestName   # Run a specific test
go vet ./...            # Run static analysis
go fmt ./...            # Format code
```

## Project Overview

tech-utils is a Go module for utility functions.


### General
- Prefer small functions over large ones
- Use dependency injection with interfaces for testability
- Do not use emojis in logs or comments
- Log should be emitted with the stdlib `slog.Logger` structure logging library
- After any changes are made, they should be reviewed for any security issues as well as for correctness

### Testing
- Use **map-based table tests** with test names as keys (`map[string]struct{}`), not slices
- Use `t.Cleanup()` instead of `defer` for test cleanup
- Tests should be deterministic and parallelizable
- Tests should use the `require` library for assertions

