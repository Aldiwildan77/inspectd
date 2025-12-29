package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Aldiwildan77/inspectd/sdk"
	"github.com/Aldiwildan77/inspectd/sdk/storage"
)

// MCPRequest represents an MCP request
type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// MCPResponse represents an MCP response
type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an MCP error
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCPTool represents an MCP tool definition
type MCPTool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

// MCPResource represents an MCP resource
type MCPResource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType,omitempty"`
}

// MCPServer handles MCP protocol communication
type MCPServer struct {
	client *sdk.Client
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer() (*MCPServer, error) {
	// Use bounded memory storage for MCP server
	memStorage := storage.NewBoundedMemoryStorage(1000)
	client := sdk.NewClient(memStorage)

	return &MCPServer{
		client: client,
	}, nil
}

// HandleRequest processes an MCP request
func (s *MCPServer) HandleRequest(req MCPRequest) MCPResponse {
	resp := MCPResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}

	switch req.Method {
	case "initialize":
		resp.Result = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
				"resources": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "inspectd-mcp",
				"version": "1.1.0",
			},
		}

	case "tools/list":
		resp.Result = map[string]interface{}{
			"tools": s.listTools(),
		}

	case "tools/call":
		var params struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &MCPError{Code: -32602, Message: "Invalid params"}
			break
		}

		result, err := s.callTool(params.Name, params.Arguments)
		if err != nil {
			resp.Error = &MCPError{Code: -32603, Message: err.Error()}
		} else {
			resp.Result = result
		}

	case "resources/list":
		resp.Result = map[string]interface{}{
			"resources": s.listResources(),
		}

	case "resources/read":
		var params struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			resp.Error = &MCPError{Code: -32602, Message: "Invalid params"}
			break
		}

		result, err := s.readResource(params.URI)
		if err != nil {
			resp.Error = &MCPError{Code: -32603, Message: err.Error()}
		} else {
			resp.Result = result
		}

	default:
		resp.Error = &MCPError{Code: -32601, Message: "Method not found"}
	}

	return resp
}

// listTools returns available MCP tools
func (s *MCPServer) listTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "collect_snapshot",
			Description: "Collect a runtime snapshot from the current Go process",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "store_snapshot",
			Description: "Store a collected snapshot to the storage backend",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "query_snapshots",
			Description: "Query stored snapshots with optional filters",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of snapshots to return",
					},
				},
			},
		},
		{
			Name:        "get_runtime_info",
			Description: "Get Go runtime information (version, goroutines, CPU, uptime)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "get_memory_info",
			Description: "Get memory usage and GC statistics",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "get_goroutine_count",
			Description: "Get the current goroutine count",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

// callTool executes an MCP tool
func (s *MCPServer) callTool(name string, args map[string]interface{}) (interface{}, error) {
	ctx := context.Background()

	switch name {
	case "collect_snapshot":
		snapshot, err := s.client.CollectSnapshot()
		if err != nil {
			return nil, err
		}
		return snapshot, nil

	case "store_snapshot":
		if err := s.client.CollectAndStore(ctx); err != nil {
			return nil, err
		}
		return map[string]interface{}{"status": "stored"}, nil

	case "query_snapshots":
		limit := 10
		if l, ok := args["limit"].(float64); ok {
			limit = int(l)
		}
		snapshots, err := s.client.QueryRecent(ctx, limit)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"snapshots": snapshots}, nil

	case "get_runtime_info":
		snapshot, err := s.client.CollectSnapshot()
		if err != nil {
			return nil, err
		}
		return snapshot.Runtime, nil

	case "get_memory_info":
		snapshot, err := s.client.CollectSnapshot()
		if err != nil {
			return nil, err
		}
		return snapshot.Memory, nil

	case "get_goroutine_count":
		snapshot, err := s.client.CollectSnapshot()
		if err != nil {
			return nil, err
		}
		return snapshot.Goroutines, nil

	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

// listResources returns available MCP resources
func (s *MCPServer) listResources() []MCPResource {
	return []MCPResource{
		{
			URI:         "inspectd://snapshot/latest",
			Name:        "Latest Snapshot",
			Description: "The most recent runtime snapshot",
			MimeType:    "application/json",
		},
		{
			URI:         "inspectd://snapshots/recent",
			Name:        "Recent Snapshots",
			Description: "Recently collected snapshots",
			MimeType:    "application/json",
		},
	}
}

// readResource reads an MCP resource
func (s *MCPServer) readResource(uri string) (interface{}, error) {
	ctx := context.Background()

	switch uri {
	case "inspectd://snapshot/latest":
		snapshots, err := s.client.QueryRecent(ctx, 1)
		if err != nil {
			return nil, err
		}
		if len(snapshots) == 0 {
			return nil, fmt.Errorf("no snapshots available")
		}
		return map[string]interface{}{
			"contents": []map[string]interface{}{
				{
					"uri":      uri,
					"mimeType": "application/json",
					"text":     mustJSON(snapshots[0]),
				},
			},
		}, nil

	case "inspectd://snapshots/recent":
		snapshots, err := s.client.QueryRecent(ctx, 10)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"contents": []map[string]interface{}{
				{
					"uri":      uri,
					"mimeType": "application/json",
					"text":     mustJSON(snapshots),
				},
			},
		}, nil

	default:
		return nil, fmt.Errorf("unknown resource: %s", uri)
	}
}

// mustJSON marshals a value to JSON string, panics on error
func mustJSON(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(data)
}

func main() {
	server, err := NewMCPServer()
	if err != nil {
		log.Fatal(err)
	}
	defer server.client.Close()

	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		var req MCPRequest
		if err := decoder.Decode(&req); err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Printf("Error decoding request: %v", err)
			continue
		}

		resp := server.HandleRequest(req)
		if err := encoder.Encode(resp); err != nil {
			log.Printf("Error encoding response: %v", err)
			continue
		}
	}
}

