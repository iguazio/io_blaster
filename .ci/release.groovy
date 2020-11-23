@Library('pipelinex@development') _

podTemplate(

    label: 'ioblaster-release',
    containers: [
        containerTemplate(name: 'jnlp', image: 'jenkins/jnlp-slave:4.0.1-1', workingDir: '/home/jenkins', resourceRequestCpu: '2000m', resourceLimitCpu: '2000m', resourceRequestMemory: '2048Mi', resourceLimitMemory: '2048Mi'),
        containerTemplate(name: 'golang', image: 'golang:1.14.12', workingDir: '/home/jenkins', ttyEnabled: true, command: 'cat'),
    ],
) {
      node("ioblaster-release") {
          common.notify_slack {
            container('golang') {

                stage('|Obtain goreleaser binary') {
                    sh "curl -sfL https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh | sh"
                }

                stage('Build and release binaries') {

                    withCredentials([
                        usernamePassword(credentialsId: 'iguazio-prod-artifactory-credentials',
                            usernameVariable: 'artifactory_user',
                            passwordVariable: 'artifactory_password'),
                        string(credentialsId: git_mirror_user_token, variable: 'GIT_MIRROR_TOKEN')
                    ]) {
                        withEnv([
                            "GITHUB_TOKEN=${github_token}",
                            "ARTIFACTORY_IGUAZIO_USERNAME=${artifactory_user}",
                            "ARTIFACTORY_IGUAZIO_SECRET=${artifactory_password}",
                        ]) {
                            checkout scm
                            sh "goreleaser release"
                        }
                    }
                }
            }
          }
      }
  }