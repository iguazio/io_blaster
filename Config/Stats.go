package Config

import (
	"encoding/json"
	"io/ioutil"
)

var (
	LatencyGroups = []int64{ // all latencies are in usec
		0, 100, 200, 300, 400, 500, 600, 700, 800, 900, // 0 msec - 1 msec
		1000, 1200, 1400, 1600, 1800, // 1 msec - 2 msec
		2000, 2500, 3000, 4000, 5000, // 2 msec - 10 msec
		10000, 25000, 50000, 75000, // 10 msec - 100 msec
		100000, 20000, 30000, 40000, 50000, 750000, // 100 msec - 1 sec
		1000000, 2000000, 3000000, 4000000, 5000000, // 1 sec - 10 sec
		10000000, 20000000, 40000000, 60000000, 80000000, 100000000, 120000000} // 10 sec - 120 sec
)

type Stats struct {
	ExactRunDuration   float64            `json:"exact_run_duration"`
	TotalRequests      uint64             `json:"total_requests"`
	Iops               float64            `json:"iops"`
	StatusCounters     map[string]uint64  `json:"status_counters"`
	StatusCountersPct  map[string]float64 `json:"status_counters_pct"`
	LatencyCounters    map[int64]uint64   `json:"latency_counters"`
	LatencyCountersPct map[int64]float64  `json:"latency_counters_pct"`
}

type StatsDumpWorkload struct {
	WorkerStats map[int64]*Stats `json:"workers"`
	Stats       *Stats           `json:"stats"`
}

type StatsDumpIoBlaster struct {
	WorkloadsStats map[string]*StatsDumpWorkload `json:"workloads"`
}

func (statsDump *StatsDumpIoBlaster) WriteStatsDumpToFile(statsFile string) {
	if statsFile != "" {
		statsData, _ := json.MarshalIndent(statsDump, "", " ")
		_ = ioutil.WriteFile(statsFile, statsData, 0644)
	}
}

func GetLatencyGroup(latency int64) int64 {
	for groupIndex := 0; groupIndex < len(LatencyGroups)-1; groupIndex++ {
		if latency < LatencyGroups[groupIndex+1] {
			return LatencyGroups[groupIndex]
		}
	}
	return LatencyGroups[len(LatencyGroups)-1]
}
