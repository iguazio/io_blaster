package Config

import (
	"encoding/json"
	"fmt"
	"github.com/iguazio/io_blaster/Utils"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"time"
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

type CalculatedVars map[string]interface{}

type ConfigVarsConst struct {
	Value interface{} `json:"value"`
}

type ConfigVarsFile struct {
	Path string `json:"path"`
}

type ConfigVarsRandomOrEnum struct {
	Type     string `json:"type"`
	Length   int    `json:"length"`
	MinValue int64  `json:"min_value"`
	MaxValue int64  `json:"max_value"`
	Interval int64  `json:"interval"`
}

type ConfigVarResponseValue struct {
	UpdateOnStatus []string       `json:"update_on_status"`
	FieldPath      []string       `json:"field_path"`
	InitValue      interface{}    `json:"init_value"`
	ExpectedValues []*ConfigField `json:"expected_values"`
}

type ConfigVarsRandomOrEnumMap map[string]*ConfigVarsRandomOrEnum

type ConfigVarsRandom struct {
	Once       ConfigVarsRandomOrEnumMap `json:"once"`
	WorkerOnce ConfigVarsRandomOrEnumMap `json:"worker_once"`
	Each       ConfigVarsRandomOrEnumMap `json:"each"`
	OnTime     ConfigVarsRandomOrEnumMap `json:"on_time"`
	OnInterval ConfigVarsRandomOrEnumMap `json:"on_interval"`
}

type ConfigVarsEnum struct {
	WorkloadSimEach ConfigVarsRandomOrEnumMap `json:"workload_sim_each"`
	WorkerEach      ConfigVarsRandomOrEnumMap `json:"worker_each"`
	OnTime          ConfigVarsRandomOrEnumMap `json:"on_time"`
}

type ConfigVars struct {
	Const         map[string]*ConfigVarsConst        `json:"const"`
	File          map[string]*ConfigVarsFile         `json:"file"`
	Random        *ConfigVarsRandom                  `json:"random"`
	Enum          *ConfigVarsEnum                    `json:"enum"`
	ConfigField   map[string]*ConfigField            `json:"config_field"`
	ResponseValue map[string]*ConfigVarResponseValue `json:"response_value"`
}

type ConfigField struct {
	Type            string      `json:"type"`
	Op              string      `json:"op"`
	Format          string      `json:"format"`
	ArrayArgs       []string    `json:"array_args"`
	ArrayJoinString string      `json:"array_join_string"`
	FormatArgs      []string    `json:"args"`
	Value           interface{} `json:"value"`
	VarName         string      `json:"var_name"`
}

type ConfigHttp struct {
	RequestTimeout time.Duration           `json:"request_timeout"`
	Method         *ConfigField            `json:"method"`
	Url            *ConfigField            `json:"url"`
	Headers        map[string]*ConfigField `json:"headers"`
	Body           *ConfigField            `json:"body"`
}

type ConfigShell struct {
	User     *ConfigField `json:"user"`
	Password *ConfigField `json:"password"`
	Host     *ConfigField `json:"host"`
	Cmd      *ConfigField `json:"cmd"`
}

type ConfigWorkload struct {
	Name                 string `json:"name"`
	WorkloadObj          interface{}
	WorkloadRunWaitGroup sync.WaitGroup
	WorkersRunWaitGroup  sync.WaitGroup
	AllowedStatus        []string `json:"allowed_status"`
	AllowedStatusMap     map[string]bool
	StartTime            time.Duration           `json:"start_time"`
	Duration             time.Duration           `json:"duration"`
	DependsOnWorkload    []string                `json:"depends_on_workload"`
	EndOnVarValue        map[string]*ConfigField `json:"end_on_var_value"`
	NumWorkers           int64                   `json:"workers"`
	Vars                 *ConfigVars             `json:"vars"`
	Type                 string                  `json:"type"`
	HttpConfig           *ConfigHttp             `json:"http_config"`
	ShellConfig          *ConfigShell            `json:"shell_config"`
}

type ConfigIoBlaster struct {
	Workloads      []*ConfigWorkload `json:"workloads"`
	WorkloadsMap   map[string]*ConfigWorkload
	CurrentRunTime int64
}

func VarRunRandom(varConfig *ConfigVarsRandomOrEnum) interface{} {
	switch varConfig.Type {
	case "STRING":
		if varConfig.Length == 0 {
			log.Panicln(fmt.Sprintf("Found random string var with legnth=0. var=%+v", varConfig))
		}
		return Utils.GenerateRandomString(varConfig.Length)
	case "BASE64":
		if varConfig.Length == 0 {
			log.Panicln(fmt.Sprintf("Found random base64 var with legnth=0. var=%+v", varConfig))
		}
		return Utils.GenerateRandomBase64(varConfig.Length)
	case "INT":
		if varConfig.MaxValue <= varConfig.MinValue {
			log.Panicln(fmt.Sprintf("Found random int var with max_value <= min_value. var=%+v", varConfig))
		}
		var seededRand *rand.Rand = Utils.GetSeededRandom()
		return seededRand.Int63n(varConfig.MaxValue-varConfig.MinValue+1) + varConfig.MinValue
	default:
		log.Panicln(fmt.Sprintf("Found random var with unsupported type. var=%+v", varConfig))
	}

	return nil
}

func (calculatedVars CalculatedVars) CalculatedRandomVarsConfig(workloadName string, configVarsRandomOrEnumMap ConfigVarsRandomOrEnumMap, assertExist bool) {
	for varName, varConfig := range configVarsRandomOrEnumMap {
		if assertExist {
			if _, ok := calculatedVars[varName]; ok {
				log.Panicln(fmt.Sprintf("Workload %s contain 2 vars with same name %s", workloadName, varName))
			}
		}
		calculatedVars[varName] = VarRunRandom(varConfig)
	}
}

func (config *ConfigIoBlaster) LoadConfig(config_file_path string) {
	json_file, err := os.Open(config_file_path)
	if err != nil {
		log.Panicln("Failed to open config file")
	}
	defer json_file.Close()

	byteValue, _ := ioutil.ReadAll(json_file)
	err = json.Unmarshal(byteValue, config)
	if err != nil {
		log.Panicln("Failed to parse config file json", err)
	}
}
