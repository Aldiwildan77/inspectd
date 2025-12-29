package goroutines

import (
	"encoding/json"
	"runtime"
)

type GoroutineInfo struct {
	TotalCount int `json:"total_count"`
}

func Collect() (*GoroutineInfo, error) {
	info := &GoroutineInfo{
		TotalCount: runtime.NumGoroutine(),
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

