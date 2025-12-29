package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Aldiwildan77/inspectd/sdk"
	"github.com/Aldiwildan77/inspectd/sdk/storage"
	"github.com/Aldiwildan77/inspectd/sdk/types"
)

// CustomStorage is an example of implementing a custom storage backend.
// This example stores snapshots in a simple JSON array format.
type CustomStorage struct {
	snapshots []*types.Snapshot
}

// Ensure CustomStorage implements storage.Storage interface
var _ storage.Storage = (*CustomStorage)(nil)

func NewCustomStorage() *CustomStorage {
	return &CustomStorage{
		snapshots: make([]*types.Snapshot, 0),
	}
}

func (c *CustomStorage) Store(ctx context.Context, snapshot *types.Snapshot) error {
	// Create a copy to avoid external modifications
	snapshotCopy := *snapshot
	c.snapshots = append(c.snapshots, &snapshotCopy)
	return nil
}

func (c *CustomStorage) StoreBatch(ctx context.Context, snapshots []*types.Snapshot) error {
	for _, snapshot := range snapshots {
		if err := c.Store(ctx, snapshot); err != nil {
			return err
		}
	}
	return nil
}

func (c *CustomStorage) Query(ctx context.Context, opts *storage.QueryOptions) ([]*types.Snapshot, error) {
	// Simple implementation - in a real scenario, you would apply filters
	results := make([]*types.Snapshot, len(c.snapshots))
	for i, snapshot := range c.snapshots {
		snapshotCopy := *snapshot
		results[i] = &snapshotCopy
	}
	return results, nil
}

func (c *CustomStorage) Close() error {
	c.snapshots = nil
	return nil
}

// ExportToJSON demonstrates how to export all snapshots to JSON
func (c *CustomStorage) ExportToJSON() ([]byte, error) {
	return json.MarshalIndent(c.snapshots, "", "  ")
}

func main() {
	fmt.Println("=== Custom Storage Example ===")

	// Create custom storage
	customStorage := NewCustomStorage()
	defer customStorage.Close()

	// Create SDK client with custom storage
	client := sdk.NewClient(customStorage)

	// Collect and store snapshots
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if err := client.CollectAndStore(ctx); err != nil {
			log.Fatal(err)
		}
	}

	// Export to JSON
	jsonData, err := customStorage.ExportToJSON()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Exported snapshots to JSON:")
	fmt.Println(string(jsonData))
}

