# Product Requirement Document (PRD)

## inspectd - Go Runtime Inspection Tool

**Version:** 1.0.0  
**Last Updated:** 2025  
**Status:** Implemented (CLI + SDK)

---

## 1. Executive Summary

**inspectd** is a command-line tool and SDK designed to inspect Go runtime internals from within a running Go process. The tool provides structured, deterministic JSON output optimized for AI agent consumption and automation workflows. It is a read-only, production-safe inspection utility with minimal overhead.

### Key Characteristics

- **Read-only**: No modifications to runtime behavior
- **JSON-first**: Structured output for programmatic consumption
- **Composable**: Designed for pipes and automation
- **Production-safe**: Minimal overhead, no background processes
- **AI-optimized**: Deterministic output format for agent ingestion
- **SDK-enabled**: Programmatic API for collecting and storing snapshots
- **Production-ready**: Multiple storage backends with resource management

---

## 2. Problem Statement

### 2.1 Current Challenges

1. **Lack of Structured Runtime Inspection**
   - Existing Go debugging tools provide human-readable output
   - No standardized JSON format for programmatic access
   - Difficult to integrate runtime metrics into automated systems

2. **AI Agent Integration Gaps**
   - AI agents need structured, deterministic data formats
   - Human-readable diagnostic tools are not machine-parseable
   - No lightweight tool for on-demand runtime inspection

3. **Monitoring and Automation Needs**
   - Continuous monitoring tools are heavy and resource-intensive
   - Need for lightweight, on-demand inspection capabilities
   - Integration with CI/CD pipelines and automated testing

4. **Production Storage Requirements**
   - Need to store snapshots for historical analysis
   - Container/pod environments require resource-aware storage
   - Multiple storage backends needed for different deployment scenarios

### 2.2 Target Use Cases

- **AI Agent Runtime Analysis**: Provide structured runtime data to AI agents for code analysis and debugging
- **Automated Monitoring**: Lightweight inspection for scripts and automation tools
- **Development Tooling**: Quick runtime checks during development
- **Production Diagnostics**: Safe, read-only inspection in production environments
- **Production Storage**: Store snapshots for historical analysis and trend tracking
- **Container/Pod Deployment**: Resource-aware storage for Kubernetes and containerized environments

---

## 3. Goals and Objectives

### 3.1 Primary Goals

1. **Provide Structured Runtime Data**
   - Expose Go runtime metrics in JSON format
   - Support runtime, memory, and goroutine information
   - Enable programmatic consumption of runtime state

2. **Optimize for Automation**
   - Design for pipes and command chaining
   - Deterministic output format
   - Error handling via exit codes

3. **Ensure Production Safety**
   - Read-only operations only
   - No global mutable state
   - Minimal performance overhead
   - No background goroutines or continuous sampling

4. **AI Agent Compatibility**
   - Structured, parseable output
   - Consistent data format
   - Machine-readable error codes

5. **Production Storage Support**
   - Multiple storage backends (memory, file, database, object storage)
   - Resource management and cleanup policies
   - Production-safe implementations for containers/pods

### 3.2 Success Criteria

- ✅ All commands output valid JSON
- ✅ Zero runtime modifications
- ✅ Sub-millisecond execution time
- ✅ Standard library only (no external dependencies for CLI)
- ✅ Comprehensive error handling
- ✅ SDK with production-ready storage backends
- ✅ Resource-aware storage for container/pod environments

---

## 4. Target Users

### 4.1 Primary Users

1. **AI Agents and Automation Systems**
   - Need structured, deterministic data
   - Require consistent output format
   - Must handle errors programmatically

2. **DevOps Engineers**
   - Integration with monitoring scripts
   - CI/CD pipeline integration
   - Production diagnostics

3. **Go Developers**
   - Quick runtime checks during development
   - Debugging assistance
   - Performance analysis

### 4.2 User Personas

**Persona 1: AI Agent**

- Needs: Structured JSON output, deterministic format
- Goals: Parse runtime state, make decisions based on metrics
- Constraints: Cannot handle human-readable text

**Persona 2: Automation Script**

- Needs: Command-line interface, exit codes, pipe-friendly
- Goals: Integrate into monitoring workflows
- Constraints: Must be scriptable and reliable

**Persona 3: Developer**

- Needs: Quick insights, easy to use
- Goals: Understand runtime state during development
- Constraints: Should not impact running application

