pipeline {
    // Each parent stage provisions its own dedicated Kubernetes agent Pod.
    agent none

    options {
        timeout(time: 30, unit: 'MINUTES')
        timestamps()
        skipDefaultCheckout()
    }

    stages {
        stage('Test and Validate') {
            agent {
                kubernetes {
                    yaml '''
apiVersion: v1
kind: Pod
spec:
  serviceAccountName: jenkins-agent

  containers:
    # The Kubernetes plugin automatically adds the jnlp container.

    - name: git
      image: alpine/git:2.47.2
      command:
        - cat
      tty: true

    - name: golang
      image: golang:1.24.6-alpine
      command:
        - cat
      tty: true

    - name: helm
      image: alpine/helm:3.17.0
      command:
        - cat
      tty: true
'''
                }
            }

            stages {
                stage('Checkout') {
                    steps {
                        container('git') {
                            checkout scm
                        }
                    }
                }

                stage('Test Go Services') {
                    steps {
                        container('golang') {
                            sh '''
                                set -eu

                                cd ping-service
                                GOMAXPROCS=1 go test -p 1 -v ./...

                                cd ../pong-service
                                GOMAXPROCS=1 go test -p 1 -v ./...
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
        }

        stage('Build and Push Images') {
            agent {
                kubernetes {
                    yaml '''
apiVersion: v1
kind: Pod
spec:
  serviceAccountName: jenkins-agent

  volumes:
    # Docker layers exist only while this build Pod exists.
    - name: docker-storage
      emptyDir: {}

  containers:
    # The Kubernetes plugin automatically adds the jnlp container.

    - name: git
      image: alpine/git:2.47.2
      command:
        - cat
      tty: true

    - name: docker-cli
      image: docker:27.5.1-cli
      command:
        - cat
      tty: true
      env:
        - name: DOCKER_HOST
          value: tcp://127.0.0.1:2375

    - name: dind
      image: docker:27.5.1-dind
      securityContext:
        # Required by Docker-in-Docker in this local Minikube environment.
        privileged: true
      command:
        - dockerd
      args:
        # The Docker API is reachable only inside this agent Pod.
        - --host=tcp://127.0.0.1:2375
        - --tls=false
      volumeMounts:
        - name: docker-storage
          mountPath: /var/lib/docker
'''
                }
            }

            steps {
                // This Pod has its own empty workspace, so it must checkout again.
                container('git') {
                    checkout scm
                }

                container('git') {
                    script {
                        // Jenkins checkout scm may use the JNLP launcher.
                        // Trust this temporary shared workspace before reading Git metadata.
                        env.IMAGE_TAG = sh(
                            script: '''
                                git config --global --add safe.directory "$WORKSPACE"
                                git rev-parse --short=7 HEAD
                            ''',
                            returnStdout: true
                        ).trim()
                    }

                    echo "Images will be tagged with: ${env.IMAGE_TAG}"
                }

                container('docker-cli') {
                    withCredentials([
                        usernamePassword(
                            credentialsId: 'dockerhub-credentials',
                            usernameVariable: 'DOCKERHUB_USERNAME',
                            passwordVariable: 'DOCKERHUB_TOKEN'
                        )
                    ]) {
                        sh '''
                            set -eu

                            for attempt in $(seq 1 30); do
                                if docker info > /dev/null 2>&1; then
                                    break
                                fi

                                echo "Waiting for the Docker daemon..."
                                sleep 2
                            done

                            docker info > /dev/null

                            echo "$DOCKERHUB_TOKEN" | docker login \
                                --username "$DOCKERHUB_USERNAME" \
                                --password-stdin

                            docker build \
                                --tag "$DOCKERHUB_USERNAME/ping-service:$IMAGE_TAG" \
                                --tag "$DOCKERHUB_USERNAME/ping-service:latest" \
                                ./ping-service

                            docker build \
                                --tag "$DOCKERHUB_USERNAME/pong-service:$IMAGE_TAG" \
                                --tag "$DOCKERHUB_USERNAME/pong-service:latest" \
                                ./pong-service

                            docker push "$DOCKERHUB_USERNAME/ping-service:$IMAGE_TAG"
                            docker push "$DOCKERHUB_USERNAME/pong-service:$IMAGE_TAG"

                            docker push "$DOCKERHUB_USERNAME/ping-service:latest"
                            docker push "$DOCKERHUB_USERNAME/pong-service:latest"

                            docker logout
                        '''
                    }
                }
            }
        }
    }

    post {
        success {
            echo 'CI completed: tests passed, Helm chart validated, and images were pushed.'
        }

        failure {
            echo 'Pipeline failed. Check the failed stage logs.'
        }
    }
}
