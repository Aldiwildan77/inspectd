# inspectd SDK

The inspectd SDK provides a programmatic interface for collecting and storing Go runtime snapshots.

## Quick Start

```go
import (
    "context"
    "github.com/Aldiwildan77/inspectd/pkg/sdk"
    "github.com/Aldiwildan77/inspectd/pkg/sdk/storage"
)

func main() {
    // Create storage backend
    memStorage := storage.NewMemoryStorage()
    defer memStorage.Close()

    // Create SDK client
    client := sdk.NewClient(memStorage)

    // Collect and store a snapshot
    ctx := context.Background()
    if err := client.CollectAndStore(ctx); err != nil {
        panic(err)
    }

    // Query stored snapshots
    snapshots, err := client.QueryRecent(ctx, 10)
    if err != nil {
        panic(err)
    }
}
```

## Features

- **Easy to Use**: Simple, intuitive API
- **Flexible Storage**: Pluggable storage backends
- **Type-Safe**: Strongly typed Go interfaces
- **Well Documented**: Comprehensive documentation for AI and human consumption
- **Production Ready**: Thread-safe, error handling, context support

## Storage Backends

- **MemoryStorage**: Fast in-memory storage (for testing/caching)
- **FileStorage**: Persistent file-based storage (JSON files)
- **Custom Storage**: Implement your own backend

## Documentation

- **[SDK Documentation](docs/SDK.md)**: Complete API reference and usage guide
- **[Examples](../examples/)**: Working code examples

## Package Structure

```
pkg/
├── sdk/
│   ├── client.go          # Main SDK client
│   ├── examples/           # Usage examples
│   └── storage/            # Storage interfaces and implementations
│       ├── interface.go    # Storage interface definition
│       ├── memory.go       # In-memory storage
│       └── file.go         # File-based storage
└── types/
    └── snapshot.go         # Snapshot data types
```

## License

See the main project LICENSE file.

