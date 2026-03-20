# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build

# Run (requires at least one output flag)
go run main.go metrics.go output.go --console
go run main.go metrics.go output.go --file --web --port 8080

# Dependencies
go mod tidy

# Tests (none currently exist)
go test ./...
```

## Architecture

Three-file Go package (`package main`) for system metrics monitoring:

- **`main.go`** — CLI flag parsing (`--console`, `--file`, `--web`, `--port`), 1-second ticker loop, alert thresholds (CPU >80%, memory >90%, disk >95%), and data model structs (`Metrics`, `MemoryMetrics`, `DiskMetrics`, `NetworkMetrics`, `ProcessMetrics`)
- **`metrics.go`** — Collects system metrics via `gopsutil/v3`: CPU, memory, disk (root), network I/O, and top 10 processes by CPU usage
- **`output.go`** — Three output handlers: formatted console display (top 5 processes), JSON-lines file append, and Gin web server at `/metrics`

**Data flow:** main loop calls `collectMetrics()` → checks thresholds → fans out to enabled output functions. The web server runs in a separate goroutine and receives metrics via a buffered channel (non-blocking send with `select`/`default`).

Output modes are independent and combinable; at least one must be specified or the program exits.
