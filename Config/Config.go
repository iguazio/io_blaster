package Config

import (
	"sync"
	"time"
)

type CalculatedVars map[string]interface{}

type ConfigVarsTrigger struct {
	OnValue    *ConfigField `json:"on_value"`
	VarToSet   string       `json:"var_to_set"`
	ValueToSet *ConfigField `json:"value_to_set"`
}

type ConfigVarsConst struct {
	Value    interface{}          `json:"value"`
	Triggers []*ConfigVarsTrigger `json:"triggers"`
}

type ConfigVarsFile struct {
	Path     string               `json:"path"`
	Triggers []*ConfigVarsTrigger `json:"triggers"`
}

type ConfigVarsRandomOrEnum struct {
	Type     string               `json:"type"`
	Length   int                  `json:"length"`
	MinValue int64                `json:"min_value"`
	MaxValue int64                `json:"max_value"`
	Interval int64                `json:"interval"`
	Triggers []*ConfigVarsTrigger `json:"triggers"`
}

type ConfigVarResponseValue struct {
	UpdateOnStatus []string             `json:"update_on_status"`
	FieldPath      []string             `json:"field_path"`
	InitValue      interface{}          `json:"init_value"`
	ExpectedValues []*ConfigField       `json:"expected_values"`
	Triggers       []*ConfigVarsTrigger `json:"triggers"`
}

type ConfigVarsDist struct {
	ArrayVarName string               `json:"array_var"`
	Triggers     []*ConfigVarsTrigger `json:"triggers"`
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
	Dist          map[string]*ConfigVarsDist         `json:"dist"`
	ConfigField   map[string]*ConfigField            `json:"config_field"`
	ResponseValue map[string]*ConfigVarResponseValue `json:"response_value"`
}

type ConfigField struct {
	Type            string               `json:"type"`
	Op              string               `json:"op"`
	Format          string               `json:"format"`
	ArrayArgs       []int                `json:"array_args"`
	ArrayJoinString string               `json:"array_join_string"`
	FormatArgs      []interface{}        `json:"args"`
	Value           interface{}          `json:"value"`
	VarName         string               `json:"var_name"`
	Triggers        []*ConfigVarsTrigger `json:"triggers"`
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
