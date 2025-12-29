# inspectd SDK Documentation

## Overview

The inspectd SDK provides a programmatic interface for collecting and storing Go runtime snapshots. It is designed to be easy to understand and integrate, especially for AI agents and automation systems.

## Quick Start

### Installation

```go
import "github.com/Aldiwildan77/inspectd/pkg/sdk"
import "github.com/Aldiwildan77/inspectd/pkg/sdk/storage"
```

### Basic Usage

```go
package main

import (
    "context"
    "github.com/Aldiwildan77/inspectd/pkg/sdk"
    "github.com/Aldiwildan77/inspectd/pkg/sdk/storage"
)

func main() {
    // 1. Create a storage backend (in-memory for this example)
    memStorage := storage.NewMemoryStorage()
    defer memStorage.Close()

    // 2. Create an SDK client
    client := sdk.NewClient(memStorage)

    // 3. Collect and store a snapshot
    ctx := context.Background()
    if err := client.CollectAndStore(ctx); err != nil {
        panic(err)
    }

    // 4. Query stored snapshots
    snapshots, err := client.QueryRecent(ctx, 10)
    if err != nil {
        panic(err)
    }

    // Use the snapshots...
}
```

## Core Concepts

### 1. Snapshot

A `Snapshot` is a complete picture of the Go runtime at a specific point in time. It contains:

- **Timestamp**: When the snapshot was collected (RFC3339Nano format)
- **Runtime**: Go version, goroutine count, CPU info, uptime
- **Memory**: Heap usage, allocations, GC statistics
- **Goroutines**: Total goroutine count

```go
type Snapshot struct {
    Timestamp  string
    Runtime    *RuntimeInfo
    Memory     *MemoryInfo
    Goroutines *GoroutineInfo
}
```

### 2. Storage Interface

The SDK uses a storage interface pattern, allowing you to store snapshots to any backend:

- **MemoryStorage**: In-memory storage (for testing/caching)
- **FileStorage**: File-based storage (JSON files in a directory)
- **Custom Storage**: Implement your own storage backend

### 3. Client

The `Client` is the main entry point. It provides:

- `CollectSnapshot()`: Collect a snapshot from the current process
- `CollectAndStore()`: Collect and store in one operation
- `Store()`: Store an existing snapshot
- `Query()`: Query stored snapshots
- `QueryRecent()`: Get the most recent snapshots
- `QueryByTimeRange()`: Get snapshots within a time range

## Storage Backends

### Memory Storage

**Use Case**: Testing, caching, temporary storage

```go
memStorage := storage.NewMemoryStorage()
defer memStorage.Close()

client := sdk.NewClient(memStorage)
```

**Characteristics**:
- Fast and simple
- Data lost on process exit
- Thread-safe
- No external dependencies

### File Storage

**Use Case**: Simple persistence, debugging, local storage

```go
fileStorage, err := storage.NewFileStorage("./snapshots")
if err != nil {
    panic(err)
}
defer fileStorage.Close()

client := sdk.NewClient(fileStorage)
```

**Characteristics**:
- Each snapshot stored as a JSON file
- Files named by timestamp
- Persistent across process restarts
- Human-readable JSON format

### Custom Storage

You can implement your own storage backend for databases, APIs, or any other system:

```go
type MyStorage struct {
    // Your storage implementation
}

func (m *MyStorage) Store(ctx context.Context, snapshot *types.Snapshot) error {
    // Implement storage logic
    return nil
}

func (m *MyStorage) StoreBatch(ctx context.Context, snapshots []*types.Snapshot) error {
    // Implement batch storage
    return nil
}

func (m *MyStorage) Query(ctx context.Context, opts *storage.QueryOptions) ([]*types.Snapshot, error) {
    // Implement query logic
    return nil, nil
}

func (m *MyStorage) Close() error {
    // Cleanup resources
    return nil
}

// Use it
myStorage := &MyStorage{}
client := sdk.NewClient(myStorage)
```

## API Reference

### Client Methods

#### `NewClient(storage Storage) *Client`

Creates a new SDK client with the specified storage backend.

**Parameters**:
- `storage`: Any implementation of `storage.Storage` interface

**Returns**: A new `Client` instance

#### `CollectSnapshot() (*types.Snapshot, error)`

Collects a runtime snapshot from the current Go process.

**Returns**:
- `*types.Snapshot`: The collected snapshot
- `error`: Any error that occurred during collection

**Example**:
```go
snapshot, err := client.CollectSnapshot()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Goroutines: %d\n", snapshot.Goroutines.TotalCount)
```

#### `CollectAndStore(ctx context.Context) error`

Collects a snapshot and stores it in one operation. This is the most common use case.

**Parameters**:
- `ctx`: Context for cancellation/timeout

**Returns**: Error if collection or storage fails

