package Worker

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/iguazio/io_blaster/Config"
	"github.com/iguazio/io_blaster/Utils"
	"github.com/jmoiron/jsonq"
	log "github.com/sirupsen/logrus"
)

type IWorker interface {
	Init(config *Config.ConfigIoBlaster, configWorkload *Config.ConfigWorkload, workloadIndex int64, workerIndex int64, calculatedWorkloadConstVars Config.CalculatedVars)
	GetIndex() int64
	GetStats() Config.Stats
	GetRealStartTimeNsec() int64
	GetRealEndTimeNsec() int64
	DebugLogCurrentIO(err error)
	PanicLogCurrentIO(err error)
	Run()
	RunIO()
	InitIO()
	SendIO()
	GetIOStatus() string
	GetIOResponseData() string
	FreeIO()
	CalculateNextVars()
	ParseField(fieldConfig *Config.ConfigField) interface{}
}

type WorkerBase struct {
	worker                      IWorker
	config                      *Config.ConfigIoBlaster
	configWorkload              *Config.ConfigWorkload
	currentRunTime              int64
	workloadIndex               int64
	workerIndex                 int64
	stats                       Config.Stats
	realStartTimeNsec           int64
	realEndTimeNsec             int64
	calculatedWorkloadConstVars Config.CalculatedVars
	calculatedVars              Config.CalculatedVars
	currentIOStatus             string
	currentIOLatency            int64
}

func (worker *WorkerBase) Init(config *Config.ConfigIoBlaster, configWorkload *Config.ConfigWorkload, workloadIndex int64, workerIndex int64, calculatedWorkloadConstVars Config.CalculatedVars) {
	worker.config = config
	worker.configWorkload = configWorkload
	worker.workloadIndex = workloadIndex
	worker.workerIndex = workerIndex
	worker.stats.StatusCounters = make(map[string]uint64, 0)
	worker.stats.StatusCountersPct = make(map[string]float64, 0)
	worker.stats.LatencyCounters = make(map[int64]uint64, 0)
	worker.stats.LatencyCountersPct = make(map[int64]float64, 0)
	worker.calculatedWorkloadConstVars = calculatedWorkloadConstVars
}

func (worker *WorkerBase) GetIndex() int64 {
	return worker.workerIndex
}

func (worker *WorkerBase) GetStats() Config.Stats {
	return worker.stats
}

func (worker *WorkerBase) GetRealStartTimeNsec() int64 {
	return worker.realStartTimeNsec
}

func (worker *WorkerBase) GetRealEndTimeNsec() int64 {
	return worker.realEndTimeNsec
}

func (worker *WorkerBase) Run() {
	worker.currentRunTime = atomic.LoadInt64(&worker.config.CurrentRunTime)
	worker.InitCalculatedVars(worker.calculatedWorkloadConstVars)
	endTime := worker.currentRunTime + int64(worker.configWorkload.Duration)
	worker.realStartTimeNsec = time.Now().UnixNano()
	shouldRun := worker.currentRunTime < endTime
	for shouldRun {
		for varName, valueConfig := range worker.configWorkload.EndOnVarValue {
			if compare_res, err := Utils.CompareInterface(valueConfig.Op, worker.calculatedVars[varName], worker.ParseField(valueConfig)); err != nil {
				log.Panicln(fmt.Sprintf("Workload %s found end_on_value config with unsupported op or missmatched value types. config=%s", worker.configWorkload.Name, valueConfig))
			} else {
				shouldRun = !compare_res
			}
		}
		if !shouldRun {
			break
		}
		worker.RunIO()
		worker.currentRunTime = atomic.LoadInt64(&worker.config.CurrentRunTime)
		shouldRun = worker.currentRunTime < endTime
		if shouldRun {
			worker.CalculateNextVars()
		}
	}
	worker.realEndTimeNsec = time.Now().UnixNano()
	worker.configWorkload.WorkersRunWaitGroup.Done()
}

