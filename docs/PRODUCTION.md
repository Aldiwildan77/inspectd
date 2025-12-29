# Production Deployment Guide

This guide covers best practices for deploying inspectd SDK in production container/pod environments.

## Table of Contents

1. [Storage Backend Selection](#storage-backend-selection)
2. [Production-Ready Storage Options](#production-ready-storage-options)
3. [Kubernetes/Pod Configuration](#kubernetespod-configuration)
4. [Resource Management](#resource-management)
5. [Error Handling](#error-handling)
6. [Monitoring and Observability](#monitoring-and-observability)
7. [Best Practices](#best-practices)

## Storage Backend Selection

### When to Use Each Backend

| Backend | Use Case | Pros | Cons |
|---------|----------|------|------|
| **BoundedMemoryStorage** | Caching, short-term storage | Fast, no I/O | Data lost on restart, limited capacity |
| **ManagedFileStorage** | Local persistence, debugging | Simple, persistent | Disk space management needed |
| **DatabaseStorage** | Production, analytics | Scalable, queryable | Requires database infrastructure |
| **CloudObjectStorage** | Cloud-native, long-term | Scalable, durable | Network latency, costs |

## Production-Ready Storage Options

### 1. BoundedMemoryStorage

**Use for**: Caching recent snapshots, temporary storage

```go
import (
    "github.com/Aldiwildan77/inspectd/pkg/sdk"
    "github.com/Aldiwildan77/inspectd/pkg/sdk/storage"
)

// Create bounded storage with 1000 snapshot limit
memStorage := storage.NewBoundedMemoryStorage(1000)
defer memStorage.Close()

client := sdk.NewClient(memStorage)
```

**Features**:
- Automatic eviction of oldest snapshots
- Thread-safe
- No disk I/O
- Memory usage bounded

**Limitations**:
- Data lost on pod restart
- Limited by available memory

### 2. ManagedFileStorage

**Use for**: Local file-based persistence with automatic cleanup

```go
fileStorage, err := storage.NewManagedFileStorage("/var/lib/inspectd/snapshots", storage.ManagedFileStorageConfig{
    MaxFiles:       1000,              // Maximum number of files
    MaxAge:         7 * 24 * time.Hour, // Retain for 7 days
    CleanupInterval: 1 * time.Hour,     // Cleanup every hour
})
defer fileStorage.Close()

client := sdk.NewClient(fileStorage)
```

**Features**:
- Automatic cleanup based on age and count
- Persistent across restarts (with persistent volumes)
- Background cleanup goroutine
- Disk space management

**Kubernetes Setup**:
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: inspectd-snapshots
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
---
volumeMounts:
  - name: snapshots
    mountPath: /var/lib/inspectd/snapshots
volumes:
  - name: snapshots
    persistentVolumeClaim:
      claimName: inspectd-snapshots
```

### 3. DatabaseStorage

**Use for**: Production environments requiring queryability and scalability

```go
import (
    "database/sql"
    _ "github.com/lib/pq" // PostgreSQL driver
)

dbStorage, err := storage.NewDatabaseStorage(storage.DatabaseStorageConfig{
    Driver:        "postgres",
    DSN:           os.Getenv("DATABASE_URL"),
    TableName:     "inspectd_snapshots",
    MaxConnections: 10,
})
defer dbStorage.Close()

client := sdk.NewClient(dbStorage)
```

**Features**:
- Scalable storage
- SQL query support
- Transaction support for batch operations
- Automatic table creation

**Database Schema** (auto-created):
```sql
CREATE TABLE inspectd_snapshots (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL,
    data JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_inspectd_snapshots_timestamp ON inspectd_snapshots(timestamp);
```

### 4. CloudObjectStorage

**Use for**: Cloud-native deployments, long-term storage

```go
// Example with S3-compatible storage
type S3Client struct {
    // Your S3 client implementation
}

func (s *S3Client) PutObject(ctx context.Context, bucket, key string, data []byte) error {
    // Implement S3 upload
}

func (s *S3Client) GetObject(ctx context.Context, bucket, key string) ([]byte, error) {
    // Implement S3 download
}

func (s *S3Client) ListObjects(ctx context.Context, bucket, prefix string) ([]string, error) {
    // Implement S3 list
}

func (s *S3Client) DeleteObject(ctx context.Context, bucket, key string) error {
    // Implement S3 delete
}

s3Client := &S3Client{}
objStorage, err := storage.NewCloudObjectStorage(storage.CloudObjectStorageConfig{
    Client:        s3Client,
    Bucket:        "my-snapshots-bucket",
    Prefix:        "snapshots/",
    MaxAge:        30 * 24 * time.Hour, // 30 days
    CleanupInterval: 1 * time.Hour,
})
defer objStorage.Close()

client := sdk.NewClient(objStorage)
```

**Features**:
- Scalable and durable
- Automatic cleanup
- Cloud-native
- Cost-effective for long-term storage

## Kubernetes/Pod Configuration

### Resource Limits

Always set resource limits to prevent OOM kills:

```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "100m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

### Graceful Shutdown

Handle pod termination gracefully:

```go
func main() {
    // Setup signal handling
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    // Create client
    client := sdk.NewClient(storage)
    defer client.Close()

    // Graceful shutdown
    go func() {
        <-sigChan
        log.Println("Shutting down...")
        client.Close()
        os.Exit(0)
    }()

    // Your application logic
}
```

### Health Checks

Add health check endpoints:

```go
func healthHandler(client *sdk.Client) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
        defer cancel()

        if err := client.CollectSnapshot(); err != nil {
            http.Error(w, "unhealthy", http.StatusServiceUnavailable)
            return
        }

        w.WriteHeader(http.StatusOK)
        w.Write([]byte("healthy"))
    }
}
```

## Resource Management

### Memory Storage Limits

Always set limits for memory storage:

```go
// ❌ BAD: Unbounded growth
memStorage := storage.NewMemoryStorage()

// ✅ GOOD: Bounded with limits
memStorage := storage.NewBoundedMemoryStorage(1000)
```

### File Storage Cleanup

Always configure cleanup for file storage:

```go
// ❌ BAD: No cleanup, disk will fill up
fileStorage, _ := storage.NewFileStorage("./snapshots")

// ✅ GOOD: Automatic cleanup
fileStorage, _ := storage.NewManagedFileStorage("./snapshots", storage.ManagedFileStorageConfig{
    MaxFiles: 1000,
    MaxAge:   7 * 24 * time.Hour,
})
```

### Context Timeouts

Always use context timeouts:

```go
// ❌ BAD: No timeout
err := client.CollectAndStore(context.Background())

// ✅ GOOD: With timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
err := client.CollectAndStore(ctx)
```

## Error Handling

### Always Check Errors

```go
snapshot, err := client.CollectSnapshot()
if err != nil {
    log.Printf("Failed to collect snapshot: %v", err)
    // Handle error appropriately
    return
}

err = client.Store(ctx, snapshot)
if err != nil {
    log.Printf("Failed to store snapshot: %v", err)
    // Handle error appropriately
    return
}
```

### Handle Context Cancellation

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := client.CollectAndStore(ctx)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        log.Println("Operation timed out")
    } else if errors.Is(err, context.Canceled) {
        log.Println("Operation canceled")
    } else {
        log.Printf("Operation failed: %v", err)
    }
}
```

## Monitoring and Observability

### Metrics to Monitor

1. **Snapshot Collection Rate**: Track successful/failed collections
2. **Storage Operations**: Track store/query success rates
3. **Storage Size**: Monitor memory usage or disk space
4. **Operation Latency**: Track collection and storage times

### Example Metrics Collection

```go
var (
    snapshotCount = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "inspectd_snapshots_collected_total",
            Help: "Total number of snapshots collected",
        },
        []string{"status"},
    )
    storageLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "inspectd_storage_operation_duration_seconds",
            Help: "Storage operation duration",
        },
        []string{"operation"},
    )
)

func collectWithMetrics(client *sdk.Client) {
    start := time.Now()
    err := client.CollectAndStore(context.Background())
    duration := time.Since(start)

    storageLatency.WithLabelValues("store").Observe(duration.Seconds())

    if err != nil {
        snapshotCount.WithLabelValues("error").Inc()
    } else {
        snapshotCount.WithLabelValues("success").Inc()
    }
}
```

## Best Practices

### ✅ DO

1. **Use BoundedMemoryStorage** for caching with size limits
2. **Use ManagedFileStorage** with cleanup policies for file storage
3. **Use DatabaseStorage** or CloudObjectStorage for production
4. **Set resource limits** in Kubernetes
5. **Use context timeouts** for all operations
6. **Handle graceful shutdown** properly
7. **Monitor storage usage** and errors
8. **Use persistent volumes** for file storage
9. **Implement health checks**
10. **Log errors** appropriately

### ❌ DON'T

1. **Don't use unbounded MemoryStorage** in production
2. **Don't use FileStorage without cleanup** policies
3. **Don't ignore context cancellation**
4. **Don't store snapshots in ephemeral storage** without limits
5. **Don't skip error handling**
6. **Don't forget to close** storage backends
7. **Don't use blocking operations** without timeouts
8. **Don't ignore resource limits** in containers

## Example Production Deployment

See `pkg/sdk/examples/production/` for complete production-ready examples.

## Troubleshooting

### Issue: OOM Kills

**Solution**: Use `BoundedMemoryStorage` with appropriate limits, or use external storage.

### Issue: Disk Full

**Solution**: Use `ManagedFileStorage` with cleanup policies, or use object storage.

### Issue: Slow Operations

**Solution**: Use context timeouts, monitor latency, consider async operations.

### Issue: Data Loss on Restart

**Solution**: Use persistent storage (file with volumes, database, or object storage).

## Additional Resources

- [SDK Documentation](SDK.md)
- [Examples](../pkg/sdk/examples/)
- [PRD](PRD.md)

