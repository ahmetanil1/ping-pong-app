SHELL := /bin/bash

CLUSTER_NAME := hepapi-cluster

NAMESPACE := ping-pong
RELEASE_NAME := ping-pong-app
CHART_PATH := ./deployments

JENKINS_NAMESPACE := jenkins
JENKINS_RELEASE := jenkins
JENKINS_CHART := jenkinsci/jenkins
JENKINS_VALUES_FILE := ./jenkins/values.yaml
JENKINS_JCASC_FILE := ./jenkins/jcasc/jenkins.yaml
JENKINS_RBAC_FILE := ./jenkins/rbac/ping-pong-deployer.yaml
DOCKERHUB_SECRET_NAME := dockerhub-credentials

.PHONY: setup
# Creates the local Minikube cluster and enables its required addons.
setup:
	@chmod +x scripts/cluster-setup.sh
	@./scripts/cluster-setup.sh
.PHONY: jenkins-rbac
# Creates the application namespace and applies Jenkins agent deploy permissions.
jenkins-rbac:
	@kubectl create namespace $(NAMESPACE) \
		--dry-run=client \
		--output=yaml \
		| kubectl apply --filename=-
	@kubectl apply --filename $(JENKINS_RBAC_FILE)

.PHONY: jenkins-secret
# Creates the Jenkins namespace and Docker Hub Credential Secret.
# The token is requested only when the Secret does not already exist.
jenkins-secret:
	@kubectl create namespace $(JENKINS_NAMESPACE) \
		--dry-run=client \
		--output=yaml \
		| kubectl apply --filename=-
	@set -e; \
	if kubectl --namespace $(JENKINS_NAMESPACE) \
		get secret $(DOCKERHUB_SECRET_NAME) > /dev/null 2>&1; then \
		echo "Docker Hub Secret already exists. Skipping creation."; \
	else \
		read -r -p "Docker Hub username: " dockerhub_username; \
		read -r -s -p "Docker Hub access token: " dockerhub_token; \
		echo; \
		kubectl --namespace $(JENKINS_NAMESPACE) \
			create secret generic $(DOCKERHUB_SECRET_NAME) \
			--from-literal=dockerhub-username="$$dockerhub_username" \
			--from-literal=dockerhub-token="$$dockerhub_token"; \
		unset dockerhub_username dockerhub_token; \
	fi

.PHONY: jenkins-install
# Installs Jenkins for the first time or upgrades the existing Jenkins release.
jenkins-install: jenkins-rbac jenkins-secret
	@helm repo add jenkinsci https://charts.jenkins.io --force-update
	@helm repo update
	@helm upgrade --install $(JENKINS_RELEASE) $(JENKINS_CHART) \
		--namespace $(JENKINS_NAMESPACE) \
		--values $(JENKINS_VALUES_FILE) \
		--set-file controller.JCasC.configScripts.project-configuration=$(JENKINS_JCASC_FILE) \
		--wait \
		--atomic \
		--timeout 10m

.PHONY: build-images
# Builds the Docker images for the services before loading them into Minikube.
build-images:
	@echo "Building Docker images..."
	@docker build -t ping-service:latest ./ping-service
	@docker build -t pong-service:latest ./pong-service

.PHONY: load-images
# Loads locally built images into the Minikube nodes.
load-images: build-images
	@echo "Loading Docker images into the Kubernetes cluster $(CLUSTER_NAME)..."
	@minikube image load ping-service:latest -p $(CLUSTER_NAME)
	@minikube image load pong-service:latest -p $(CLUSTER_NAME)

.PHONY: lint
# Builds dependencies and validates the Helm chart locally.
lint:
	@echo "Building Helm dependencies..."
	@helm dependency build $(CHART_PATH) > /dev/null 2>&1
	@echo "Linting Helm charts..."
	@helm lint $(CHART_PATH)
	@rm -f $(CHART_PATH)/charts/*.tgz $(CHART_PATH)/Chart.lock

.PHONY: deploy
# Deploys the application with local Minikube images.
deploy: load-images
	@echo "Deploying the application via Helm..."
	@helm upgrade --install $(RELEASE_NAME) $(CHART_PATH) \
		--namespace $(NAMESPACE) \
		--create-namespace

.PHONY: undeploy
# Removes the application Helm release.
undeploy:
	@echo "Undeploying the application via Helm."
	@helm uninstall $(RELEASE_NAME) --namespace $(NAMESPACE) || true