package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/Aldiwildan77/inspectd/pkg/sdk/types"
)

// FileStorage stores snapshots as individual JSON files in a directory.
// Each snapshot is stored as a file named with its timestamp.
// Useful for simple file-based persistence and debugging.
type FileStorage struct {
	mu       sync.RWMutex
	baseDir  string
	fileMode os.FileMode
}

// NewFileStorage creates a new file storage instance.
// The baseDir will be created if it doesn't exist.
func NewFileStorage(baseDir string) (*FileStorage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &FileStorage{
		baseDir:  baseDir,
		fileMode: 0644,
	}, nil
}

// Store saves a snapshot as a JSON file.
func (f *FileStorage) Store(ctx context.Context, snapshot *types.Snapshot) error {
	// Parse timestamp to create filename
	timestamp, err := snapshot.ParseTimestamp()
	if err != nil {
		return fmt.Errorf("invalid timestamp: %w", err)
	}

	// Create filename from timestamp (sanitized for filesystem)
	filename := timestamp.Format("2006-01-02T15-04-05.000000000Z") + ".json"
	filePath := filepath.Join(f.baseDir, filename)

	// Marshal to JSON
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	// Write to file
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := os.WriteFile(filePath, data, f.fileMode); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// StoreBatch saves multiple snapshots as individual files.
func (f *FileStorage) StoreBatch(ctx context.Context, snapshots []*types.Snapshot) error {
	for _, snapshot := range snapshots {
		if err := f.Store(ctx, snapshot); err != nil {
			return err
		}
	}
	return nil
}

// Query retrieves snapshots by reading files from the directory.
func (f *FileStorage) Query(ctx context.Context, opts *QueryOptions) ([]*types.Snapshot, error) {
	if opts == nil {
		opts = &QueryOptions{}
	}

	f.mu.RLock()
	defer f.mu.RUnlock()

	// Read all files in directory
	entries, err := os.ReadDir(f.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	results := make([]*types.Snapshot, 0)

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		filePath := filepath.Join(f.baseDir, entry.Name())

		// Read and parse JSON
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue // Skip files that can't be read
		}

		snapshot, err := types.FromJSON(data)
		if err != nil {
			continue // Skip invalid JSON
		}

		// Apply filters
		timestamp, err := snapshot.ParseTimestamp()
		if err != nil {
			continue
		}

		if opts.StartTime != nil && timestamp.Before(*opts.StartTime) {
			continue
		}
		if opts.EndTime != nil && timestamp.After(*opts.EndTime) {
			continue
		}

		results = append(results, snapshot)
	}

	// Sort results
	sortSnapshots(results, opts.OrderBy)

	// Apply limit
	if opts.Limit > 0 && len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	return results, nil
}

// Close releases resources (no-op for file storage).
func (f *FileStorage) Close() error {
	return nil
}

// sortSnapshots sorts snapshots by timestamp.
func sortSnapshots(snapshots []*types.Snapshot, orderBy OrderBy) {
	sort.Slice(snapshots, func(i, j int) bool {
		ti, _ := snapshots[i].ParseTimestamp()
		tj, _ := snapshots[j].ParseTimestamp()
		if orderBy == OrderByTimeAsc {
			return ti.Before(tj)
		}
		return tj.Before(ti) // Default: newest first
	})
}

