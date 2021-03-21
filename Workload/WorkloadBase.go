package Workload

import (
	"fmt"
	//"io/ioutil"
	//"os"
	"time"

	"github.com/iguazio/io_blaster/Workload/Worker"

	"github.com/iguazio/io_blaster/Config"
	log "github.com/sirupsen/logrus"
)

type CreateWorker func() Worker.IWorker

type WorkloadHooks struct {
	createWorker CreateWorker
}

type IWorkload interface {
	Init(config *Config.ConfigIoBlaster, workloadIndex int64, calculatedGlobalConstVars Config.CalculatedVars)
	Name() string
	GetWorkers() []Worker.IWorker
	GetStartTime() time.Duration
	Run()
	WaitUntilDone()
}

type WorkloadBase struct {
	workloadHooks               WorkloadHooks
	config                      *Config.ConfigIoBlaster
	configWorkload              *Config.ConfigWorkload
	workloadIndex               int64
	workers                     []Worker.IWorker
	calculatedWorkloadConstVars Config.CalculatedVars
	calculatedGlobalConstVars   Config.CalculatedVars
}

func (workload *WorkloadBase) Init(config *Config.ConfigIoBlaster, workloadIndex int64, calculatedGlobalConstVars Config.CalculatedVars) {
	workload.config = config
	workload.workloadIndex = workloadIndex
	workload.configWorkload = workload.config.Workloads[workload.workloadIndex]
	workload.configWorkload.WorkloadObj = workload
	workload.configWorkload.AllowedStatusMap = make(map[string]bool)
	workload.calculatedGlobalConstVars = calculatedGlobalConstVars
	workload.CalculateWorkloadConstVars()
	for _, allowedStatus := range workload.configWorkload.AllowedStatus {
		workload.configWorkload.AllowedStatusMap[allowedStatus] = true
	}

	workload.workers = make([]Worker.IWorker, workload.configWorkload.NumWorkers)
	for workerIndex := int64(0); workerIndex < workload.configWorkload.NumWorkers; workerIndex++ {
		var worker Worker.IWorker = workload.workloadHooks.createWorker()
		worker.Init(workload.config, workload.configWorkload, workload.workloadIndex, workerIndex, workload.calculatedWorkloadConstVars)
		workload.workers[workerIndex] = worker
		workload.configWorkload.WorkersRunWaitGroup.Add(1)
	}
	workload.configWorkload.WorkloadRunWaitGroup.Add(1)
}

func (workload *WorkloadBase) Run() {

	for _, workloadName := range workload.configWorkload.DependsOnWorkload {
		workloadConfig, ok := workload.config.WorkloadsMap[workloadName]
		if !ok {
			log.Panicln(fmt.Sprintf("Failed to run workload %s. Depends on workload %s that doesn't exist in config", workload.Name(), workloadName))
		}
		workloadConfig.WorkloadObj.(IWorkload).WaitUntilDone()
	}

	log.Infoln(fmt.Sprintf("Workload %s starting", workload.Name()))
	for _, worker := range workload.workers {
		go worker.Run()
	}
	workload.configWorkload.WorkersRunWaitGroup.Wait()
	log.Infoln(fmt.Sprintf("Workload %s done", workload.Name()))
	workload.configWorkload.WorkloadRunWaitGroup.Done()
}

func (workload *WorkloadBase) CalculateWorkloadConstVars() {

	workload.calculatedWorkloadConstVars = make(Config.CalculatedVars)

	for key, val := range workload.calculatedGlobalConstVars {
		workload.calculatedWorkloadConstVars[key] = val
	}

	workload.calculatedWorkloadConstVars.CalculateConstVars(workload.configWorkload.GetVarsConfigLogPrefix(), workload.configWorkload.Vars)
}

func (workload *WorkloadBase) WaitUntilDone() {
	workload.configWorkload.WorkloadRunWaitGroup.Wait()
}

func (workload *WorkloadBase) Name() string {
	return workload.configWorkload.Name
}

func (workload *WorkloadBase) GetWorkers() []Worker.IWorker {
	return workload.workers
}

func (workload *WorkloadBase) GetStartTime() time.Duration {
	return workload.configWorkload.StartTime
}
