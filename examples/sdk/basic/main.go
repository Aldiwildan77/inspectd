package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Aldiwildan77/inspectd/sdk"
	"github.com/Aldiwildan77/inspectd/sdk/storage"
)

func main() {
	// Example 1: Using in-memory storage
	fmt.Println("=== Example 1: In-Memory Storage ===")
	memStorage := storage.NewMemoryStorage()
	defer memStorage.Close()

	client := sdk.NewClient(memStorage)

	// Collect and store a snapshot
	ctx := context.Background()
	if err := client.CollectAndStore(ctx); err != nil {
		log.Fatal(err)
	}

	// Query recent snapshots
	snapshots, err := client.QueryRecent(ctx, 10)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Stored %d snapshot(s)\n", len(snapshots))
	if len(snapshots) > 0 {
		fmt.Printf("Latest snapshot timestamp: %s\n", snapshots[0].Timestamp)
		fmt.Printf("Goroutines: %d\n", snapshots[0].Goroutines.TotalCount)
		fmt.Printf("Heap allocated: %d bytes\n", snapshots[0].Memory.HeapAllocatedBytes)
	}

	// Example 2: Using file storage
	fmt.Println("\n=== Example 2: File Storage ===")
	fileStorage, err := storage.NewFileStorage("./snapshots")
	if err != nil {
		log.Fatal(err)
	}
	defer fileStorage.Close()

	fileClient := sdk.NewClient(fileStorage)

	// Collect and store multiple snapshots
	for i := 0; i < 3; i++ {
		if err := fileClient.CollectAndStore(ctx); err != nil {
			log.Fatal(err)
		}
		time.Sleep(100 * time.Millisecond) // Small delay between snapshots
	}

	// Query all snapshots
	allSnapshots, err := fileClient.QueryRecent(ctx, 0)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Stored %d snapshot(s) to files\n", len(allSnapshots))

	// Example 3: Query by time range
	fmt.Println("\n=== Example 3: Query by Time Range ===")
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now()

	rangeSnapshots, err := fileClient.QueryByTimeRange(ctx, startTime, endTime, 10)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d snapshot(s) in the last hour\n", len(rangeSnapshots))
}

