package sdk

import (
	"context"
	"time"

	"github.com/Aldiwildan77/inspectd/internal/goroutines"
	"github.com/Aldiwildan77/inspectd/internal/memory"
	"github.com/Aldiwildan77/inspectd/internal/runtimeinfo"
	"github.com/Aldiwildan77/inspectd/pkg/sdk/storage"
	"github.com/Aldiwildan77/inspectd/pkg/sdk/types"
)

// Client provides a high-level API for collecting and storing inspectd snapshots.
// This is the main entry point for using the inspectd SDK.
type Client struct {
	storage storage.Storage
}

// NewClient creates a new SDK client with the provided storage backend.
// The storage can be any implementation of the storage.Storage interface.
func NewClient(storage storage.Storage) *Client {
	return &Client{
		storage: storage,
	}
}

// CollectSnapshot collects a new runtime snapshot from the current process.
// Returns a Snapshot object containing runtime, memory, and goroutine information.
func (c *Client) CollectSnapshot() (*types.Snapshot, error) {
	// Collect runtime information
	runtimeInfo, err := runtimeinfo.Collect()
	if err != nil {
		return nil, err
	}

	// Collect memory information
	memInfo, err := memory.Collect()
	if err != nil {
		return nil, err
	}

	// Collect goroutine information
	goroutineInfo, err := goroutines.Collect()
	if err != nil {
		return nil, err
	}

	// Convert internal types to SDK types
	snapshot := &types.Snapshot{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Runtime: &types.RuntimeInfo{
			GoVersion:     runtimeInfo.GoVersion,
			NumGoroutines: runtimeInfo.NumGoroutines,
			GOMAXPROCS:    runtimeInfo.GOMAXPROCS,
			NumCPU:        runtimeInfo.NumCPU,
			UptimeSeconds: runtimeInfo.Uptime,
		},
		Memory: &types.MemoryInfo{
			HeapInUseBytes:     memInfo.HeapInUse,
			HeapAllocatedBytes: memInfo.HeapAllocated,
			HeapObjects:        memInfo.HeapObjects,
			TotalAllocBytes:    memInfo.TotalAlloc,
			GCCycles:           memInfo.GCCycles,
			LastGCPauseSeconds: memInfo.LastGCPause,
			GCCPUFraction:      memInfo.GCCPUFraction,
		},
		Goroutines: &types.GoroutineInfo{
			TotalCount: goroutineInfo.TotalCount,
		},
	}

	return snapshot, nil
}

// CollectAndStore collects a snapshot and stores it in the configured storage backend.
// This is a convenience method that combines CollectSnapshot and Store.
func (c *Client) CollectAndStore(ctx context.Context) error {
	snapshot, err := c.CollectSnapshot()
	if err != nil {
		return err
	}
	return c.Store(ctx, snapshot)
}

// Store saves a snapshot to the configured storage backend.
func (c *Client) Store(ctx context.Context, snapshot *types.Snapshot) error {
	return c.storage.Store(ctx, snapshot)
}

// StoreBatch saves multiple snapshots to the storage backend.
// This is more efficient than multiple Store calls for bulk operations.
func (c *Client) StoreBatch(ctx context.Context, snapshots []*types.Snapshot) error {
	return c.storage.StoreBatch(ctx, snapshots)
}

// Query retrieves snapshots from the storage backend based on query options.
func (c *Client) Query(ctx context.Context, opts *storage.QueryOptions) ([]*types.Snapshot, error) {
	return c.storage.Query(ctx, opts)
}

// QueryByTimeRange retrieves snapshots within a time range.
// This is a convenience method for common time-based queries.
func (c *Client) QueryByTimeRange(ctx context.Context, startTime, endTime time.Time, limit int) ([]*types.Snapshot, error) {
	opts := &storage.QueryOptions{
		StartTime: &startTime,
		EndTime:   &endTime,
		Limit:     limit,
		OrderBy:   storage.OrderByTimeDesc,
	}
	return c.storage.Query(ctx, opts)
}

// QueryRecent retrieves the most recent snapshots.
// This is a convenience method for getting the latest data.
func (c *Client) QueryRecent(ctx context.Context, limit int) ([]*types.Snapshot, error) {
	opts := &storage.QueryOptions{
		Limit:   limit,
		OrderBy: storage.OrderByTimeDesc,
	}
	return c.storage.Query(ctx, opts)
}

// Close closes the storage backend and releases resources.
// Should be called when the client is no longer needed.
func (c *Client) Close() error {
	return c.storage.Close()
}
