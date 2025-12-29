package runtimeinfo

import (
	"encoding/json"
	"runtime"
	"time"
)

type RuntimeInfo struct {
	GoVersion    string  `json:"go_version"`
	NumGoroutines int    `json:"num_goroutines"`
	GOMAXPROCS   int     `json:"gomaxprocs"`
	NumCPU       int     `json:"num_cpu"`
	Uptime       float64 `json:"uptime_seconds"`
}

func Collect() (*RuntimeInfo, error) {
	uptime := time.Since(startTime).Seconds()
	
	info := &RuntimeInfo{
		GoVersion:     runtime.Version(),
		NumGoroutines: runtime.NumGoroutine(),
		GOMAXPROCS:    runtime.GOMAXPROCS(0),
		NumCPU:        runtime.NumCPU(),
		Uptime:        uptime,
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

var startTime = time.Now()