**Example**:
```go
ctx := context.Background()
if err := client.CollectAndStore(ctx); err != nil {
    log.Fatal(err)
}
```

#### `Store(ctx context.Context, snapshot *types.Snapshot) error`

Stores an existing snapshot to the storage backend.

**Parameters**:
- `ctx`: Context for cancellation/timeout
- `snapshot`: The snapshot to store

**Returns**: Error if storage fails

**Example**:
```go
snapshot, _ := client.CollectSnapshot()
err := client.Store(ctx, snapshot)
```

#### `StoreBatch(ctx context.Context, snapshots []*types.Snapshot) error`

Stores multiple snapshots in a single operation. More efficient than multiple `Store` calls.

**Parameters**:
- `ctx`: Context for cancellation/timeout
- `snapshots`: Slice of snapshots to store

**Returns**: Error if storage fails

**Example**:
```go
snapshots := []*types.Snapshot{snapshot1, snapshot2, snapshot3}
err := client.StoreBatch(ctx, snapshots)
```

#### `Query(ctx context.Context, opts *storage.QueryOptions) ([]*types.Snapshot, error)`

Queries stored snapshots with flexible filtering options.

**Parameters**:
- `ctx`: Context for cancellation/timeout
- `opts`: Query options (time range, limit, ordering)

**Returns**: Slice of matching snapshots, or error

**Example**:
```go
opts := &storage.QueryOptions{
    Limit:   10,
    OrderBy: storage.OrderByTimeDesc,
}
snapshots, err := client.Query(ctx, opts)
```

#### `QueryRecent(ctx context.Context, limit int) ([]*types.Snapshot, error)`

Convenience method to get the most recent snapshots.

**Parameters**:
- `ctx`: Context for cancellation/timeout
- `limit`: Maximum number of snapshots to return (0 = no limit)

**Returns**: Slice of recent snapshots, or error

**Example**:
```go
snapshots, err := client.QueryRecent(ctx, 10)
```

#### `QueryByTimeRange(ctx context.Context, startTime, endTime time.Time, limit int) ([]*types.Snapshot, error)`

Convenience method to query snapshots within a time range.

**Parameters**:
- `ctx`: Context for cancellation/timeout
- `startTime`: Start of time range (inclusive)
- `endTime`: End of time range (inclusive)
- `limit`: Maximum number of snapshots to return (0 = no limit)

**Returns**: Slice of matching snapshots, or error

**Example**:
```go
start := time.Now().Add(-1 * time.Hour)
end := time.Now()
snapshots, err := client.QueryByTimeRange(ctx, start, end, 100)
```

#### `Close() error`

Closes the storage backend and releases resources. Always call this when done.

**Returns**: Error if cleanup fails

**Example**:
```go
defer client.Close()
```

### Storage Interface

All storage backends implement the `storage.Storage` interface:

```go
type Storage interface {
    Store(ctx context.Context, snapshot *types.Snapshot) error
    StoreBatch(ctx context.Context, snapshots []*types.Snapshot) error
    Query(ctx context.Context, opts *QueryOptions) ([]*types.Snapshot, error)
    Close() error
}
```

### Query Options

```go
type QueryOptions struct {
    StartTime *time.Time  // Filter from this time (inclusive)
    EndTime   *time.Time  // Filter until this time (inclusive)
    Limit     int         // Maximum results (0 = no limit)
    OrderBy   OrderBy     // Ordering (OrderByTimeAsc or OrderByTimeDesc)
}
```

### Snapshot Types

#### Snapshot

```go
type Snapshot struct {
    Timestamp  string
    Runtime    *RuntimeInfo
    Memory     *MemoryInfo
    Goroutines *GoroutineInfo
}
```

#### RuntimeInfo

```go
type RuntimeInfo struct {
    GoVersion     string
    NumGoroutines int
    GOMAXPROCS    int
    NumCPU        int
    UptimeSeconds float64
}
```

#### MemoryInfo

```go
type MemoryInfo struct {
    HeapInUseBytes     uint64
    HeapAllocatedBytes uint64
    HeapObjects        uint64
    TotalAllocBytes    uint64
    GCCycles           uint32
    LastGCPauseSeconds float64
    GCCPUFraction      float64
}
```

#### GoroutineInfo

```go
type GoroutineInfo struct {
    TotalCount int
}
```

## Common Patterns

### Pattern 1: Periodic Snapshot Collection

```go
func collectPeriodically(client *sdk.Client, interval time.Duration) {
    ctx := context.Background()
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for range ticker.C {
        if err := client.CollectAndStore(ctx); err != nil {
            log.Printf("Failed to collect snapshot: %v", err)
        }
    }
}

// Usage
go collectPeriodically(client, 5*time.Second)
```

### Pattern 2: Collect and Analyze

