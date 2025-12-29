# MCP (Model Context Protocol) Integration

inspectd can be used as an MCP server, allowing AI assistants to interact with Go runtime inspection capabilities.

## Overview

The MCP server exposes inspectd's functionality through the Model Context Protocol, enabling AI clients to:
- Collect runtime snapshots
- Query stored snapshots
- Access runtime, memory, and goroutine information
- Store snapshots for historical analysis

## Installation

The MCP server is included in the inspectd repository:

```bash
go build -o inspectd-mcp ./cmd/inspectd-mcp
```

## Usage

### Running the MCP Server

The MCP server communicates via stdin/stdout using JSON-RPC 2.0:

```bash
./inspectd-mcp
```

### MCP Client Configuration

Configure your MCP client (e.g., Claude Desktop, Cursor) to use inspectd-mcp:

**Claude Desktop** (`claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "inspectd": {
      "command": "/path/to/inspectd-mcp",
      "args": []
    }
  }
}
```

**Cursor** (`.cursor/mcp.json`):
```json
{
  "mcpServers": {
    "inspectd": {
      "command": "/path/to/inspectd-mcp"
    }
  }
}
```

## Available Tools

### 1. `collect_snapshot`

Collects a runtime snapshot from the current Go process.

**Parameters**: None

**Returns**: Complete snapshot object with runtime, memory, and goroutine information

**Example**:
```json
{
  "method": "tools/call",
  "params": {
    "name": "collect_snapshot",
    "arguments": {}
  }
}
```

### 2. `store_snapshot`

Collects and stores a snapshot to the storage backend.

**Parameters**: None

**Returns**: Status confirmation

**Example**:
```json
{
  "method": "tools/call",
  "params": {
    "name": "store_snapshot",
    "arguments": {}
  }
}
```

### 3. `query_snapshots`

Queries stored snapshots with optional filters.

**Parameters**:
- `limit` (integer, optional): Maximum number of snapshots to return (default: 10)

**Returns**: Array of snapshots

**Example**:
```json
{
  "method": "tools/call",
  "params": {
    "name": "query_snapshots",
    "arguments": {
      "limit": 5
    }
  }
}
```

### 4. `get_runtime_info`

Gets Go runtime information (version, goroutines, CPU, uptime).

**Parameters**: None

**Returns**: Runtime information object

**Example**:
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_runtime_info",
    "arguments": {}
  }
}
```

### 5. `get_memory_info`

Gets memory usage and GC statistics.

**Parameters**: None

**Returns**: Memory information object

**Example**:
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_memory_info",
    "arguments": {}
  }
}
```

### 6. `get_goroutine_count`

Gets the current goroutine count.

**Parameters**: None

**Returns**: Goroutine information object

**Example**:
```json
{
  "method": "tools/call",
  "params": {
    "name": "get_goroutine_count",
    "arguments": {}
  }
}
```

## Available Resources

### 1. `inspectd://snapshot/latest`

The most recent runtime snapshot.

**Example**:
```json
{
  "method": "resources/read",
  "params": {
    "uri": "inspectd://snapshot/latest"
  }
}
```

### 2. `inspectd://snapshots/recent`

Recently collected snapshots (up to 10).

**Example**:
```json
{
  "method": "resources/read",
  "params": {
    "uri": "inspectd://snapshots/recent"
  }
}
```

## Protocol Details

The MCP server implements JSON-RPC 2.0 over stdin/stdout:

- **Protocol Version**: 2024-11-05
- **Transport**: stdio (stdin/stdout)
- **Format**: JSON

### Request Format

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "collect_snapshot",
    "arguments": {}
  }
}
```

### Response Format

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "timestamp": "2024-01-01T12:00:00.123456789Z",
    "runtime": {...},
    "memory": {...},
    "goroutines": {...}
  }
}
```

## Use Cases

1. **AI-Assisted Debugging**: AI assistants can query runtime state to help debug issues
2. **Performance Analysis**: Collect and analyze runtime metrics over time
3. **Monitoring**: Track goroutine counts, memory usage, and GC statistics
4. **Development**: Quick runtime checks during development

## Limitations

- The MCP server uses in-memory storage (bounded to 1000 snapshots)
- Data is lost when the server process exits
- For persistent storage, modify the server to use file or database storage

## Customization

To use persistent storage, modify `cmd/inspectd-mcp/main.go`:

```go
// Use file storage instead
fileStorage, err := storage.NewManagedFileStorage("./snapshots", storage.ManagedFileStorageConfig{
    MaxFiles: 1000,
    MaxAge:   7 * 24 * time.Hour,
})
client := sdk.NewClient(fileStorage)
```

## Testing

Test the MCP server using the MCP Inspector:

```bash
npx @modelcontextprotocol/inspector
```

Then connect to your inspectd-mcp server to test tools and resources.

## See Also

- [SDK Documentation](SDK.md) - For programmatic SDK usage
- [Production Guide](PRODUCTION.md) - For production deployment
- [MCP Specification](https://modelcontextprotocol.io) - Official MCP documentation