---

## 5. Features and Requirements

### 5.1 Core Features

#### 5.1.1 Runtime Information Command (`inspectd runtime`)

**Description**: Reports basic Go runtime information

**Output Fields**:

- `go_version` (string): Go version string
- `num_goroutines` (int): Current number of goroutines
- `gomaxprocs` (int): GOMAXPROCS setting
- `num_cpu` (int): Number of CPU cores
- `uptime_seconds` (float64): Process uptime in seconds

**Requirements**:

- Must output valid JSON
- Must use UTC timestamps
- Must handle errors gracefully

**Example Output**:

```json
{
  "go_version": "go1.24.5",
  "num_goroutines": 6,
  "gomaxprocs": 8,
  "num_cpu": 8,
  "uptime_seconds": 5.234
}
```

#### 5.1.2 Memory Information Command (`inspectd memory`)

**Description**: Reports memory usage and garbage collection statistics

**Output Fields**:

- `heap_in_use_bytes` (uint64): Heap memory in use
- `heap_allocated_bytes` (uint64): Currently allocated heap memory
- `heap_objects` (uint64): Number of heap objects
- `total_alloc_bytes` (uint64): Total bytes allocated
- `gc_cycles` (uint32): Number of GC cycles
- `last_gc_pause_seconds` (float64): Last GC pause duration
- `gc_cpu_fraction` (float64): Fraction of CPU time spent in GC

**Requirements**:

- Must read from `runtime.MemStats`
- Must calculate GC pause from pause history
- Must handle zero GC cycles gracefully

**Example Output**:

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

#### 5.1.3 Goroutine Information Command (`inspectd goroutines`)

**Description**: Reports goroutine count

**Output Fields**:

- `total_count` (int): Total number of goroutines

**Requirements**:

- Must use `runtime.NumGoroutine()`
- Must output valid JSON

**Example Output**:

```json
{
  "total_count": 6
}
```

#### 5.1.4 Snapshot Command (`inspectd snapshot`)

**Description**: Combines all runtime information into a single snapshot

**Output Structure**:

- `timestamp` (string): RFC3339Nano formatted timestamp
- `runtime` (object): Runtime information
- `memory` (object): Memory information
- `goroutines` (object): Goroutine information

**Requirements**:

- Must include UTC timestamp
- Must combine all three information sources
- Must maintain consistent structure

**Example Output**:

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

#### 5.1.5 SDK Features

**Description**: Programmatic API for collecting and storing runtime snapshots

**Core Components**:

1. **SDK Client**: High-level API for snapshot collection and storage
2. **Storage Interface**: Pluggable storage backend system
3. **Production Storage Backends**:
   - **BoundedMemoryStorage**: In-memory storage with size limits and automatic eviction
   - **ManagedFileStorage**: File-based storage with cleanup and retention policies
   - **DatabaseStorage**: SQL database storage (PostgreSQL, MySQL, etc.)
   - **CloudObjectStorage**: Object storage interface (S3, GCS, Azure Blob, etc.)

**Requirements**:

- Must provide thread-safe operations
- Must support context timeouts for all operations
- Must implement graceful shutdown
- Must include resource management (limits, cleanup)
- Must be production-safe for container/pod environments

**Example Usage**:

```go
// Bounded memory storage (production-safe)
memStorage := storage.NewBoundedMemoryStorage(1000)
client := sdk.NewClient(memStorage)

// Collect and store
ctx := context.Background()
client.CollectAndStore(ctx)

// Query recent snapshots
snapshots, _ := client.QueryRecent(ctx, 10)
```

### 5.2 Functional Requirements

#### FR-1: Command-Line Interface

- **Requirement**: Tool must accept commands as first argument
- **Validation**: Invalid commands must exit with non-zero code
- **Implementation**: `os.Args[1]` parsing in CLI module

#### FR-2: JSON Output

- **Requirement**: All commands must output valid JSON to stdout
- **Validation**: Output must be parseable by standard JSON parsers
- **Implementation**: `encoding/json` package for marshaling

#### FR-3: Error Handling

- **Requirement**: Errors must result in non-zero exit codes
- **Validation**: No JSON output on error
- **Implementation**: `os.Exit(1)` on errors

#### FR-4: Read-Only Operations

