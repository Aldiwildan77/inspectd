package storage

import (
	"context"
	"sort"
	"sync"

	"github.com/Aldiwildan77/inspectd/pkg/sdk/types"
)

// MemoryStorage is an in-memory storage implementation.
// Useful for testing, caching, or temporary storage.
// Data is lost when the storage is closed or the process exits.
type MemoryStorage struct {
	mu        sync.RWMutex
	snapshots []*types.Snapshot
}

// NewMemoryStorage creates a new in-memory storage instance.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		snapshots: make([]*types.Snapshot, 0),
	}
}

// Store saves a snapshot to memory.
func (m *MemoryStorage) Store(ctx context.Context, snapshot *types.Snapshot) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a copy to avoid external modifications
	snapshotCopy := *snapshot
	m.snapshots = append(m.snapshots, &snapshotCopy)

	return nil
}

// StoreBatch saves multiple snapshots to memory.
func (m *MemoryStorage) StoreBatch(ctx context.Context, snapshots []*types.Snapshot) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, snapshot := range snapshots {
		snapshotCopy := *snapshot
		m.snapshots = append(m.snapshots, &snapshotCopy)
	}

	return nil
}

// Query retrieves snapshots from memory based on query options.
func (m *MemoryStorage) Query(ctx context.Context, opts *QueryOptions) ([]*types.Snapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if opts == nil {
		opts = &QueryOptions{}
	}

	results := make([]*types.Snapshot, 0)

	for _, snapshot := range m.snapshots {
		// Parse timestamp for filtering
		timestamp, err := snapshot.ParseTimestamp()
		if err != nil {
			continue // Skip invalid timestamps
		}

		// Apply time filters
		if opts.StartTime != nil && timestamp.Before(*opts.StartTime) {
			continue
		}
		if opts.EndTime != nil && timestamp.After(*opts.EndTime) {
			continue
		}

		// Create a copy
		snapshotCopy := *snapshot
		results = append(results, &snapshotCopy)
	}

	// Sort results
	sort.Slice(results, func(i, j int) bool {
		ti, _ := results[i].ParseTimestamp()
		tj, _ := results[j].ParseTimestamp()
		if opts.OrderBy == OrderByTimeAsc {
			return ti.Before(tj)
		}
		return tj.Before(ti) // Default: newest first
	})

	// Apply limit
	if opts.Limit > 0 && len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	return results, nil
}

// Close releases resources (no-op for memory storage).
func (m *MemoryStorage) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.snapshots = nil
	return nil
}

// GetAll returns all stored snapshots (useful for testing and debugging).
func (m *MemoryStorage) GetAll() []*types.Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make([]*types.Snapshot, len(m.snapshots))
	for i, snapshot := range m.snapshots {
		snapshotCopy := *snapshot
		results[i] = &snapshotCopy
	}
	return results
}

// Count returns the number of stored snapshots.
func (m *MemoryStorage) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.snapshots)
}

