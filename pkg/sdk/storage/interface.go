package storage

import (
	"context"
	"time"

	"github.com/Aldiwildan77/inspectd/pkg/sdk/types"
)

// Storage defines the interface for storing inspectd snapshots.
// Implementations can store data to any backend (file, database, API, etc.).
type Storage interface {
	// Store saves a snapshot to the storage backend.
	// Returns an error if the operation fails.
	Store(ctx context.Context, snapshot *types.Snapshot) error

	// StoreBatch saves multiple snapshots in a single operation.
	// This is useful for bulk operations and can be more efficient than multiple Store calls.
	// Returns an error if the operation fails.
	StoreBatch(ctx context.Context, snapshots []*types.Snapshot) error

	// Query retrieves snapshots based on the provided query options.
	// Returns a slice of snapshots matching the query criteria.
	// Returns an error if the operation fails.
	Query(ctx context.Context, opts *QueryOptions) ([]*types.Snapshot, error)

	// Close releases any resources held by the storage backend.
	// Should be called when the storage is no longer needed.
	Close() error
}

// QueryOptions defines parameters for querying stored snapshots.
type QueryOptions struct {
	// StartTime filters snapshots from this time onwards (inclusive).
	// If nil, no start time filter is applied.
	StartTime *time.Time

	// EndTime filters snapshots up to this time (inclusive).
	// If nil, no end time filter is applied.
	EndTime *time.Time

	// Limit restricts the maximum number of snapshots to return.
	// If 0, no limit is applied.
	Limit int

	// OrderBy specifies how results should be ordered.
	// Default is OrderByTimeDesc (newest first).
	OrderBy OrderBy
}

// OrderBy specifies the ordering of query results.
type OrderBy int

const (
	// OrderByTimeAsc orders by timestamp ascending (oldest first).
	OrderByTimeAsc OrderBy = iota
	// OrderByTimeDesc orders by timestamp descending (newest first).
	OrderByTimeDesc
)

