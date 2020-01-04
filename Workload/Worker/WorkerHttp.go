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

	worker.httpClient = &fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) {
			return fasthttp.DialTimeout(addr, time.Second*120)
		},
		MaxConnsPerHost: 1,
	}
}

func (worker *WorkerHttp) RunIO() {
	worker.request = fasthttp.AcquireRequest()
	worker.response = fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(worker.request)
	defer fasthttp.ReleaseResponse(worker.response)

	method := worker.ParseField(worker.configWorkload.HttpConfig.Method).(string)
	url := worker.ParseField(worker.configWorkload.HttpConfig.Url).(string)
	body := worker.ParseField(worker.configWorkload.HttpConfig.Body).(string)
	worker.request.Header.SetMethod(method)
	worker.request.SetRequestURI(url)
	for headerName, headerConfig := range worker.configWorkload.HttpConfig.Headers {
		worker.request.Header.Add(headerName, fmt.Sprintf("%v", worker.ParseField(headerConfig)))
	}
	worker.request.SetBodyString(body)

	startTime := time.Now()
	if err := worker.httpClient.DoTimeout(worker.request, worker.response, 120*time.Second); err != nil {
		log.Panicln(fmt.Sprintf("workload=%s worker=%d url=%s failed to send request. err=%s", worker.configWorkload.Name, worker.workerIndex, url, err.Error()))
	} else {
		latency := time.Now().Sub(startTime).Nanoseconds() / 1000
		statusStr := strconv.Itoa(worker.response.StatusCode())

		log.Debugln(fmt.Sprintf("workload=%s worker=%d url=%s status=%d latency=%d", worker.configWorkload.Name, worker.workerIndex, url, worker.response.StatusCode(), latency))

		worker.UpdateStats(statusStr, latency)

		if _, ok := worker.configWorkload.AllowedStatusMap[statusStr]; !ok {
			log.Panicln(fmt.Sprintf("workload=%s worker=%d got unallowed status request=\n%+v\n\nresponse=\n%+v\n", worker.configWorkload.Name, worker.workerIndex, worker.request, worker.response))
		}

		worker.UpdateResponseVars(statusStr, string(worker.response.Body()))
	}
}

func (worker *WorkerHttp) DebugLogCurrentIO(err error) {
	log.Debugln(fmt.Sprintf("%s request=\n%+v\n\nresponse=\n%+v\n", err.Error(), worker.request, worker.response))
}

func (worker *WorkerHttp) WarnLogCurrentIO(err error) {
	log.Warnln(fmt.Sprintf("%s request=\n%+v\n\nresponse=\n%+v\n", err.Error(), worker.request, worker.response))
}

func (worker *WorkerHttp) PanicLogCurrentIO(err error) {
	log.Panicln(fmt.Sprintf("%s request=\n%+v\n\nresponse=\n%+v\n", err.Error(), worker.request, worker.response))
}