- **Requirement**: No modifications to runtime state
- **Validation**: Only read operations from `runtime` package
- **Implementation**: No write operations, no state mutations

#### FR-5: No Background Processes

- **Requirement**: Tool must be stateless and on-demand
- **Validation**: No goroutines, no timers, no continuous sampling
- **Implementation**: Single execution, immediate return

#### FR-6: SDK Storage Interface

- **Requirement**: SDK must support pluggable storage backends
- **Validation**: All storage backends implement common interface
- **Implementation**: `storage.Storage` interface with Store, StoreBatch, Query, Close methods

#### FR-7: Production Storage Backends

- **Requirement**: Must provide production-ready storage implementations
- **Validation**: Resource limits, cleanup policies, context timeouts
- **Implementation**: BoundedMemoryStorage, ManagedFileStorage, DatabaseStorage, CloudObjectStorage

#### FR-8: Resource Management

- **Requirement**: Storage backends must prevent resource exhaustion
- **Validation**: Memory limits, disk cleanup, connection pooling
- **Implementation**: Size limits, retention policies, automatic cleanup

### 5.3 Non-Functional Requirements

#### NFR-1: Performance

- **Target**: Sub-millisecond execution time
- **Measurement**: Time from command invocation to output
- **Constraint**: Must not impact running application

#### NFR-2: Resource Usage

- **Target**: Minimal memory footprint
- **Measurement**: Memory allocated during execution
- **Constraint**: Should be negligible compared to application

#### NFR-3: Dependencies

- **Requirement**: Standard library only
- **Validation**: No external package dependencies
- **Implementation**: Only `encoding/json`, `runtime`, `time`, `os`, `fmt`

#### NFR-4: Compatibility

- **Requirement**: Works with all Go versions 1.18+
- **Validation**: Tested across Go versions
- **Implementation**: Use only stable runtime APIs

#### NFR-5: Safety

- **Requirement**: Production-safe operations
- **Validation**: No unsafe operations, no panics
- **Implementation**: Error handling, graceful degradation

#### NFR-6: Container/Pod Compatibility

- **Requirement**: Must work safely in containerized environments
- **Validation**: Resource limits respected, graceful shutdown, no OOM risks
- **Implementation**: Bounded storage, cleanup policies, context timeouts

#### NFR-7: Storage Scalability

- **Requirement**: Storage backends must scale with usage
- **Validation**: Efficient batch operations, connection pooling, cleanup
- **Implementation**: StoreBatch methods, connection limits, automatic cleanup

---

## 6. Technical Specifications

### 6.1 Architecture

```
┌─────────────────┐
│   cmd/inspectd  │
│     main.go     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  internal/cli   │
│     cli.go      │
└────────┬────────┘
         │
    ┌────┴────┬──────────┬─────────────┐
    ▼         ▼          ▼             ▼
┌────────┐ ┌───────┐ ┌──────────┐ ┌──────────┐
│runtime │ │memory │ │goroutines│ │ snapshot │
│ info   │ │ info  │ │   info   │ │          │
└────────┘ └───────┘ └──────────┘ └──────────┘
    │         │          │             │
    └─────────┴──────────┴─────────────┘
                    │
                    ▼
            ┌───────────────┐
            │ runtime pkg   │
            │ encoding/json │
            └───────────────┘

┌─────────────────────────────────────────┐
│              SDK Layer                   │
│  pkg/sdk/client.go                      │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│         Storage Interface               │
│  pkg/sdk/storage/interface.go           │
└──────────────┬──────────────────────────┘
               │
    ┌──────────┼──────────┬──────────────┐
    ▼          ▼          ▼              ▼
┌─────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐
│Bounded  │ │ Managed  │ │Database  │ │Cloud Object  │
│Memory   │ │  File    │ │ Storage  │ │  Storage    │
└─────────┘ └──────────┘ └──────────┘ └──────────────┘
```

### 6.2 Module Structure

