package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {
	fmt.Println("Demo program for inspectd")
	fmt.Println("This program creates goroutines and allocates memory")
	fmt.Println("Run inspectd commands in another terminal to inspect this process")
	fmt.Println()

	// Create some goroutines
	for i := 0; i < 5; i++ {
		go func(id int) {
			for {
				time.Sleep(time.Second)
				_ = fmt.Sprintf("goroutine %d", id)
			}
		}(i)
	}

	// Allocate some memory
	var data [][]byte
	for i := 0; i < 100; i++ {
		buf := make([]byte, 1024*1024)
		data = append(data, buf)
	}

	fmt.Printf("Created %d goroutines\n", runtime.NumGoroutine())
	fmt.Printf("Allocated ~%d MB\n", len(data))
	fmt.Println("Press Ctrl+C to exit")

	// Keep running
	select {}
}
