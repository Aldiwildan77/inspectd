# Examples

## Demo Program

The `demo` directory contains a simple Go program that creates goroutines and allocates memory. This can be used to test inspectd.

### Running the Demo

1. Build and run the demo program:
```bash
cd examples/demo
go run main.go
```

2. In another terminal, run inspectd commands:
```bash
# Inspect runtime information
inspectd runtime

# Inspect memory usage
inspectd memory

# Inspect goroutine count
inspectd goroutines

# Get a complete snapshot
inspectd snapshot
```

### Example Output

When running `inspectd runtime`:
```json
{
  "go_version": "go1.24.5",
  "num_goroutines": 6,
  "gomaxprocs": 8,
  "num_cpu": 8,
  "uptime_seconds": 5.234
}
```

When running `inspectd memory`:
```json
{
  "heap_in_use_bytes": 104857600,
  "heap_allocated_bytes": 104857600,
  "heap_objects": 100,
  "total_alloc_bytes": 104857600,
  "gc_cycles": 0,
  "last_gc_pause_seconds": 0,
  "gc_cpu_fraction": 0
}
```

When running `inspectd snapshot`:
```json
{
  "timestamp": "2025-01-01T12:00:00.123456789Z",
  "runtime": {
    "go_version": "go1.24.5",
    "num_goroutines": 6,
    "gomaxprocs": 8,
    "num_cpu": 8,
    "uptime_seconds": 5.234
  },
  "memory": {
    "heap_in_use_bytes": 104857600,
    "heap_allocated_bytes": 104857600,
    "heap_objects": 100,
    "total_alloc_bytes": 104857600,
    "gc_cycles": 0,
    "last_gc_pause_seconds": 0,
    "gc_cpu_fraction": 0
  },
  "goroutines": {
    "total_count": 6
  }
}
```

## Using with jq

For better readability, pipe output through `jq`:

```bash
inspectd snapshot | jq
inspectd runtime | jq '.num_goroutines'
inspectd memory | jq '.heap_allocated_bytes'
```

## Using in Scripts

Example bash script that monitors memory:

```bash
#!/bin/bash
while true; do
  inspectd memory | jq -r '.heap_allocated_bytes'
  sleep 1
done
```

