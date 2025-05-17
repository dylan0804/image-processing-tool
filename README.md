# Image Processing Tool

This project is a Go-based API for performing image processing tasks. It includes a backend API, session management with Redis, and a full production-like deployment setup using Docker, Kubernetes, and an ELK stack (Elasticsearch, Logstash/Filebeat, Kibana) for centralized logging.

## Features

*   Image processing capabilities (e.g., sharpen, grayscale - *details to be added based on actual API features*).
*   RESTful API for interacting with the service.
*   Session management using Redis.
*   Containerized with Docker.
*   Orchestrated with Kubernetes for scalability and resilience.
*   Centralized logging using Filebeat, Elasticsearch, and Kibana.

## Architecture Overview

The system consists of several key components:

1.  **API Service (`image-processing-tool`):**
    *   A Go application providing HTTP endpoints for image processing.
    *   Connects to Redis for session storage.
    *   Outputs logs in JSON format to `stdout`.

2.  **Redis:**
    *   In-memory data store used for session management.

3.  **NGINX Ingress Controller (in Kubernetes):**
    *   Manages external access to the API and Kibana services.
    *   Routes requests based on hostnames (`api.example.com`, `kibana.example.com`).

4.  **Logging Stack (ELK/EFK - Elasticsearch, Filebeat, Kibana):**
    *   **Filebeat:** Deployed as a DaemonSet in Kubernetes. It runs on each node, collects container logs (from `stdout`/`stderr`), enriches them with Kubernetes metadata, and forwards them to Elasticsearch.
    *   **Elasticsearch:** A search and analytics engine that stores and indexes the logs sent by Filebeat.
    *   **Kibana:** A web UI for Elasticsearch that allows searching, viewing, and visualizing log data.

**Request Flow (Kubernetes):**
```
External User --> NGINX Ingress --> API Service (K8s) --> API Pod(s)
                                        |                   |
                                        |                   v
                                        |                 Redis (for sessions)
                                        |
                                        |--> Kibana Service (K8s) --> Kibana Pod (for UI)

Log Flow (Kubernetes):
API Pod(s) --(stdout/stderr)--> Kubelet/Docker on Node --(Filebeat Pod on Node)--> Elasticsearch Service --> Elasticsearch Pod(s)
                                                                                                                 ^
                                                                                                                 |
                                                                                                      Kibana (queries logs)
```

## Prerequisites

*   **Go:** Version 1.23 or higher (for building the application).
*   **Docker:** For building container images and running locally with Docker Compose.
*   **Docker Compose:** (Optional, for easy local development of all services).
*   **Minikube:** For running a local Kubernetes cluster.
*   **kubectl:** Kubernetes command-line tool.
*   **make:** (Optional, for using Makefile shortcuts).

## Local Development with Docker Compose

The `docker-compose.yml` file allows you to run the API, Redis, Elasticsearch, Kibana, and Filebeat locally.

1.  **Build the API image:**
    ```bash
    make build # Uses VERSION from Makefile, e.g., v1.0.0
    # Or manually: docker build -t image-processing-tool:your-tag .
    ```
    *Note: `docker-compose.yml` is set to build the image using the Dockerfile in the current directory, so a pre-built tag isn't strictly necessary for `docker-compose up` if it builds, but it's good practice.* 

2.  **Ensure your Go application in `cmd/api/main.go` and `internal/api/storage/redis_session_store.go` uses environment variables for Redis host/port with fallbacks to `localhost:6379` for local development.**
    The `docker-compose.yml` sets `REDIS_HOST=redis` for the API service.

3.  **Start all services:**
    ```bash
    docker-compose up -d
    ```

4.  **Access services:**
    *   **API:** `http://localhost:8080`
    *   **Redis:** (Internally accessible to the API on `redis:6379`)
    *   **Elasticsearch:** `http://localhost:9200`
    *   **Kibana:** `http://localhost:5601`

5.  **View logs in Kibana:**
    *   Open Kibana (`http://localhost:5601`).
    *   Set up an index pattern (e.g., `filebeat-*` or based on your `filebeat.yml` output configuration).
    *   Go to "Discover" to see logs from the API service.

6.  **Stop services:**
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
   *   `k8s/elasticsearch.yaml`
   *   `k8s/kibana.yaml` (including its Ingress)
   *   `k8s/api.yaml` (including its Ingress)
   *   It does **not** yet include the Filebeat logging components from `k8s/logging/`.

   **To deploy the Filebeat logging stack (after the main `make deploy`):**
   ```bash
   kubectl apply -f k8s/logging/filebeat-rbac.yaml
   kubectl apply -f k8s/logging/filebeat-config.yaml
   kubectl apply -f k8s/logging/filebeat-daemonset.yaml
   ```

**5. Verify Deployments:**
   ```bash
   make verify
   # Or kubectl get all -n image-processing-tool
   ```
   Check that all pods are `Running` and ready (e.g., `2/2` for deployments with 2 replicas).
   Check Filebeat pods: `kubectl get pods -n image-processing-tool -l k8s-app=filebeat`

**6. Access Services via Ingress:**
   a. **Start `minikube tunnel`:** In a **separate, dedicated terminal window**, run:
      ```bash
      minikube tunnel
      ```
      (Enter your macOS password if prompted. Leave this terminal running.)

   b. **Access the API:**
      Your API is configured to be accessible at `http://api.example.com`.
      With `minikube tunnel` running, this usually resolves to `127.0.0.1`.
      Use `curl` or Postman:
      ```bash
      curl -H "Host: api.example.com" http://127.0.0.1/your-api-endpoint
      # Example: curl -H "Host: api.example.com" http://127.0.0.1/health
      ```
      Alternatively, add `127.0.0.1 api.example.com` to your `/etc/hosts` file, then you can use `http://api.example.com` directly.

   c. **Access Kibana:**
      Kibana is configured to be accessible at `http://kibana.example.com`.
      Use your browser or `curl` as with the API:
      ```bash
      # Add to /etc/hosts: 127.0.0.1 kibana.example.com
      # Then open http://kibana.example.com in your browser.
      ```
      Or use `curl -H "Host: kibana.example.com" http://127.0.0.1`

**7. View Logs in Kibana:**
   *   Open Kibana.
   *   Navigate to Management -> Stack Management -> Kibana -> Index Patterns.
   *   Create an index pattern, likely `filebeat-*` (Filebeat typically creates daily indices like `filebeat-VERSION-YYYY.MM.DD`).
   *   Once the index pattern is created, go to "Discover" to view and search your application logs from the `api` pods.

## Makefile Targets

*   `make run`: Runs the Go API locally (without Docker).
*   `make test`: Runs Go tests.
*   `make build`: Builds the Docker image for the API, tagged with `$(VERSION)`.
*   `make load-to-minikube`: Loads the versioned Docker image into Minikube (depends on `build`).
*   `make deploy`: Deploys core application and service resources to Kubernetes (depends on `load-to-minikube`). *Does not deploy Filebeat logging stack automatically yet.*
*   `make verify`: Checks the status of all resources in the `image-processing-tool` namespace.

## Further Development

*   **Implement actual image processing endpoints** in the Go API.
*   Add more robust error handling and logging.
*   Enhance the `Makefile` to include deployment of the Filebeat logging stack.
*   Consider adding Prometheus and Grafana for metrics and monitoring.
*   Secure Elasticsearch and Kibana with authentication/authorization.
*   Implement CI/CD pipelines for automated building and deployment. 