```
inspectd/
├── cmd/
│   └── inspectd/
│       └── main.go          # CLI entry point
├── internal/
│   ├── cli/
│   │   └── cli.go           # CLI routing
│   ├── runtimeinfo/
│   │   └── runtime.go       # Runtime metrics
│   ├── memory/
│   │   └── memory.go        # Memory metrics
│   ├── goroutines/
│   │   └── goroutines.go    # Goroutine metrics
│   └── snapshot/
│       └── snapshot.go      # Combined snapshot
├── pkg/
│   └── sdk/
│       ├── client.go        # SDK client
│       ├── types/
│       │   └── snapshot.go  # SDK types
│       ├── storage/
│       │   ├── interface.go      # Storage interface
│       │   ├── memory.go         # Basic memory storage
│       │   ├── bounded_memory.go # Production memory storage
│       │   ├── file.go            # Basic file storage
│       │   ├── managed_file.go   # Production file storage
│       │   ├── database.go       # Database storage
│       │   └── object_storage.go # Object storage
│       └── examples/
│           ├── basic/            # Basic examples
│           ├── production/      # Production examples
│           └── custom_storage/  # Custom storage example
├── examples/
│   └── demo/
│       └── main.go          # Demo program
└── docs/
    ├── PRD.md               # This document
    ├── SDK.md               # SDK documentation
    └── PRODUCTION.md         # Production guide
```

### 6.3 Data Models

#### RuntimeInfo

```go
type RuntimeInfo struct {
    GoVersion     string  `json:"go_version"`
    NumGoroutines int     `json:"num_goroutines"`
    GOMAXPROCS    int     `json:"gomaxprocs"`
    NumCPU        int     `json:"num_cpu"`
    Uptime        float64 `json:"uptime_seconds"`
}
```

#### MemoryInfo

```go
type MemoryInfo struct {
    HeapInUse     uint64  `json:"heap_in_use_bytes"`
    HeapAllocated uint64  `json:"heap_allocated_bytes"`
    HeapObjects   uint64  `json:"heap_objects"`
    TotalAlloc    uint64  `json:"total_alloc_bytes"`
    GCCycles      uint32  `json:"gc_cycles"`
    LastGCPause   float64 `json:"last_gc_pause_seconds"`
    GCCPUFraction float64 `json:"gc_cpu_fraction"`
}
```

#### GoroutineInfo

```go
type GoroutineInfo struct {
    TotalCount int `json:"total_count"`
}
```

#### Snapshot

```go
type Snapshot struct {
    Timestamp  string                    `json:"timestamp"`
    Runtime    *runtimeinfo.RuntimeInfo  `json:"runtime"`
    Memory     *memory.MemoryInfo        `json:"memory"`
    Goroutines *goroutines.GoroutineInfo `json:"goroutines"`
}
```

### 6.4 API Specifications

#### Command Interface

**Pattern**: `inspectd <command>`

**Commands**:

- `runtime`: Get runtime information
- `memory`: Get memory information
- `goroutines`: Get goroutine count
- `snapshot`: Get combined snapshot

**Exit Codes**:

- `0`: Success
- `1`: Error (invalid command, collection failure, JSON marshaling error)

**Output**:

- **Success**: Valid JSON to stdout
- **Error**: No output, exit code 1

### 6.5 Implementation Details

#### Runtime Information Collection

- Uses `runtime.Version()` for Go version
- Uses `runtime.NumGoroutine()` for goroutine count
- Uses `runtime.GOMAXPROCS(0)` for GOMAXPROCS (read-only)
- Uses `runtime.NumCPU()` for CPU count
- Tracks process start time for uptime calculation

#### Memory Information Collection

- Uses `runtime.ReadMemStats()` to read memory statistics
- Calculates last GC pause from `PauseNs` circular buffer
- Handles zero GC cycles (returns 0 for pause time)

#### Goroutine Information Collection

- Uses `runtime.NumGoroutine()` for count
- Simple wrapper for consistency with other commands

#### Snapshot Collection

- Collects all three information sources
- Combines into single structure
- Adds UTC timestamp in RFC3339Nano format

---

## 7. Usage Patterns

### 7.1 Basic Usage

```bash
# Get runtime information
inspectd runtime

# Get memory information
inspectd memory

# Get goroutine count
inspectd goroutines

# Get complete snapshot
inspectd snapshot
```

### 7.2 With jq (JSON Processor)

```bash
# Pretty print
inspectd snapshot | jq

# Extract specific fields
inspectd runtime | jq '.num_goroutines'
inspectd memory | jq '.heap_allocated_bytes'
```

### 7.3 In Scripts

```bash
#!/bin/bash
# Monitor memory usage
while true; do
  inspectd memory | jq -r '.heap_allocated_bytes'
  sleep 1
done
```

### 7.4 AI Agent Integration

