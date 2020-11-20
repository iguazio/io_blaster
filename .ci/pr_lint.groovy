@Library('pipelinex@development') _
label = "${UUID.randomUUID().toString()}"

podTemplate(containers: [
    containerTemplate(name: 'golang', image: 'golang:1.14.12', ttyEnabled: true, command: 'cat'),
    containerTemplate(name: 'golangci-lint', image: 'golangci/golangci-lint:v1.32-alpine', ttyEnabled: true, command: 'cat'),
  ],
    envVars: [
        envVar(key: 'GO111MODULE', value: 'on'), 
        envVar(key: 'GOPROXY', value: 'https://goproxy.devops.iguazeng.com')
    ],
  ) {
      node("io_blaster-pr-lint-${label}") {
          common.notify_slack {
            stage('Check running containers') {
                container('golang') {
                    sh "export"
                    checkout scm 
                    sh "go mod download"
                }
                container('golangci-lint') {
                    sh "golangci-lint run"
                }
            }
          }
      }
  }