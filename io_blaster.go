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

	workload_manager Workload.WorkloadManager
	workloads        map[string]Workload.IWorkload
)

const AppVersion = "1.1.0"

func parse_cmd_line_args() {
	flag.BoolVar(&showVersion, "version", false, "show version")
	flag.StringVar(&config_file, "c", "", "config file path")
	flag.StringVar(&results_file, "o", "", "results file path")
	flag.BoolVar(&verbose, "v", false, "print debug logs")
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
	var err error = nil
	log_file, err = os.Create(file_name)
	if err != nil {
		log.Panicln("failed to open log file")
	} else {
		var log_writers io.Writer
		log_writers = io.MultiWriter(os.Stdout, log_file)
		log.SetOutput(log_writers)
	}
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
			workload.Init(&config, int64(workloadIndex))
			break
		case "SHELL":
			workload = new(Workload.WorkloadShell)
			workload.Init(&config, int64(workloadIndex))
			break
		default:
			log.Panicln(fmt.Sprintf("Found workload with unsupported type. %+v", workload_config))
		}

		workloads[workload_config.Name] = workload
		config.WorkloadsMap[workload_config.Name] = workload_config
	}
}

func process_results() {
	for _, workload := range workloads {
		var workloadStats Config.Stats
		workloadStats.StatusCounters = make(map[string]uint64, 0)
		workloadStats.StatusCountersPct = make(map[string]float64, 0)
		workloadStats.LatencyCounters = make(map[int64]uint64, 0)
		workloadStats.LatencyCountersPct = make(map[int64]float64, 0)
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

			log.Debugln(fmt.Sprintf("workload %s worker %d stats: %+v", workload.Name(), worker.GetIndex(), workerStats))
		}

		for status, StatusCount := range workloadStats.StatusCounters {
			workloadStats.StatusCountersPct[status] = math.Round((float64(StatusCount)*100/float64(workloadStats.TotalRequests))*1000) / 1000
		}

		for latencyGroup, latencyGroupCount := range workloadStats.LatencyCounters {
			workloadStats.LatencyCountersPct[latencyGroup] = math.Round((float64(latencyGroupCount)*100/float64(workloadStats.TotalRequests))*1000) / 1000
		}

		workloadStats.Iops = math.Round(workloadStats.Iops*1000) / 1000

		log.Infoln(fmt.Sprintf("workload %s stats: %+v", workload.Name(), workloadStats))
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	parse_cmd_line_args()
	init_log()
	log.Infoln("io_blaster started")
	config.LoadConfig(config_file)
	create_workloads()
	workload_manager.Init(&config, workloads)
	workload_manager.Run()
	process_results()
	log.Infoln("io_blaster finished")
}
