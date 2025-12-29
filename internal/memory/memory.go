package memory

import (
	"encoding/json"
	"runtime"
)

type MemoryInfo struct {
	HeapInUse      uint64  `json:"heap_in_use_bytes"`
	HeapAllocated  uint64  `json:"heap_allocated_bytes"`
	HeapObjects    uint64  `json:"heap_objects"`
	TotalAlloc     uint64  `json:"total_alloc_bytes"`
	GCCycles       uint32  `json:"gc_cycles"`
	LastGCPause    float64 `json:"last_gc_pause_seconds"`
	GCCPUFraction  float64 `json:"gc_cpu_fraction"`
}

func Collect() (*MemoryInfo, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	var lastGCPause float64
	if m.NumGC > 0 {
		lastGCPause = float64(m.PauseNs[(m.NumGC+255)%256]) / 1e9
	}
	
	info := &MemoryInfo{
		HeapInUse:     m.HeapInuse,
		HeapAllocated: m.Alloc,
		HeapObjects:   m.HeapObjects,
		TotalAlloc:    m.TotalAlloc,
		GCCycles:      m.NumGC,
		LastGCPause:   lastGCPause,
		GCCPUFraction: m.GCCPUFraction,
	}
	
	return info, nil
}

func CollectJSON() ([]byte, error) {
	info, err := Collect()
	if err != nil {
		return nil, err
	}
	return json.Marshal(info)
}

