package types

import (
	"encoding/json"
	"time"
)

// Snapshot represents a complete runtime snapshot at a point in time.
// This is the main data structure that can be stored using the SDK.
type Snapshot struct {
	// Timestamp is the UTC time when this snapshot was collected.
	// Format: RFC3339Nano (e.g., "2024-01-01T12:00:00.123456789Z")
	Timestamp string `json:"timestamp"`

	// Runtime contains Go runtime information.
	Runtime *RuntimeInfo `json:"runtime"`

	// Memory contains memory usage and garbage collection statistics.
	Memory *MemoryInfo `json:"memory"`

	// Goroutines contains goroutine count information.
	Goroutines *GoroutineInfo `json:"goroutines"`
}

// RuntimeInfo contains Go runtime metrics.
type RuntimeInfo struct {
	// GoVersion is the Go version string (e.g., "go1.24.5").
	GoVersion string `json:"go_version"`

	// NumGoroutines is the current number of goroutines.
	NumGoroutines int `json:"num_goroutines"`

	// GOMAXPROCS is the maximum number of CPUs that can be used simultaneously.
	GOMAXPROCS int `json:"gomaxprocs"`

	// NumCPU is the number of logical CPUs available.
	NumCPU int `json:"num_cpu"`

	// UptimeSeconds is the process uptime in seconds.
	UptimeSeconds float64 `json:"uptime_seconds"`
}

// MemoryInfo contains memory usage and GC statistics.
type MemoryInfo struct {
	// HeapInUseBytes is the number of bytes in use by the heap.
	HeapInUseBytes uint64 `json:"heap_in_use_bytes"`

	// HeapAllocatedBytes is the number of bytes currently allocated on the heap.
	HeapAllocatedBytes uint64 `json:"heap_allocated_bytes"`

	// HeapObjects is the number of allocated heap objects.
	HeapObjects uint64 `json:"heap_objects"`

	// TotalAllocBytes is the cumulative bytes allocated for heap objects.
	TotalAllocBytes uint64 `json:"total_alloc_bytes"`

	// GCCycles is the number of completed GC cycles.
	GCCycles uint32 `json:"gc_cycles"`

	// LastGCPauseSeconds is the duration of the last GC pause in seconds.
	LastGCPauseSeconds float64 `json:"last_gc_pause_seconds"`

	// GCCPUFraction is the fraction of CPU time spent in GC.
	GCCPUFraction float64 `json:"gc_cpu_fraction"`
}

// GoroutineInfo contains goroutine count information.
type GoroutineInfo struct {
	// TotalCount is the total number of goroutines.
	TotalCount int `json:"total_count"`
}

// ParseTimestamp parses the timestamp string and returns a time.Time.
// Returns an error if the timestamp format is invalid.
func (s *Snapshot) ParseTimestamp() (time.Time, error) {
	return time.Parse(time.RFC3339Nano, s.Timestamp)
}

// ToJSON converts the snapshot to JSON bytes.
// Returns an error if marshaling fails.
func (s *Snapshot) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// FromJSON creates a Snapshot from JSON bytes.
// Returns an error if unmarshaling fails.
func FromJSON(data []byte) (*Snapshot, error) {
	var snapshot Snapshot
	err := json.Unmarshal(data, &snapshot)
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