func (worker *WorkerBase) RunIO() {
	worker.worker.InitIO()
	startTime := time.Now()
	worker.worker.SendIO()
	worker.currentIOLatency = time.Now().Sub(startTime).Nanoseconds() / 1000
	worker.currentIOStatus = worker.worker.GetIOStatus()
	worker.worker.DebugLogCurrentIO(nil)
	if _, ok := worker.configWorkload.AllowedStatusMap[worker.currentIOStatus]; !ok {
		worker.worker.PanicLogCurrentIO(errors.New("got unallowed status"))
	}
	worker.UpdateStats(worker.currentIOStatus, worker.currentIOLatency)
	worker.UpdateResponseVars(worker.currentIOStatus, string(worker.worker.GetIOResponseData()))
	worker.worker.FreeIO()

}

func (worker *WorkerBase) DebugLogCurrentIO(err error) {
}

func (worker *WorkerBase) PanicLogCurrentIO(err error) {
}

func (worker *WorkerBase) InitCalculatedVars(calculatedWorkloadConstVars Config.CalculatedVars) {
	worker.calculatedVars = make(Config.CalculatedVars, 0)
	for key, val := range calculatedWorkloadConstVars {
		worker.calculatedVars[key] = val
	}

	if _, ok := worker.calculatedVars["io_blaster_uid"]; ok {
		log.Panicln(fmt.Sprintf("Workload %s contain var with reserved name %s", worker.configWorkload.Name, "io_blaster_uid"))
	}
	worker.calculatedVars["io_blaster_uid"] = worker.GetRequestUid()

	if _, ok := worker.calculatedVars["io_blaster_worker_id"]; ok {
		log.Panicln(fmt.Sprintf("Workload %s contain var with reserved name %s", worker.configWorkload.Name, "io_blaster_worker_id"))
	}
	worker.calculatedVars["io_blaster_worker_id"] = worker.workerIndex

	if worker.configWorkload.Vars == nil {
		return
	}

	for varName, varConfig := range worker.configWorkload.Vars.ResponseValue {
		if _, ok := worker.calculatedVars[varName]; ok {
			log.Panicln(fmt.Sprintf("Workload %s contain 2 vars with same name %s", worker.configWorkload.Name, varName))
		}
		worker.calculatedVars[varName] = varConfig.InitValue
		worker.calculatedVars.RunTriggers(varConfig.Triggers, varName, worker.configWorkload.Name)
	}

	if worker.configWorkload.Vars.Random != nil {
		worker.calculatedVars.CalculatedRandomVarsConfig(worker.configWorkload.Name, worker.configWorkload.Vars.Random.WorkerOnce, true)
		worker.calculatedVars.CalculatedRandomVarsConfig(worker.configWorkload.Name, worker.configWorkload.Vars.Random.Each, true)
	}

	if worker.configWorkload.Vars.Enum != nil {
		for varName, varConfig := range worker.configWorkload.Vars.Enum.WorkerEach {
			if _, ok := worker.calculatedVars[varName]; ok {
				log.Panicln(fmt.Sprintf("Workload %s contain 2 vars with same name %s", worker.configWorkload.Name, varName))
			}
			worker.calculatedVars[varName] = varConfig.MinValue
			worker.calculatedVars.RunTriggers(varConfig.Triggers, varName, worker.configWorkload.Name)
			if _, ok := worker.calculatedVars[varName].(float64); ok {
				worker.calculatedVars[varName] = int64(worker.calculatedVars[varName].(float64))
			}
		}

		for varName, varConfig := range worker.configWorkload.Vars.Enum.WorkloadSimEach {
			if _, ok := worker.calculatedVars[varName]; ok {
				log.Panicln(fmt.Sprintf("Workload %s contain 2 vars with same name %s", worker.configWorkload.Name, varName))
			}
			worker.calculatedVars[varName] = varConfig.MinValue + int64(worker.workerIndex)
			worker.calculatedVars.RunTriggers(varConfig.Triggers, varName, worker.configWorkload.Name)
			if _, ok := worker.calculatedVars[varName].(float64); ok {
				worker.calculatedVars[varName] = int64(worker.calculatedVars[varName].(float64))
			}
		}

		for varName, varConfig := range worker.configWorkload.Vars.Enum.OnTime {
			if _, ok := worker.calculatedVars[varName]; ok {
				log.Panicln(fmt.Sprintf("Workload %s contain 2 vars with same name %s", worker.configWorkload.Name, varName))
			}
			worker.calculatedVars[varName] = varConfig.MinValue
			worker.calculatedVars.RunTriggers(varConfig.Triggers, varName, worker.configWorkload.Name)
			if _, ok := worker.calculatedVars[varName].(float64); ok {
				worker.calculatedVars[varName] = int64(worker.calculatedVars[varName].(float64))
			}
		}
	}

	// keep this var parsing last (othar than config field) since it depends on other array var
	for varName, varConfig := range worker.configWorkload.Vars.Dist {
		if _, ok := worker.calculatedVars[varName]; ok {
			log.Panicln(fmt.Sprintf("Workload %s contain 2 vars with same name %s", worker.configWorkload.Name, varName))
		}

		if _, ok := worker.calculatedVars[varConfig.ArrayVarName]; !ok {
			log.Panicln(fmt.Sprintf("Workload %s contain Dist var %s that is mapped to non existing array var name %s", worker.configWorkload.Name, varName, varConfig.ArrayVarName))
		}

		arrayVar := worker.calculatedVars[varConfig.ArrayVarName].([]interface{})
		arrayVarLen := int64(len(arrayVar))
		arrayIndex := int64(worker.workerIndex) % arrayVarLen
		worker.calculatedVars[varName] = arrayVar[arrayIndex]
		worker.calculatedVars.RunTriggers(varConfig.Triggers, varName, worker.configWorkload.Name)
	}

	// keep this var parsing last since it might depend on other vars
	if worker.configWorkload.Vars.ConfigField != nil {
		for varName, varConfig := range worker.configWorkload.Vars.ConfigField {
			if _, ok := worker.calculatedVars[varName]; ok {
				log.Panicln(fmt.Sprintf("Workload %s contain 2 vars with same name %s", worker.configWorkload.Name, varName))
			}
			worker.calculatedVars[varName] = worker.ParseField(varConfig)
			worker.calculatedVars.RunTriggers(varConfig.Triggers, varName, worker.configWorkload.Name)
		}
	}
}

