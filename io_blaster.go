package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"runtime"

	"github.com/iguazio/io_blaster/Config"
	"github.com/iguazio/io_blaster/Workload"

	logrus_stack "github.com/Gurpartap/logrus-stack"
	log "github.com/sirupsen/logrus"
)

var (
	showVersion  bool
	config_file  string
	config       Config.ConfigIoBlaster
	results_file string
	log_file     *os.File
	verbose      bool = false
	stats_file   string

	workload_manager          Workload.WorkloadManager
	workloads                 map[string]Workload.IWorkload
	calculatedGlobalConstVars Config.CalculatedVars

	// will be overwritten by goreleaser
	AppVersion = "development"
)

func parse_cmd_line_args() {
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.StringVar(&config_file, "c", "", "config file path")
	flag.StringVar(&results_file, "o", "", "results file path")
	flag.BoolVar(&verbose, "v", false, "print debug logs")
	flag.StringVar(&stats_file, "s", "", "stats file path")
	flag.Parse()

	if showVersion {
		fmt.Println(AppVersion)
		os.Exit(0)
	}

	if config_file == "" {
		fmt.Println("Usage error: must set config file path")
		os.Exit(1)
	}

	if results_file == "" {
		fmt.Println("Usage error: must set result file path")
		os.Exit(1)
	}
}

func init_log() {

	log.SetFormatter(&log.TextFormatter{ForceColors: true,
		FullTimestamp:   true,
		TimestampFormat: "2006/01/02-15:04:05"})

	if verbose {
		log.SetLevel(log.DebugLevel)
		log.AddHook(logrus_stack.StandardHook())
	} else {
		log.SetLevel(log.InfoLevel)
	}

	file_name := fmt.Sprintf("%s.log", results_file)
	var err error
	log_file, err = os.Create(file_name)
	if err != nil {
		log.Panicln("failed to open log file")
	} else {
		var log_writers io.Writer = io.MultiWriter(os.Stdout, log_file)
		log.SetOutput(log_writers)
	}
}

func initGlobalVars() {
	calculatedGlobalConstVars = make(Config.CalculatedVars)
	calculatedGlobalConstVars.CalculateConstVars("Global vars", config.Vars)
}

func create_workloads() {
	if len(config.Workloads) == 0 {
		log.Panicln("No workloads defined in config json")
	}

	workloads = make(map[string]Workload.IWorkload)
	config.WorkloadsMap = make(map[string]*Config.ConfigWorkload)

	for workloadIndex, workload_config := range config.Workloads {
		log.Debugln(fmt.Sprintf("found workload %+v", workload_config))
		var workload Workload.IWorkload

		if workload_config.Name == "" {
			log.Panicln(fmt.Sprintf("Found workload config with no name. %+v", workload_config))
		}

		if workload_config.Duration <= 0 {
			log.Panicln(fmt.Sprintf("Found workload config with end_time <= start_time. %+v", workload_config))
		}

		if workload_config.NumWorkers == 0 {
			log.Panicln(fmt.Sprintf("Found workload config with no workers. %+v", workload_config))
		}

		if workload_config.Type == "" {
			log.Panicln(fmt.Sprintf("Found workload config with no type. %+v", workload_config))
		}

		if _, ok := workloads[workload_config.Name]; ok {
			log.Panicln(fmt.Sprintf("Found multiple workloads with same name. name:%s", workload_config.Name))
		}

		switch workload_config.Type {
		case "HTTP":
			workload = new(Workload.WorkloadHttp)
			workload.Init(&config, int64(workloadIndex), calculatedGlobalConstVars)
		case "SHELL":
			workload = new(Workload.WorkloadShell)
			workload.Init(&config, int64(workloadIndex), calculatedGlobalConstVars)
		default:
			log.Panicln(fmt.Sprintf("Found workload with unsupported type. %+v", workload_config))
		}

		workloads[workload_config.Name] = workload
		config.WorkloadsMap[workload_config.Name] = workload_config
	}
}

func process_results() {
	statsData := new(Config.StatsDumpIoBlaster)
	statsData.WorkloadsStats = make(map[string]*Config.StatsDumpWorkload)

	for _, workload := range workloads {
		var workloadStats Config.Stats
		workloadStatsDump := new(Config.StatsDumpWorkload)
		workloadStatsDump.WorkerStats = make(map[int64]*Config.Stats)
		statsData.WorkloadsStats[workload.Name()] = workloadStatsDump
		workloadStats.StatusCounters = make(map[string]uint64)
		workloadStats.StatusCountersPct = make(map[string]float64)
		workloadStats.LatencyCounters = make(map[int64]uint64)
		workloadStats.LatencyCountersPct = make(map[int64]float64)

		for _, worker := range workload.GetWorkers() {
			workerStats := worker.GetStats()

			workloadStats.TotalRequests += workerStats.TotalRequests

			for status, StatusCount := range workerStats.StatusCounters {
				workloadStats.StatusCounters[status] += StatusCount
				workerStats.StatusCountersPct[status] = math.Round((float64(StatusCount)*100/float64(workerStats.TotalRequests))*1000) / 1000
			}

			for latencyGroup, latencyGroupCount := range workerStats.LatencyCounters {
				workloadStats.LatencyCounters[latencyGroup] += latencyGroupCount
				workerStats.LatencyCountersPct[latencyGroup] = math.Round((float64(latencyGroupCount)*100/float64(workerStats.TotalRequests))*1000) / 1000
			}

			workerStats.ExactRunDuration = float64(worker.GetRealEndTimeNsec()-worker.GetRealStartTimeNsec()) / 1000000000
			workerStats.Iops = math.Round((float64(workerStats.TotalRequests)/workerStats.ExactRunDuration)*1000) / 1000
			workloadStats.Iops += workerStats.Iops
			if workerStats.ExactRunDuration > workloadStats.ExactRunDuration {
				workloadStats.ExactRunDuration = workerStats.ExactRunDuration
			}

			workloadStatsDump.WorkerStats[worker.GetIndex()] = &workerStats

			log.Debugln(fmt.Sprintf("workload %s worker %d stats: %+v", workload.Name(), worker.GetIndex(), workerStats))
		}

		for status, StatusCount := range workloadStats.StatusCounters {
			workloadStats.StatusCountersPct[status] = math.Round((float64(StatusCount)*100/float64(workloadStats.TotalRequests))*1000) / 1000
		}

		for latencyGroup, latencyGroupCount := range workloadStats.LatencyCounters {
			workloadStats.LatencyCountersPct[latencyGroup] = math.Round((float64(latencyGroupCount)*100/float64(workloadStats.TotalRequests))*1000) / 1000
		}

		workloadStats.Iops = math.Round(workloadStats.Iops*1000) / 1000

		workloadStatsDump.Stats = &workloadStats

		log.Infoln(fmt.Sprintf("workload %s stats: %+v", workload.Name(), workloadStats))
	}

	statsData.WriteStatsDumpToFile(stats_file)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	parse_cmd_line_args()
	init_log()
	log.Infoln("io_blaster started")
	config.LoadConfig(config_file)
	initGlobalVars()
	create_workloads()
	workload_manager.Init(&config, workloads)
	workload_manager.Run()
	process_results()
	log.Infoln("io_blaster finished")
}
