package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Aldiwildan77/inspectd/sdk"
	"github.com/Aldiwildan77/inspectd/sdk/storage"
)

// Example production usage with managed file storage (with cleanup)
func main() {
	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create managed file storage with retention policies
	// - Max 1000 files
	// - Max age: 7 days
	// - Cleanup every hour
	fileStorage, err := storage.NewManagedFileStorage("/var/lib/inspectd/snapshots", storage.ManagedFileStorageConfig{
		MaxFiles:       1000,
		MaxAge:         7 * 24 * time.Hour,
		CleanupInterval: 1 * time.Hour,
	})
	if err != nil {
		log.Fatalf("Failed to create file storage: %v", err)
	}
	defer fileStorage.Close()

	// Create SDK client
	client := sdk.NewClient(fileStorage)
	defer client.Close()

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Collect and store multiple snapshots
	for i := 0; i < 5; i++ {
		if err := client.CollectAndStore(ctx); err != nil {
			log.Printf("Failed to collect snapshot %d: %v", i+1, err)
		} else {
			fmt.Printf("Snapshot %d stored\n", i+1)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Get storage stats
	count, err := fileStorage.Stats()
	if err != nil {
		log.Printf("Failed to get stats: %v", err)
	} else {
		fmt.Printf("Total snapshots stored: %d\n", count)
	}

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutting down gracefully...")
}