```python
import subprocess
import json

def get_runtime_snapshot():
    result = subprocess.run(
        ['inspectd', 'snapshot'],
        capture_output=True,
        text=True
    )
    if result.returncode == 0:
        return json.loads(result.stdout)
    return None
```

---

## 8. Constraints and Limitations

### 8.1 Design Constraints

1. **Read-Only Constraint**
   - Cannot modify runtime behavior
   - Cannot trigger GC or force goroutine scheduling
   - Cannot change GOMAXPROCS or other settings

2. **No External Dependencies**
   - Must use only standard library
   - Cannot add third-party packages
   - Limits advanced features

3. **On-Demand Only**
   - No continuous monitoring
   - No background processes
   - No state persistence

4. **Single Process Scope**
   - Only inspects the current process
   - Cannot inspect other processes
   - No remote inspection capabilities

### 8.2 Known Limitations

1. **GC Pause History**
   - Only last GC pause available
   - Limited to 256 pause history entries
   - No average or percentile calculations

2. **Goroutine Details**
   - Only total count available
   - No individual goroutine information
   - No stack traces or goroutine states

3. **Memory Granularity**
   - Heap-level information only
   - No per-type allocation breakdown
   - No stack memory information

4. **No Historical Data**
   - Each invocation is independent
   - No trend analysis
   - No rate calculations

---

## 9. Testing Requirements

### 9.1 Unit Tests

- Test each collection function independently
- Verify JSON marshaling correctness
- Test error handling paths
- Validate data types and ranges

### 9.2 Integration Tests

- Test CLI command routing
- Test exit codes
- Test JSON output validity
- Test with demo program

### 9.3 Performance Tests

- Measure execution time
- Measure memory allocation
- Verify no performance degradation
- Test under load

### 9.4 Compatibility Tests

- Test across Go versions
- Test on different platforms
- Test with various runtime configurations

---

## 10. Future Considerations

### 10.1 Potential Enhancements

1. **Extended Metrics**
   - CPU usage statistics
   - Network I/O statistics
   - File descriptor counts
   - CGO call statistics

2. **Goroutine Details**
   - Stack traces for all goroutines
   - Goroutine state breakdown
   - Blocked goroutine analysis

3. **Memory Analysis**
   - Per-type allocation breakdown
   - Stack memory information
   - GC heap size breakdown

4. **Historical Tracking** ✅ (Partially Implemented)
   - ✅ Snapshot storage and querying (via SDK)
   - ✅ Time-range queries
   - ⏳ Rate calculations (goroutines/sec, allocations/sec)
   - ⏳ Trend analysis
   - ⏳ Delta calculations between snapshots

5. **Filtering and Querying** ✅ (Partially Implemented)
   - ✅ Time-range filtering
   - ✅ Limit and ordering
   - ⏳ Field selection
   - ⏳ Conditional output
   - ⏳ Format options (compact vs. pretty)

### 10.2 Out of Scope

- HTTP API server
- Continuous monitoring mode
- Remote process inspection
- Profiling capabilities
- Debugging features
- Human-readable output modes

### 10.3 Recently Implemented (v1.1.0)

1. **SDK for Programmatic Access** ✅
   - High-level client API
   - Snapshot collection and storage
   - Query capabilities

2. **Production Storage Backends** ✅
   - BoundedMemoryStorage with size limits
   - ManagedFileStorage with cleanup policies
   - DatabaseStorage for SQL databases
   - CloudObjectStorage interface for object storage

3. **Production Deployment Support** ✅
   - Container/pod resource management
   - Graceful shutdown handling
   - Context timeout support
   - Automatic cleanup and retention policies

---

## 11. Success Metrics

### 11.1 Technical Metrics

- ✅ **Execution Time**: < 1ms for all commands
- ✅ **Memory Overhead**: < 1MB per invocation
- ✅ **JSON Validity**: 100% valid JSON output
- ✅ **Error Rate**: 0% for valid commands

### 11.2 Functional Metrics

- ✅ **Command Coverage**: All 4 commands implemented
- ✅ **Output Consistency**: Deterministic JSON structure
- ✅ **Error Handling**: Proper exit codes for all error cases
- ✅ **Documentation**: Complete usage examples
- ✅ **SDK Implementation**: Full SDK with client API
- ✅ **Storage Backends**: 4 production-ready storage implementations
- ✅ **Production Features**: Resource management, cleanup, timeouts

