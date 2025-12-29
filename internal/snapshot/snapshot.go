package snapshot

import (
	"encoding/json"
	"time"

	"github.com/Aldiwildan77/inspectd/internal/goroutines"
	"github.com/Aldiwildan77/inspectd/internal/memory"
	"github.com/Aldiwildan77/inspectd/internal/runtimeinfo"
)

type Snapshot struct {
	Timestamp  string                    `json:"timestamp"`
	Runtime    *runtimeinfo.RuntimeInfo  `json:"runtime"`
	Memory     *memory.MemoryInfo        `json:"memory"`
	Goroutines *goroutines.GoroutineInfo `json:"goroutines"`
}

func Collect() (*Snapshot, error) {
	runtimeInfo, err := runtimeinfo.Collect()
	if err != nil {
		return nil, err
	}

	memInfo, err := memory.Collect()
	if err != nil {
		return nil, err
	}

	goroutineInfo, err := goroutines.Collect()
	if err != nil {
		return nil, err
	}

	snapshot := &Snapshot{
		Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
		Runtime:    runtimeInfo,
		Memory:     memInfo,
		Goroutines: goroutineInfo,
	}

	return snapshot, nil
}

func CollectJSON() ([]byte, error) {
	snapshot, err := Collect()
	if err != nil {
		return nil, err
	}
	return json.Marshal(snapshot)
}
