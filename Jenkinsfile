pipeline {
    agent {
        kubernetes {
            defaultContainer 'git'
            yaml '''
apiVersion: v1
kind: Pod
spec:
  # Uses the ServiceAccount created by the Jenkins Helm chart.
  serviceAccountName: jenkins-agent

  containers:
    # This container clones the repository.
    - name: git
      image: alpine/git:2.47.2
      command:
        - cat
      tty: true

    # This container runs Go tests.
    - name: golang
      image: golang:1.24.6-alpine
      command:
        - cat
      # tty is required for the container to stay alive while the Jenkins agent is running.
      # allows the container to run commands(cat) interactively.
      tty: true

    # This container validates the Helm chart.
    - name: helm
      image: alpine/helm:3.17.0
      command:
        - cat
      tty: true
'''
        }
    }

    options {
        timestamps()

        // Checkout is performed explicitly in the Checkout stage.
        skipDefaultCheckout()
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Test Go Services') {
            steps {
                container('golang') {
                    sh '''
                        set -eu

                        cd ping-service
                        go test -v ./...

                        cd ../pong-service
                        go test -v ./...
                    '''
                }
            }
        }

        stage('Validate Helm Chart') {
            steps {
                container('helm') {
                    sh '''
                        set -eu

                        helm dependency build deployments
                        helm lint deployments
                    '''
                }
            }
        }
    }

    post {
        success {
            echo 'Dynamic Kubernetes agent completed test and Helm validation stages.'
        }

        failure {
            echo 'Pipeline failed. Check the failed stage logs.'
        }
    }
}
