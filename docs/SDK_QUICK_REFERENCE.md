# SDK Quick Reference for AI Integration

This document provides a concise reference for AI agents integrating the inspectd SDK.

## Import Paths

```go
import "github.com/Aldiwildan77/inspectd/pkg/sdk"
import "github.com/Aldiwildan77/inspectd/pkg/sdk/storage"
import "github.com/Aldiwildan77/inspectd/pkg/sdk/types"
```

## Minimal Working Example

```go
// 1. Create storage
storage := storage.NewMemoryStorage()
defer storage.Close()

// 2. Create client
client := sdk.NewClient(storage)

// 3. Collect and store
ctx := context.Background()
client.CollectAndStore(ctx)

// 4. Query
snapshots, _ := client.QueryRecent(ctx, 10)
```

## Key Types

### Snapshot
```go
type Snapshot struct {
    Timestamp  string
    Runtime    *RuntimeInfo
    Memory     *MemoryInfo
    Goroutines *GoroutineInfo
}
```

### Storage Interface
```go
type Storage interface {
    Store(ctx context.Context, snapshot *types.Snapshot) error
    StoreBatch(ctx context.Context, snapshots []*types.Snapshot) error
    Query(ctx context.Context, opts *QueryOptions) ([]*types.Snapshot, error)
    Close() error
}
```

## Available Storage Implementations

1. **MemoryStorage**: `storage.NewMemoryStorage()`
2. **FileStorage**: `storage.NewFileStorage(dir string)`
3. **Custom**: Implement `storage.Storage` interface

## Client Methods

- `CollectSnapshot() (*types.Snapshot, error)`
- `CollectAndStore(ctx context.Context) error`
- `Store(ctx context.Context, snapshot *types.Snapshot) error`
- `StoreBatch(ctx context.Context, snapshots []*types.Snapshot) error`
- `Query(ctx context.Context, opts *storage.QueryOptions) ([]*types.Snapshot, error)`
- `QueryRecent(ctx context.Context, limit int) ([]*types.Snapshot, error)`
- `QueryByTimeRange(ctx context.Context, startTime, endTime time.Time, limit int) ([]*types.Snapshot, error)`
- `Close() error`

## Common Patterns

### Pattern 1: Single Snapshot
```go
client.CollectAndStore(ctx)
```

### Pattern 2: Multiple Snapshots
```go
snapshots := []*types.Snapshot{}
for i := 0; i < 10; i++ {
    s, _ := client.CollectSnapshot()
    snapshots = append(snapshots, s)
}
client.StoreBatch(ctx, snapshots)
```

### Pattern 3: Query Recent
```go
snapshots, _ := client.QueryRecent(ctx, 10)
```

### Pattern 4: Query Time Range
```go
start := time.Now().Add(-1 * time.Hour)
end := time.Now()
snapshots, _ := client.QueryByTimeRange(ctx, start, end, 100)
```

## Error Handling

All methods return errors. Always check:
```go
if err != nil {
    // handle error
}
```

## Thread Safety

- Client: ✅ Thread-safe
- MemoryStorage: ✅ Thread-safe
- FileStorage: ✅ Thread-safe

## See Also

- Full documentation: `docs/SDK.md`
- Examples: `pkg/sdk/examples/`

