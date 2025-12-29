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

// Example production usage with graceful shutdown and bounded memory storage
func main() {
	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create bounded memory storage (production-safe with limits)
	// Limit to 1000 snapshots to prevent OOM
	memStorage := storage.NewBoundedMemoryStorage(1000)
	defer memStorage.Close()

	// Create SDK client
	client := sdk.NewClient(memStorage)
	defer client.Close()

	// Context with timeout for operations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Collect and store snapshot
	if err := client.CollectAndStore(ctx); err != nil {
		log.Printf("Failed to collect snapshot: %v", err)
	} else {
		fmt.Println("Snapshot collected and stored successfully")
	}

	// Query recent snapshots
	snapshots, err := client.QueryRecent(ctx, 10)
	if err != nil {
		log.Printf("Failed to query snapshots: %v", err)
	} else {
		fmt.Printf("Found %d recent snapshot(s)\n", len(snapshots))
		if len(snapshots) > 0 {
			latest := snapshots[0]
			fmt.Printf("Latest: %s - Goroutines: %d, Heap: %d bytes\n",
				latest.Timestamp,
				latest.Goroutines.TotalCount,
				latest.Memory.HeapAllocatedBytes)
		}
	}

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutting down gracefully...")
}