func (worker *WorkerBase) CalculateNextVars() {
	if worker.configWorkload.Vars == nil {
		return
	}

	if worker.configWorkload.Vars.Random != nil {
		worker.calculatedVars.CalculatedRandomVarsConfig(worker.configWorkload.Name, worker.configWorkload.Vars.Random.Each, false)
	}

	if worker.configWorkload.Vars.Enum != nil {
		for varName, varConfig := range worker.configWorkload.Vars.Enum.WorkerEach {
			worker.calculatedVars[varName] = worker.calculatedVars[varName].(int64) + 1
			worker.calculatedVars.RunTriggers(varConfig.Triggers, varName, worker.configWorkload.Name)
			if _, ok := worker.calculatedVars[varName].(float64); ok {
				worker.calculatedVars[varName] = int64(worker.calculatedVars[varName].(float64))
			}
		}

		for varName, varConfig := range worker.configWorkload.Vars.Enum.WorkloadSimEach {
			worker.calculatedVars[varName] = worker.calculatedVars[varName].(int64) + worker.configWorkload.NumWorkers
			worker.calculatedVars.RunTriggers(varConfig.Triggers, varName, worker.configWorkload.Name)
			if _, ok := worker.calculatedVars[varName].(float64); ok {
				worker.calculatedVars[varName] = int64(worker.calculatedVars[varName].(float64))
			}
		}

		for varName, varConfig := range worker.configWorkload.Vars.Enum.OnTime {
			if worker.currentRunTime%varConfig.Interval == 0 {
				worker.calculatedVars[varName] = varConfig.MinValue + worker.currentRunTime
				worker.calculatedVars.RunTriggers(varConfig.Triggers, varName, worker.configWorkload.Name)
				if _, ok := worker.calculatedVars[varName].(float64); ok {
					worker.calculatedVars[varName] = int64(worker.calculatedVars[varName].(float64))
				}
			}
		}
	}

	worker.calculatedVars["io_blaster_uid"] = worker.GetRequestUid()

	// keep this var parsing last since it might depend on other vars
	if worker.configWorkload.Vars.ConfigField != nil {
		for varName, varConfig := range worker.configWorkload.Vars.ConfigField {
			worker.calculatedVars[varName] = worker.ParseField(varConfig)
			worker.calculatedVars.RunTriggers(varConfig.Triggers, varName, worker.configWorkload.Name)
		}
	}
}

