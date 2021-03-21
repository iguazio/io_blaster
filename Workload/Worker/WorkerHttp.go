package Worker

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/iguazio/io_blaster/Config"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

type WorkerHttp struct {
	WorkerBase
	httpClient *fasthttp.Client
	request    *fasthttp.Request
	response   *fasthttp.Response
}

func (worker *WorkerHttp) Init(config *Config.ConfigIoBlaster, configWorkload *Config.ConfigWorkload, workloadIndex int64, workerIndex int64, calculatedWorkloadConstVars Config.CalculatedVars) {
	worker.worker = worker
	worker.WorkerBase.Init(config, configWorkload, workloadIndex, workerIndex, calculatedWorkloadConstVars)

	if configWorkload.HttpConfig.RequestTimeout == 0 {
		configWorkload.HttpConfig.RequestTimeout = 120
	}

	worker.httpClient = &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) {
			return fasthttp.DialTimeout(addr, time.Second*120)
		},
		MaxConnsPerHost: 1,
	}
}

func (worker *WorkerHttp) InitIO() {
	worker.request = fasthttp.AcquireRequest()
	worker.response = fasthttp.AcquireResponse()
	method := worker.ParseField(worker.configWorkload.HttpConfig.Method).(string)
	url := worker.ParseField(worker.configWorkload.HttpConfig.Url).(string)
	body := worker.ParseField(worker.configWorkload.HttpConfig.Body).(string)
	worker.request.Header.SetMethod(method)
	worker.request.SetRequestURI(url)
	for headerName, headerConfig := range worker.configWorkload.HttpConfig.Headers {
		worker.request.Header.Add(headerName, fmt.Sprintf("%v", worker.ParseField(headerConfig)))
	}
	worker.request.SetBodyString(body)
}

func (worker *WorkerHttp) SendIO() {
	if err := worker.httpClient.DoTimeout(worker.request, worker.response, worker.configWorkload.HttpConfig.RequestTimeout*time.Second); err != nil {
		worker.PanicLogCurrentIO(err)
	}
}

func (worker *WorkerHttp) GetIOStatus() string {
	return strconv.Itoa(worker.response.StatusCode())
}

func (worker *WorkerHttp) GetIOResponseData() string {
	return string(worker.response.Body())
}

func (worker *WorkerHttp) FreeIO() {
	defer fasthttp.ReleaseRequest(worker.request)
	defer fasthttp.ReleaseResponse(worker.response)
}

func (worker *WorkerHttp) DebugLogCurrentIO(err error) {
	if err != nil {
		log.Debugln(fmt.Sprintf("workload=%s worker=%d status=%s latency=%d err=%s request=\n%+v\n\nresponse=\n%+v\n", worker.configWorkload.Name, worker.workerIndex, worker.currentIOStatus, worker.currentIOLatency, err.Error(), worker.request, worker.response))
	} else {
		log.Debugln(fmt.Sprintf("workload=%s worker=%d status=%s latency=%d request=\n%+v\n\nresponse=\n%+v\n", worker.configWorkload.Name, worker.workerIndex, worker.currentIOStatus, worker.currentIOLatency, worker.request, worker.response))
	}
}

func (worker *WorkerHttp) PanicLogCurrentIO(err error) {
	if err != nil {
		log.Panicln(fmt.Sprintf("workload=%s worker=%d status=%s latency=%d err=%s request=\n%+v\n\nresponse=\n%+v\ncalculatedVars=%+v\n", worker.configWorkload.Name, worker.workerIndex, worker.currentIOStatus, worker.currentIOLatency, err.Error(), worker.request, worker.response, worker.calculatedVars))
	} else {
		log.Panicln(fmt.Sprintf("workload=%s worker=%d status=%s latency=%d request=\n%+v\n\nresponse=\n%+v\ncalculatedVars=%+v\n", worker.configWorkload.Name, worker.workerIndex, worker.currentIOStatus, worker.currentIOLatency, worker.request, worker.response, worker.calculatedVars))
	}
}
