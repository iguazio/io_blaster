package Worker

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	sshLib "github.com/0xef53/go-sshpool"
	"github.com/iguazio/io_blaster/Config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type WorkerShell struct {
	WorkerBase
	sshPool          *sshLib.SSHPool
	sshConfig        *sshLib.SSHConfig
	currentCmd       string
	sendErr          error
	currentCmdOutput []byte
}

func (worker *WorkerShell) Init(config *Config.ConfigIoBlaster, configWorkload *Config.ConfigWorkload, workloadIndex int64, workerIndex int64, calculatedWorkloadConstVars Config.CalculatedVars) {
	worker.worker = worker
	worker.WorkerBase.Init(config, configWorkload, workloadIndex, workerIndex, calculatedWorkloadConstVars)

	worker.sshPool = sshLib.NewPool(&sshLib.PoolConfig{MaxConns: 100, GCInterval: time.Second * 60})

	if agentSocket, ok := os.LookupEnv("SSH_AUTH_SOCK"); !ok {
		cmd := exec.Command("eval", "\"$(ssh-agent -s)\"")
		_, err := cmd.CombinedOutput()
		if err != nil {
			log.Panicln(fmt.Sprintf("workload=%s worker=%d could not connect/start SSH_AUTH_SOCK. Is ssh-agent running?", worker.configWorkload.Name, worker.workerIndex))
		}
	} else {
		worker.sshConfig = &sshLib.SSHConfig{
			Port:            22,
			AgentSocket:     agentSocket,
			Timeout:         120 * time.Second,
			HostKeyCallback: ssh.InsecureIgnoreHostKey()}
	}
}

func (worker *WorkerShell) InitIO() {
	worker.sshConfig.User = worker.ParseField(worker.configWorkload.ShellConfig.User).(string)
	worker.sshConfig.Host = worker.ParseField(worker.configWorkload.ShellConfig.Host).(string)
	if worker.configWorkload.ShellConfig.Password != nil {
		worker.sshConfig.Auth = append(worker.sshConfig.Auth, ssh.Password(worker.ParseField(worker.configWorkload.ShellConfig.Password).(string)))
	}
	worker.currentCmd = worker.ParseField(worker.configWorkload.ShellConfig.Cmd).(string)
}

func (worker *WorkerShell) SendIO() {
	worker.currentCmdOutput, worker.sendErr = worker.sshPool.CombinedOutput(worker.sshConfig, worker.currentCmd, nil, nil)
}

func (worker *WorkerShell) GetIOStatus() string {
	var currentCmdStatus string
	if worker.sendErr != nil {
		if exitError, ok := worker.sendErr.(*ssh.ExitError); ok {
			currentCmdStatus = strconv.Itoa(exitError.ExitStatus())
		} else {
			currentCmdStatus = "N/A"
			worker.PanicLogCurrentIO(worker.sendErr)
		}
	} else {
		currentCmdStatus = "0"
	}
	return currentCmdStatus
}

func (worker *WorkerShell) GetIOResponseData() string {
	return string(worker.currentCmdOutput)
}

func (worker *WorkerShell) FreeIO() {
}

func (worker *WorkerShell) DebugLogCurrentIO(err error) {
	if err != nil {
		log.Debugln(fmt.Sprintf("workload=%s worker=%d host=%s user=%s status=%s latency=%d err=%s\ncmd=\n%s\n\noutput=\n%s\n", worker.configWorkload.Name, worker.workerIndex, worker.sshConfig.Host, worker.sshConfig.User, worker.currentIOStatus, worker.currentIOLatency, err.Error(), worker.currentCmd, worker.currentCmdOutput))
	} else {
		log.Debugln(fmt.Sprintf("workload=%s worker=%d host=%s user=%s status=%s latency=%d\ncmd=\n%s\n\noutput=\n%s\n", worker.configWorkload.Name, worker.workerIndex, worker.sshConfig.Host, worker.sshConfig.User, worker.currentIOStatus, worker.currentIOLatency, worker.currentCmd, worker.currentCmdOutput))
	}
}

func (worker *WorkerShell) PanicLogCurrentIO(err error) {
	if err != nil {
		log.Panicln(fmt.Sprintf("workload=%s worker=%d host=%s user=%s status=%s latency=%d err=%s\ncmd=\n%s\n\noutput=\n%s\ncalculatedVars=%+v\n", worker.configWorkload.Name, worker.workerIndex, worker.sshConfig.Host, worker.sshConfig.User, worker.currentIOStatus, worker.currentIOLatency, err.Error(), worker.currentCmd, worker.currentCmdOutput, worker.calculatedVars))
	} else {
		log.Panicln(fmt.Sprintf("workload=%s worker=%d host=%s user=%s status=%s latency=%d\ncmd=\n%s\n\noutput=\n%s\ncalculatedVars=%+v\n", worker.configWorkload.Name, worker.workerIndex, worker.sshConfig.Host, worker.sshConfig.User, worker.currentIOStatus, worker.currentIOLatency, worker.currentCmd, worker.currentCmdOutput, worker.calculatedVars))
	}
}