func (worker *WorkerBase) UpdateStats(statusStr string, latency int64) {
	worker.stats.StatusCounters[statusStr]++
	worker.stats.LatencyCounters[Config.GetLatencyGroup(latency)]++
	worker.stats.TotalRequests++
}

func (worker *WorkerBase) UpdateResponseVars(statusStr string, responseData string) {
	if worker.configWorkload.Vars == nil {
		return
	}

	var responseHaveValidJson bool = true
	var jsonQuery *jsonq.JsonQuery
	jsonData := map[string]interface{}{}
	jsonDecoder := json.NewDecoder(strings.NewReader(responseData))
	if err := jsonDecoder.Decode(&jsonData); err != nil {
		responseHaveValidJson = false
	} else {
		jsonQuery = jsonq.NewQuery(jsonData)
	}

	var err error
	for varName, varConfig := range worker.configWorkload.Vars.ResponseValue {
		var need_to_update bool
		if len(varConfig.UpdateOnStatus) == 0 {
			need_to_update = true
		} else {
			for _, currentVarConfigStatusStr := range varConfig.UpdateOnStatus {
				if currentVarConfigStatusStr == statusStr {
					need_to_update = true
					break
				}
			}
		}
		if need_to_update {
			var responseValue interface{}
			if len(varConfig.FieldPath) == 0 {
				responseValue = responseData
			} else if responseHaveValidJson {
				if responseValue, err = jsonQuery.Interface(varConfig.FieldPath...); err != nil {
					err = errors.New(fmt.Sprintf("failed to find response_value var %s in response", varName))
					worker.worker.DebugLogCurrentIO(err)
					continue
				}
			} else {
				err = errors.New(fmt.Sprintf("failed to find response_value var %s in response", varName))
				worker.worker.DebugLogCurrentIO(err)
				continue
			}

			worker.calculatedVars[varName] = responseValue
			if len(varConfig.ExpectedValues) > 0 {
				expectedValueFound := false
				for _, expectedValuesConfig := range varConfig.ExpectedValues {
					if worker.ParseField(expectedValuesConfig) == responseValue {
						expectedValueFound = true
						break
					}
				}
				if !expectedValueFound {
					err = errors.New(fmt.Sprintf("failed to find response_value var %s in response", varName))
					worker.worker.PanicLogCurrentIO(err)
				}
			}
			worker.calculatedVars.RunTriggers(varConfig.Triggers, varName, worker.configWorkload.Name)
		}
	}
}

func (worker *WorkerBase) ParseField(fieldConfig *Config.ConfigField) interface{} {
	return Config.ParseField(fieldConfig, worker.calculatedVars, worker.configWorkload.Name)
}

func (worker *WorkerBase) GetRequestUid() uint64 {
	return uint64(worker.workloadIndex<<48) + uint64((worker.workerIndex << 32)) + worker.stats.TotalRequests
}
