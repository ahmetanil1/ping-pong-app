pipeline {
    agent {
        kubernetes {
            yaml '''
apiVersion: v1
kind: Pod
spec:
  # Every dynamic agent uses the ServiceAccount created by the Jenkins Helm chart.
  serviceAccountName: jenkins-agent

  volumes:
    # Docker layers live only for the lifetime of this dynamic agent Pod.
    - name: docker-storage
      emptyDir: {}

  # why we need git? difference betwwen git and jnlp?
  # The git container is used to clone the repository and perform Git operations,
  # while the jnlp container is used for Jenkins agent communication with the Jenkins master.
  # The jnlp container runs the Jenkins agent process,
  # which connects to the Jenkins master and executes the pipeline steps.
  # It has to be git inside because the jnlp container needs to have access to the Git binary to perform Git operations.
  # The git container is specifically for Git-related tasks,
  # while the jnlp container handles the overall communication and execution of the pipeline.
  containers:
    # The Kubernetes plugin automatically adds the jnlp container.
    # It connects this Pod to the Jenkins controller.

    # Clones the repository and calculates the commit SHA image tag.
    - name: git
      image: alpine/git:2.47.2
      command:
        - cat
      tty: true

    # Runs Go unit tests.
    - name: golang
      image: golang:1.24.6-alpine
      command:
        - cat
      tty: true

    # Validates the Helm chart and will deploy it in the next stage.
    - name: helm
      image: alpine/helm:3.17.0
      command:
        - cat
      tty: true

    # Sends Docker build and push commands to the DinD daemon.
    - name: docker-cli
      image: docker:27.5.1-cli
      command:
        - cat
      tty: true
      env:
        # The Docker daemon runs in the dind sidecar in the same Pod.
        - name: DOCKER_HOST
          value: tcp://127.0.0.1:2375

    # Provides an isolated Docker daemon for this ephemeral agent Pod.
    - name: dind
      image: docker:27.5.1-dind
      securityContext:
        # DinD requires privileged mode in this local Minikube setup.
        privileged: true
      command:
        - dockerd
      args:
        # The daemon is reachable only from containers in this same Pod.
        - --host=tcp://127.0.0.1:2375
        - --tls=false
      volumeMounts:
        - name: docker-storage
          mountPath: /var/lib/docker
'''
        }
    }

    options {
        // Stops a stuck Pipeline instead of consuming resources indefinitely.
        timeout(time: 30, unit: 'MINUTES')

        // Adds a timestamp to every console log line.
        timestamps()

        // Checkout is performed explicitly in the Checkout stage.
        skipDefaultCheckout()
    }

    stages {
        stage('Checkout') {
            steps {
                // Makes the Git container responsible for cloning the repository.
                container('git') {
                    checkout scm
                }
            }
        }

        stage('Prepare Image Tag') {
            steps {
                container('git') {
                    script {
                         // Jenkins checkout scm can run through the JNLP agent launcher,
                         // while this command runs in the Git container.
                         // Both containers share the same workspace but can use different Linux users.
                         // Git requires the shared ephemeral workspace to be explicitly trusted.
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

        stage('Build and Push Images') {
            steps {
                container('docker-cli') {
                    withCredentials([
                        usernamePassword(
                            credentialsId: 'dockerhub-credentials',
                            usernameVariable: 'DOCKERHUB_USERNAME',
                            passwordVariable: 'DOCKERHUB_TOKEN'
                        )
                    ])  {
                        sh '''
                            set -eu

                            # Waits briefly until the DinD daemon is ready.
                            for attempt in $(seq 1 30); do
                                if docker info > /dev/null 2>&1; then
                                    break
                                fi

                                echo "Waiting for the Docker daemon..."
                                sleep 2
                            done

                            # Fails explicitly if the Docker daemon did not start.
                            docker info > /dev/null

                            # The token is passed through standard input and is masked in Jenkins logs.
                            echo "$DOCKERHUB_TOKEN" | docker login \
                                --username "$DOCKERHUB_USERNAME" \
                                --password-stdin

                            # Both images are built before any image is pushed.
                            docker build \
                                --tag "$DOCKERHUB_USERNAME/ping-service:$IMAGE_TAG" \
                                --tag "$DOCKERHUB_USERNAME/ping-service:latest" \
                                ./ping-service

                            docker build \
                                --tag "$DOCKERHUB_USERNAME/pong-service:$IMAGE_TAG" \
                                --tag "$DOCKERHUB_USERNAME/pong-service:latest" \
                                ./pong-service

                            # The commit SHA is the deployable and traceable image version.
                            docker push "$DOCKERHUB_USERNAME/ping-service:$IMAGE_TAG"
                            docker push "$DOCKERHUB_USERNAME/pong-service:$IMAGE_TAG"

                            # latest is provided only as a convenience tag.
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
