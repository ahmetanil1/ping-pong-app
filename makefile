.PHONY: setup deploy undeploy build-images load-images # define the targets as phony to avoid conflicts with files of the same name

CLUSTER_NAME := hepapi-cluster
NAMESPACE := ping-pong
RELEASE_NAME := ping-pong-app
CHART_PATH := ./deployments/helm/ping-pong-charts

# if you want to run the cluster setup script, you can use the following command:
setup:
	@chmod +x scripts/cluster-setup.sh
	@./scripts/cluster-setup.sh

# Build the Docker images for the services. This is necessary before loading them into the Kubernetes cluster.
build-images:
	@echo "Building Docker images..."
	@docker build -t ping-service:latest ./ping-service
	@docker build -t pong-service:latest ./pong-service

# firstly, you have to build the images, then you can load them into the Kubernetes cluster.
# Then load the Docker images into the Kubernetes cluster.
# This is necessary because the Kubernetes cluster does not have access to your local Docker images by default.
load-images: build-images
	@echo "Loading Docker images into the Kubernetes cluster $(CLUSTER_NAME)..."
	@minikube image load ping-service:latest -p $(CLUSTER_NAME)
	@minikube image load pong-service:latest -p $(CLUSTER_NAME)

# Starts all services if they are not already running. If they are already running, updates the services.
deploy: load-images
	@echo "Deploying the application via Helm..."
	@helm upgrade --install $(RELEASE_NAME) $(CHART_PATH) --namespace $(NAMESPACE) --create-namespace

# remove the entire services from the kubernetes cluster
# true => if the service is not found, it will not throw an error. it's necessary for the CI/CD pipeline workflow to not fail if the service is not found.
undeploy:
	@echo "Undeploying the application via Helm."
	@helm uninstall $(RELEASE_NAME) --namespace $(NAMESPACE) || true 