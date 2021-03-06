package Workload

import (
	"github.com/iguazio/io_blaster/Config"
	"github.com/iguazio/io_blaster/Workload/Worker"
)

type WorkloadHttp struct {
	WorkloadBase
}

func (workload *WorkloadHttp) Init(config *Config.ConfigIoBlaster, workloadIndex int64, calculatedGlobalConstVars Config.CalculatedVars) {
	workload.workloadHooks.createWorker = createHttpWorker
	workload.WorkloadBase.Init(config, workloadIndex, calculatedGlobalConstVars)
}

func createHttpWorker() Worker.IWorker {
	return new(Worker.WorkerHttp)
}