### 11.3 Quality Metrics

- ✅ **Code Quality**: Standard library only, no unsafe operations
- ✅ **Safety**: Read-only operations, no state mutations
- ✅ **Reliability**: Graceful error handling, no panics
- ✅ **Maintainability**: Clear module structure, documented code

---

## 12. Dependencies and Requirements

### 12.1 Runtime Requirements

- **Go Version**: 1.18 or higher
- **Platform**: Linux, macOS, Windows
- **Architecture**: amd64, arm64

### 12.2 Build Requirements

- Go compiler (1.18+)
- Standard Go toolchain
- No external build tools

### 12.3 Runtime Dependencies

- None (standard library only)

---

## 13. Documentation Requirements

### 13.1 User Documentation

- ✅ README.md with usage examples
- ✅ Examples directory with demo program
- ✅ Command reference documentation
- ✅ JSON schema documentation
- ✅ SDK documentation (SDK.md)
- ✅ Production deployment guide (PRODUCTION.md)
- ✅ Production usage examples

### 13.2 Developer Documentation

- ✅ Code comments
- ✅ Module documentation
- ✅ Architecture documentation
- ✅ This PRD document
- ✅ SDK API reference
- ✅ Storage interface documentation
- ✅ Production best practices

---

## 14. Version History

### Version 1.1.0 (Current)

- **SDK Implementation**
  - High-level client API for programmatic access
  - Snapshot collection and storage API
  - Query capabilities with time-range filtering

- **Production Storage Backends**
  - BoundedMemoryStorage: In-memory storage with size limits and automatic eviction
  - ManagedFileStorage: File-based storage with cleanup and retention policies
  - DatabaseStorage: SQL database storage (PostgreSQL, MySQL support)
  - CloudObjectStorage: Interface for object storage (S3, GCS, Azure Blob)

- **Production Features**
  - Resource management and limits
  - Automatic cleanup and retention policies
  - Context timeout support
  - Graceful shutdown handling
  - Thread-safe operations
  - Container/pod compatibility

- **Documentation**
  - SDK documentation (SDK.md)
  - Production deployment guide (PRODUCTION.md)
  - Production usage examples

### Version 1.0.0

- Initial implementation
- All core commands (runtime, memory, goroutines, snapshot)
- JSON output format
- CLI interface
- Demo program
- Documentation

---

## 15. Approval and Sign-off

**Product Owner**: [To be filled]  
**Technical Lead**: [To be filled]  
**Date**: [To be filled]

---

## Appendix A: JSON Schema

### Runtime Info Schema

```json
{
  "type": "object",
  "properties": {
    "go_version": {"type": "string"},
    "num_goroutines": {"type": "integer"},
    "gomaxprocs": {"type": "integer"},
    "num_cpu": {"type": "integer"},
    "uptime_seconds": {"type": "number"}
  },
  "required": ["go_version", "num_goroutines", "gomaxprocs", "num_cpu", "uptime_seconds"]
}
```

### Memory Info Schema

```json
{
  "type": "object",
  "properties": {
    "heap_in_use_bytes": {"type": "integer"},
    "heap_allocated_bytes": {"type": "integer"},
    "heap_objects": {"type": "integer"},
    "total_alloc_bytes": {"type": "integer"},
    "gc_cycles": {"type": "integer"},
    "last_gc_pause_seconds": {"type": "number"},
    "gc_cpu_fraction": {"type": "number"}
  },
  "required": ["heap_in_use_bytes", "heap_allocated_bytes", "heap_objects", "total_alloc_bytes", "gc_cycles", "last_gc_pause_seconds", "gc_cpu_fraction"]
}
```

### Goroutine Info Schema

```json
{
  "type": "object",
  "properties": {
    "total_count": {"type": "integer"}
  },
  "required": ["total_count"]
}
```

### Snapshot Schema

```json
{
  "type": "object",
  "properties": {
    "timestamp": {"type": "string", "format": "date-time"},
    "runtime": {"$ref": "#/definitions/RuntimeInfo"},
    "memory": {"$ref": "#/definitions/MemoryInfo"},
    "goroutines": {"$ref": "#/definitions/GoroutineInfo"}
  },
  "required": ["timestamp", "runtime", "memory", "goroutines"]
}
```

---

**Document End**
