package Workload

import (
	"github.com/iguazio/io_blaster/Config"
	"github.com/iguazio/io_blaster/Workload/Worker"
)

type WorkloadShell struct {
	WorkloadBase
}

func (workload *WorkloadShell) Init(config *Config.ConfigIoBlaster, workloadIndex int64) {
	workload.workloadHooks.createWorker = createShellWorker
	workload.WorkloadBase.Init(config, workloadIndex)
}

func createShellWorker() Worker.IWorker {
	return new(Worker.WorkerShell)
}
