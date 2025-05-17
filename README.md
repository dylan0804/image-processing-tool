# Image Processing Tool

This project is a simple Go-based API for performing image processing tasks. It includes a backend API, session management with Redis, and a full production-like deployment setup using Docker and Kubernetes

## Features

*   Image processing capabilities (e.g., upload, blur, and sharpen)
*   RESTful API for interacting with the service.
*   Session management using Redis.
*   Containerized with Docker.
*   Orchestrated with Kubernetes for scalability

## Architecture Overview

The system consists of several key components:

1.  **API Service (`image-processing-tool`):**
    *   A Go application providing HTTP endpoints for image processing.
    *   Connects to Redis for session storage.
    *   Outputs logs in JSON format to `stdout`.

2.  **Redis:**
    *   In-memory data store used for image session management.

3.  **NGINX Ingress Controller (in Kubernetes):**
    *   Manages external access to the API.
    *   Routes requests based on hostnames (`api.example.com`).

## Prerequisites

*   **Go:** Version 1.24 or higher (for building the application).
*   **Docker:** For building container images and running locally with Docker Compose.
*   **Docker Compose:** (Optional, for easy local development of all services).
*   **Minikube:** For running a local Kubernetes cluster.
*   **kubectl:** Kubernetes command-line tool.
*   **make:** (Optional, for using Makefile shortcuts).

## Local Development with Docker Compose

The `docker-compose.yml` file allows you to run the API and Redis locally.

1.  **Build the API image:**
    ```bash
    make build # Uses VERSION from Makefile, e.g., v1.0.0
    # Or manually: docker build -t image-processing-tool:your-tag .
    ```

2.  **Ensure your Go application in `cmd/api/main.go` and `internal/api/storage/redis_session_store.go` uses environment variables for Redis host/port with fallbacks to `localhost:6379` for local development.**
    The `docker-compose.yml` sets `REDIS_HOST=redis` for the API service.

3.  **Start all services:**
    ```bash
    docker-compose up -d
    ```

4.  **Access services:**
    *   **API:** `http://localhost:8080`
    *   **Redis:** (Internally accessible to the API on `redis:6379`)

5.  **Stop services:**
    ```bash
    docker-compose down
    ```

## Deployment to Kubernetes (Minikube)

Follow these steps to deploy the application and its services to a Minikube cluster.

**1. Start Minikube:**
   ```bash
   minikube start
   ```

**2. Enable Minikube Addons:**
   Enable the Ingress controller:
   ```bash
   minikube addons enable ingress
   ```

**3. Set Version and Build/Load Image:**
   a. Decide on a version (e.g., `v1.0.0`).
   b. Update `VERSION := vX.Y.Z` in your `Makefile`.
   c. Update `image: image-processing-tool:vX.Y.Z` in `k8s/api.yaml`.
   d. Build the image and load it into Minikube:
      ```bash
      make build        # Builds image-processing-tool:$(VERSION)
      make load-to-minikube # Loads it into Minikube
      ```

**4. Deploy All Kubernetes Resources:**
   The `Makefile` provides a convenient target. This assumes you have updated the versions in `Makefile` and `k8s/api.yaml` as per step 3.
   ```bash
   make deploy
   ```
   This will apply:
   *   `k8s/namespace.yaml`
   *   `k8s/redis.yaml`
   *   `k8s/api.yaml` (including its Ingress)

**5. Verify Deployments:**
   ```bash
   make verify
   # Or kubectl get all -n image-processing-tool
   ```
   Check that all pods are `Running` and ready (e.g., `2/2` for deployments with 2 replicas).

**6. Access Services via Ingress:**
   a. **Start `minikube tunnel`:** In a **separate terminal window**, run:
      ```bash
      minikube tunnel
      ```
      (Enter your password if prompted. Leave this terminal running.)

   b. **Access the API:**
      Your API is configured to be accessible at `http://api.example.com`.
      With `minikube tunnel` running, this usually resolves to `127.0.0.1`.
      Use `curl` or Postman:
      ```bash
      curl -H "Host: api.example.com" http://127.0.0.1/your-api-endpoint
      # Example: curl -H "Host: api.example.com" http://127.0.0.1/health
      ```

## Makefile Targets

*   `make run`: Runs the Go API locally (without Docker).
*   `make test`: Runs Go tests.
*   `make build`: Builds the Docker image for the API, tagged with `$(VERSION)`.
*   `make load-to-minikube`: Loads the versioned Docker image into Minikube (depends on `build`).
*   `make deploy`: Deploys core application and service resources to Kubernetes (depends on `load-to-minikube`). *Does not deploy Filebeat logging stack automatically yet.*
*   `make verify`: Checks the status of all resources in the `image-processing-tool` namespace.
