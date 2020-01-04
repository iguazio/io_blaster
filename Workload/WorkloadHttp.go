package Workload

import (
	"github.com/iguazio/io_blaster/Config"
	"github.com/iguazio/io_blaster/Workload/Worker"
)

type WorkloadHttp struct {
	WorkloadBase
}

func (workload *WorkloadHttp) Init(config *Config.ConfigIoBlaster, workloadIndex int64) {
	workload.workloadHooks.createWorker = createHttpWorker
	workload.WorkloadBase.Init(config, workloadIndex)
}

func createHttpWorker() Worker.IWorker {
	return new(Worker.WorkerHttp)
}
