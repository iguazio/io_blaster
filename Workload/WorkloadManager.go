package Workload

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/iguazio/io_blaster/Config"
	log "github.com/sirupsen/logrus"
)

type WorkloadManager struct {
	config       *Config.ConfigIoBlaster
	workloadsMap map[string]IWorkload
	runDone      bool
}

func (workloadManager *WorkloadManager) Init(config *Config.ConfigIoBlaster, workloadsMap map[string]IWorkload) {
	workloadManager.config = config
	workloadManager.workloadsMap = workloadsMap
}

func (workloadManager *WorkloadManager) Run() {
	workloadManager.config.CurrentRunTime = 0

	go time.AfterFunc(1*time.Second, workloadManager.Tick)

	for _, workload := range workloadManager.workloadsMap {
		go time.AfterFunc(workload.GetStartTime()*time.Second, workload.Run)
	}

	for _, workload := range workloadManager.workloadsMap {
		workload.WaitUntilDone()
	}
	workloadManager.runDone = true
}

func (workloadManager *WorkloadManager) Tick() {
	atomic.AddInt64(&workloadManager.config.CurrentRunTime, 1)
	log.Debugln(fmt.Sprintf("workload manager tick time:%d", workloadManager.config.CurrentRunTime))
	if !workloadManager.runDone {
		time.AfterFunc(1*time.Second, workloadManager.Tick)
	}
}
