# inspectd

A CLI tool and SDK for inspecting Go runtime internals from inside a running process. Designed for AI agent consumption with structured, deterministic JSON output.

## Features

- **CLI Tool**: Read-only inspection of Go runtime metrics via command-line
- **SDK**: Programmatic API for collecting and storing runtime snapshots
- **JSON-First**: Structured output optimized for automation and AI agents
- **Production-Safe**: Minimal overhead, no background processes
- **Multiple Storage Backends**: Memory, file, database, and object storage support
- **Resource-Aware**: Container/pod-friendly storage implementations

## What inspectd is

- A read-only inspection tool for Go runtime metrics
- A JSON-first CLI tool with no human-facing output
- A composable tool designed for pipes and automation
- A production-safe tool with minimal overhead
- An SDK for programmatic runtime inspection and storage

## What inspectd is NOT

- Not an HTTP server or API
- Not a continuous monitoring tool
- Not a debugging tool with explanations
- Not a human-readable diagnostic tool
- Not a tool that modifies runtime behavior

## Installation

```bash
go install github.com/Aldiwildan77/inspectd/cmd/inspectd@latest
```

## Commands

### `inspectd runtime`

Reports Go version, goroutine count, GOMAXPROCS, CPU count, and process uptime.

### `inspectd memory`

Reports heap usage, allocations, GC cycles, and GC statistics.

### `inspectd goroutines`

Reports total goroutine count.

### `inspectd snapshot`

Combines runtime, memory, and goroutine information with a timestamp. Designed for agent ingestion.

## Usage for AI Agents

All commands output JSON to stdout. Errors result in non-zero exit codes.

```bash
inspectd runtime | jq
inspectd memory | jq
inspectd snapshot | jq
```

The tool is designed to be invoked on-demand. It does not maintain state, background goroutines, or continuous sampling.

## SDK Usage

The inspectd SDK provides a programmatic API for collecting and storing runtime snapshots.

### Basic Example

```go
package main

import (
    "context"
    "github.com/Aldiwildan77/inspectd/sdk"
    "github.com/Aldiwildan77/inspectd/sdk/storage"
)

func main() {
    // Create storage backend
    memStorage := storage.NewMemoryStorage()
    defer memStorage.Close()
    
    // Create client with functional options
    client := sdk.NewClient(sdk.WithStorage(memStorage))
    defer client.Close()
    
    // Collect and store a snapshot
    ctx := context.Background()
    if err := client.CollectAndStore(ctx); err != nil {
        panic(err)
    }
    
    // Query recent snapshots
    snapshots, err := client.QueryRecent(ctx, 10)
    if err != nil {
        panic(err)
    }
    
    // Use snapshots...
}
```

### Storage Backends

The SDK supports multiple storage backends:

- **Memory Storage**: In-memory storage for testing and development
- **File Storage**: Persistent file-based storage
- **Database Storage**: SQL database storage (PostgreSQL, MySQL, SQLite)
- **Object Storage**: S3-compatible object storage
- **Bounded Memory**: Memory storage with size limits for containers/pods

See [examples/](examples/) for more usage examples and [docs/SDK.md](docs/SDK.md) for complete SDK documentation.

## Output Format

All commands output JSON. The snapshot command provides a combined view:

```json
{
  "timestamp": "2024-01-01T00:00:00Z",
  "runtime": {...},
  "memory": {...},
  "goroutines": {...}
}
```

## Documentation

- [SDK Documentation](docs/SDK.md) - Complete SDK API reference and examples
- [Production Guide](docs/PRODUCTION.md) - Production deployment and storage strategies
- [MCP Integration](docs/MCP.md) - Model Context Protocol integration
- [Product Requirements](docs/PRD.md) - Detailed product specification

## Safety

- Read-only operations
- No global mutable state
- Graceful error handling
- Standard library only (CLI)
- No unsafe operations
- Production-safe with minimal overhead

## License

MIT License - see [LICENSE](LICENSE) for details.
