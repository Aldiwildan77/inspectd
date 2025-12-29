package storage

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/Aldiwildan77/inspectd/sdk/types"
)

// ObjectStorage is an interface for object storage backends (S3, GCS, Azure Blob, etc.).
// This allows the SDK to work with any object storage implementation.
type ObjectStorage interface {
	// PutObject uploads data to object storage.
	PutObject(ctx context.Context, bucket, key string, data []byte) error

	// GetObject retrieves data from object storage.
	GetObject(ctx context.Context, bucket, key string) ([]byte, error)

	// ListObjects lists objects with the given prefix.
	ListObjects(ctx context.Context, bucket, prefix string) ([]string, error)

	// DeleteObject deletes an object.
	DeleteObject(ctx context.Context, bucket, key string) error
}

// CloudObjectStorage stores snapshots in object storage (S3, GCS, Azure Blob, etc.).
// This is a production-ready storage backend for cloud environments.
type CloudObjectStorage struct {
	client     ObjectStorage
	bucket     string
	prefix     string
	maxAge     time.Duration
	cleanupTicker *time.Ticker
	stopCleanup  chan struct{}
	cleanupDone  chan struct{}
}

// CloudObjectStorageConfig configures cloud object storage.
type CloudObjectStorageConfig struct {
	// Client is the object storage client implementation.
	Client ObjectStorage

	// Bucket is the bucket/container name.
	Bucket string

	// Prefix is the key prefix for all snapshots (e.g., "snapshots/").
	Prefix string

	// MaxAge is the maximum age of objects to retain (0 = no age limit).
	MaxAge time.Duration

	// CleanupInterval is how often to run cleanup (default: 1 hour).
	CleanupInterval time.Duration
}

// NewCloudObjectStorage creates a new cloud object storage instance.
// Cleanup runs automatically in the background if MaxAge is set.
func NewCloudObjectStorage(config CloudObjectStorageConfig) (*CloudObjectStorage, error) {
	if config.Client == nil {
		return nil, fmt.Errorf("object storage client is required")
	}
	if config.Bucket == "" {
		return nil, fmt.Errorf("bucket name is required")
	}
	if config.Prefix == "" {
		config.Prefix = "snapshots/"
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 1 * time.Hour
	}

	cos := &CloudObjectStorage{
		client:        config.Client,
		bucket:        config.Bucket,
		prefix:        config.Prefix,
		maxAge:        config.MaxAge,
		cleanupTicker: time.NewTicker(config.CleanupInterval),
		stopCleanup:   make(chan struct{}),
		cleanupDone:   make(chan struct{}),
	}

	// Start cleanup if maxAge is set
	if config.MaxAge > 0 {
		go cos.cleanupLoop()
	}

	return cos, nil
}

// cleanupLoop runs periodic cleanup in the background.
func (c *CloudObjectStorage) cleanupLoop() {
	defer close(c.cleanupDone)

	for {
		select {
		case <-c.cleanupTicker.C:
			if err := c.cleanup(); err != nil {
				// Log error but continue
				_ = err
			}
		case <-c.stopCleanup:
			return
		}
	}
}

// cleanup removes old objects based on retention policies.
func (c *CloudObjectStorage) cleanup() error {
	if c.maxAge == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// List all objects
	keys, err := c.client.ListObjects(ctx, c.bucket, c.prefix)
	if err != nil {
		return fmt.Errorf("failed to list objects: %w", err)
	}

	cutoff := time.Now().Add(-c.maxAge)

	// Delete old objects
	for _, key := range keys {
		// Parse timestamp from key (format: prefix/2006-01-02T15-04-05.000000000Z.json)
		// This is a simplified check - in production, you might want to fetch metadata
		if shouldDelete, err := c.shouldDeleteObject(ctx, key, cutoff); err == nil && shouldDelete {
			if err := c.client.DeleteObject(ctx, c.bucket, key); err != nil {
				// Log but continue
				_ = err
			}
		}
	}

	return nil
}

// shouldDeleteObject checks if an object should be deleted based on age.
func (c *CloudObjectStorage) shouldDeleteObject(ctx context.Context, key string, cutoff time.Time) (bool, error) {
	// Try to parse timestamp from key
	// Key format: prefix/2006-01-02T15-04-05.000000000Z.json
	// Extract timestamp part and parse
	// This is a simplified implementation
	// In production, you might fetch object metadata for accurate timestamp
	
	// For now, we'll fetch the object and parse the snapshot
	data, err := c.client.GetObject(ctx, c.bucket, key)
	if err != nil {
		return false, err
	}

	snapshot, err := types.FromJSON(data)
	if err != nil {
		return false, err
	}

	timestamp, err := snapshot.ParseTimestamp()
	if err != nil {
		return false, err
	}

	return timestamp.Before(cutoff), nil
}

// Store saves a snapshot to object storage.
func (c *CloudObjectStorage) Store(ctx context.Context, snapshot *types.Snapshot) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Parse timestamp for key
	timestamp, err := snapshot.ParseTimestamp()
	if err != nil {
		return fmt.Errorf("invalid timestamp: %w", err)
	}

	// Create key from timestamp
	key := c.prefix + timestamp.Format("2006-01-02T15-04-05.000000000Z") + ".json"

	// Marshal to JSON
	jsonData, err := snapshot.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	// Upload to object storage
	if err := c.client.PutObject(ctx, c.bucket, key, jsonData); err != nil {
		return fmt.Errorf("failed to upload snapshot: %w", err)
	}

	return nil
}

// StoreBatch saves multiple snapshots to object storage.
func (c *CloudObjectStorage) StoreBatch(ctx context.Context, snapshots []*types.Snapshot) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	for _, snapshot := range snapshots {
		if err := c.Store(ctx, snapshot); err != nil {
			// Continue on error, but return last error
			// In production, you might want to collect all errors
			return err
		}
	}

	return nil
}

// Query retrieves snapshots from object storage.
func (c *CloudObjectStorage) Query(ctx context.Context, opts *QueryOptions) ([]*types.Snapshot, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	if opts == nil {
		opts = &QueryOptions{}
	}

	// List all objects
	keys, err := c.client.ListObjects(ctx, c.bucket, c.prefix)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	results := make([]*types.Snapshot, 0)

	// Fetch and filter objects
	for _, key := range keys {
		data, err := c.client.GetObject(ctx, c.bucket, key)
		if err != nil {
			continue // Skip objects that can't be read
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
	c.sortSnapshots(results, opts.OrderBy)

	// Apply limit
	if opts.Limit > 0 && len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	return results, nil
}

// Close stops cleanup and releases resources.
func (c *CloudObjectStorage) Close() error {
	if c.maxAge > 0 {
		// Stop cleanup goroutine
		close(c.stopCleanup)
		c.cleanupTicker.Stop()

		// Wait for cleanup to finish
		select {
		case <-c.cleanupDone:
		case <-time.After(5 * time.Second):
			// Timeout waiting for cleanup
		}
	}

	return nil
}

// sortSnapshots sorts snapshots by timestamp.
func (c *CloudObjectStorage) sortSnapshots(snapshots []*types.Snapshot, orderBy OrderBy) {
	sort.Slice(snapshots, func(i, j int) bool {
		ti, _ := snapshots[i].ParseTimestamp()
		tj, _ := snapshots[j].ParseTimestamp()
		if orderBy == OrderByTimeAsc {
			return ti.Before(tj)
		}
		return tj.Before(ti) // Default: newest first
	})
}

