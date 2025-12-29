# inspectd

A CLI tool that inspects Go runtime internals from inside a running process. Designed for AI agent consumption with structured, deterministic JSON output.

## What inspectd is

- A read-only inspection tool for Go runtime metrics
- A JSON-first CLI tool with no human-facing output
- A composable tool designed for pipes and automation
- A production-safe tool with minimal overhead

## What inspectd is NOT

- Not an HTTP server or API
- Not a continuous monitoring tool
- Not a debugging tool with explanations
- Not a human-readable diagnostic tool
- Not a tool that modifies runtime behavior

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

See [examples/](examples/) for usage examples and a demo program.

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

## Safety

- Read-only operations
- No global mutable state
- Graceful error handling
- Standard library only
- No unsafe operations
