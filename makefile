.PHONY: setup deploy undeploy # define the targets as phony to avoid conflicts with files of the same name


# if you want to run the cluster setup script, you can use the following command:
setup:
	@chmod +x scripts/cluster-setup.sh
	@./scripts/cluster-setup.sh

# Starts all services if they are not already running. If they are already running, updates the services.
deploy:
	@echo "Deploying the application via Helm..."
	@helm upgrade --install ping-pong-app ./deployments/helm/ping-pong-charts --create-namespace --namespace ping-pong

# remove the entire services from the kubernetes cluster
# true => if the service is not found, it will not throw an error. it's necessary for the CI/CD pipeline workflow to not fail if the service is not found.
undeploy:
	@echo "Undeploying the application via Helm."
	@helm uninstall ping-pong-app --namespace ping-pong || true 