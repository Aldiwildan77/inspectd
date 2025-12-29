package storage

import (
	"context"
	"sort"
	"sync"

	"github.com/Aldiwildan77/inspectd/pkg/sdk/types"
)

// BoundedMemoryStorage is a production-ready in-memory storage with size limits.
// It automatically evicts oldest snapshots when capacity is reached.
// Suitable for caching recent snapshots in production environments.
type BoundedMemoryStorage struct {
	mu        sync.RWMutex
	snapshots []*types.Snapshot
	maxSize   int
}

// NewBoundedMemoryStorage creates a new bounded memory storage instance.
// maxSize specifies the maximum number of snapshots to retain.
// When maxSize is reached, oldest snapshots are evicted (FIFO).
func NewBoundedMemoryStorage(maxSize int) *BoundedMemoryStorage {
	if maxSize <= 0 {
		maxSize = 1000 // Default limit
	}
	return &BoundedMemoryStorage{
		snapshots: make([]*types.Snapshot, 0, maxSize),
		maxSize:   maxSize,
	}
}

// Store saves a snapshot to memory, evicting oldest if at capacity.
func (m *BoundedMemoryStorage) Store(ctx context.Context, snapshot *types.Snapshot) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Evict oldest if at capacity
	if len(m.snapshots) >= m.maxSize {
		// Remove oldest (first element)
		m.snapshots = m.snapshots[1:]
	}

	// Create a copy to avoid external modifications
	snapshotCopy := *snapshot
	m.snapshots = append(m.snapshots, &snapshotCopy)

	return nil
}

// StoreBatch saves multiple snapshots, evicting as needed.
func (m *BoundedMemoryStorage) StoreBatch(ctx context.Context, snapshots []*types.Snapshot) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Calculate how many we need to evict
	totalAfter := len(m.snapshots) + len(snapshots)
	if totalAfter > m.maxSize {
		evictCount := totalAfter - m.maxSize
		if evictCount > len(m.snapshots) {
			evictCount = len(m.snapshots)
		}
		m.snapshots = m.snapshots[evictCount:]
	}

	// Add new snapshots
	for _, snapshot := range snapshots {
		snapshotCopy := *snapshot
		m.snapshots = append(m.snapshots, &snapshotCopy)
	}

	return nil
}

// Query retrieves snapshots from memory based on query options.
func (m *BoundedMemoryStorage) Query(ctx context.Context, opts *QueryOptions) ([]*types.Snapshot, error) {
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

// Close releases resources.
func (m *BoundedMemoryStorage) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.snapshots = nil
	return nil
}

// Count returns the number of stored snapshots.
func (m *BoundedMemoryStorage) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.snapshots)
}

// MaxSize returns the maximum capacity.
func (m *BoundedMemoryStorage) MaxSize() int {
	return m.maxSize
}

// Clear removes all stored snapshots.
func (m *BoundedMemoryStorage) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.snapshots = m.snapshots[:0]
}