```go
snapshot, err := client.CollectSnapshot()
if err != nil {
    log.Fatal(err)
}

// Analyze memory usage
if snapshot.Memory.HeapAllocatedBytes > 100*1024*1024 {
    log.Printf("High memory usage: %d bytes", snapshot.Memory.HeapAllocatedBytes)
}

// Check goroutine count
if snapshot.Goroutines.TotalCount > 1000 {
    log.Printf("High goroutine count: %d", snapshot.Goroutines.TotalCount)
}
```

### Pattern 3: Store and Query Later

```go
// Store snapshots
for i := 0; i < 10; i++ {
    client.CollectAndStore(ctx)
    time.Sleep(1 * time.Second)
}

// Query later
snapshots, err := client.QueryRecent(ctx, 10)
if err != nil {
    log.Fatal(err)
}

// Analyze trends
for _, snapshot := range snapshots {
    fmt.Printf("Time: %s, Goroutines: %d\n", 
        snapshot.Timestamp, 
        snapshot.Goroutines.TotalCount)
}
```

### Pattern 4: Export to JSON

```go
snapshot, err := client.CollectSnapshot()
if err != nil {
    log.Fatal(err)
}

jsonData, err := snapshot.ToJSON()
if err != nil {
    log.Fatal(err)
}

// Send to API, write to file, etc.
os.WriteFile("snapshot.json", jsonData, 0644)
```

### Pattern 5: Load from JSON

```go
data, err := os.ReadFile("snapshot.json")
if err != nil {
    log.Fatal(err)
}

snapshot, err := types.FromJSON(data)
if err != nil {
    log.Fatal(err)
}

// Use the snapshot
fmt.Printf("Goroutines: %d\n", snapshot.Goroutines.TotalCount)
```

## Error Handling

All SDK methods return errors. Always check and handle them:

```go
snapshot, err := client.CollectSnapshot()
if err != nil {
    // Handle error appropriately
    log.Printf("Failed to collect snapshot: %v", err)
    return
}

err = client.Store(ctx, snapshot)
if err != nil {
    log.Printf("Failed to store snapshot: %v", err)
    return
}
```

## Thread Safety

- **Client**: Safe for concurrent use
- **MemoryStorage**: Thread-safe
- **FileStorage**: Thread-safe
- **Custom Storage**: Depends on your implementation

## Performance Considerations

1. **Batch Operations**: Use `StoreBatch` for multiple snapshots
2. **Context Timeouts**: Use context timeouts for long-running operations
3. **Storage Choice**: Memory storage is fastest, file storage is persistent
4. **Query Limits**: Always set reasonable limits for queries

## Integration Examples

### Example 1: HTTP Handler

```go
func snapshotHandler(client *sdk.Client) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        snapshot, err := client.CollectSnapshot()
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        jsonData, err := snapshot.ToJSON()
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        w.Write(jsonData)
    }
}
```

### Example 2: Background Monitor

```go
func startMonitor(client *sdk.Client) {
    go func() {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()

        for range ticker.C {
            snapshot, err := client.CollectSnapshot()
            if err != nil {
                continue
            }

            // Check thresholds
            if snapshot.Memory.HeapAllocatedBytes > threshold {
                alert("High memory usage detected")
            }

            // Store for later analysis
            client.Store(context.Background(), snapshot)
        }
    }()
}
```

### Example 3: Database Storage

```go
type DatabaseStorage struct {
    db *sql.DB
}

func (d *DatabaseStorage) Store(ctx context.Context, snapshot *types.Snapshot) error {
    jsonData, err := snapshot.ToJSON()
    if err != nil {
        return err
    }

    _, err = d.db.ExecContext(ctx,
        "INSERT INTO snapshots (timestamp, data) VALUES ($1, $2)",
        snapshot.Timestamp, jsonData)
    return err
}

// Implement other methods...
```

## Best Practices

1. **Always Close**: Use `defer client.Close()` to ensure cleanup
2. **Context Usage**: Pass context for cancellation and timeouts
3. **Error Handling**: Always check and handle errors
4. **Batch Operations**: Use batch methods when storing multiple snapshots
5. **Query Limits**: Set reasonable limits to avoid memory issues
6. **Storage Selection**: Choose storage backend based on your needs

## Troubleshooting

### Issue: "Failed to create base directory"

**Solution**: Ensure you have write permissions for the file storage directory.

### Issue: "Invalid timestamp"

**Solution**: Ensure snapshots are created using the SDK's `CollectSnapshot()` method.

### Issue: Memory storage data lost

**Solution**: This is expected behavior. Use `FileStorage` for persistence.

### Issue: Query returns no results

**Solution**: Check your time range filters and ensure snapshots were stored successfully.

## Additional Resources

- See `examples/` directory for complete working examples
- See `PRD.md` for detailed product requirements
- See `README.md` for CLI tool usage

## Support

For issues, questions, or contributions, please refer to the project repository.

