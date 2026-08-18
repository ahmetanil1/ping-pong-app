#!/usr/bin/env bash

# if any error occurs, exit immediately and safely
set -euo pipefail

CLUSTER_NAME="hepapi-cluster"
MIN_CPUS=2
MIN_MEMORY=2048

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}==================================================${NC}"
echo -e "${GREEN}  HEPAPI Local Kubernetes Cluster Setup Script${NC}"
echo -e "${GREEN}==================================================${NC}"

# Check for required CLI dependencies
echo -e "${YELLOW}[1/4] Checking required CLI dependencies...${NC}"
for tool in docker minikube kubectl helm; do
    if ! command -v "$tool" &> /dev/null; then
        echo -e "${RED}ERROR: '$tool' is not installed or not in PATH.${NC}" >&2
        exit 1
    fi
done
echo -e "${GREEN}All required CLI tools are installed.${NC}"

# Verify Docker daemon status because minikube uses Docker as the driver for creating the cluster.
# If Docker is not running, the script will exit with an error message.
echo -e "${YELLOW}[2/4] Verifying docker daemon status...${NC}"
if ! docker info &> /dev/null; then
    echo -e "${RED} ERROR: Docker daemon is not running. Please start Docker and try again.${NC}" >&2
    exit 1
fi
echo -e "${GREEN}Docker daemon is running.${NC}"

# Prepare minikube cluster and check if it already exists.
echo -e "${YELLOW}[3/4] Preparing minikube cluster...${NC}"
if minikube status -p "$CLUSTER_NAME" &> /dev/null; then
    echo -e "${YELLOW}Minikube cluster '$CLUSTER_NAME' already exists.${NC}"
else
    minikube delete -p "$CLUSTER_NAME" &> /dev/null || true

    echo -e "${YELLOW}Creating minikube cluster '$CLUSTER_NAME'...${NC}"
    minikube start \
        -p "$CLUSTER_NAME" \
        --driver=docker \
        --nodes=2 \
        --cpus="$MIN_CPUS" \
        --memory="$MIN_MEMORY" \
        --kubernetes-version=v1.30.0
    echo -e "${GREEN}Minikube cluster '$CLUSTER_NAME' created successfully.${NC}"
fi

# Taints are used to differentiate between database and application workloads.
echo -e "${YELLOW}Configuring Database Node (${CLUSTER_NAME})...${NC}"
kubectl label node "$CLUSTER_NAME" workload=database --overwrite
echo -e "${YELLOW}Configuring Application Node (${CLUSTER_NAME}-m02)...${NC}"
kubectl label node "$CLUSTER_NAME-m02" workload=application --overwrite


# Enable required cluster addons for metrics-server, and storage provisioner.
echo -e "${YELLOW}[4/4] Enabling required cluster addons.${NC}"
minikube addons enable ingress -p "$CLUSTER_NAME" # Enable ingress controller for routing external traffic to services
minikube addons enable metrics-server -p "$CLUSTER_NAME" # Enable metrics-server for resource monitoring
minikube addons enable storage-provisioner -p "$CLUSTER_NAME" # Enable default storage provisioner for dynamic volume provisioning.

echo -e "${YELLOW}Waiting for Ingress deployment to be created in API server...${NC}"
# Ensure the Ingress deployment is created before applying the taint to avoid scheduling issues.
sleep 3
# Why we patch the Ingress deployment:
# If we don't patch it, the K8s scheduler might randomly place the Ingress pod on the Database Node before the taint is applied.
# This breaks database isolation, consumes DB node resources, and creates an unpredictable topology upon restarts.
# By patching with a nodeSelector, we strictly pin the Ingress to the Application Node, keeping the DB Node completely isolated.

# Since the taint is "NoSchedule" (which prevents new scheduling but doesn't evict existing pods), Ingress would stay there.
# if any pod already exists on the DB node, it will remain there until it is deleted or evicted. 
kubectl taint node "$CLUSTER_NAME" workload=database:NoSchedule --overwrite
kubectl config use-context "$CLUSTER_NAME" > /dev/null # Set the current context to the created cluster

echo -e "${GREEN}==================================================${NC}"
echo -e "${GREEN} Cluster Setup Completed Successfully!${NC}"
echo -e "${GREEN}==================================================${NC}"