<!-- Written by Gemini CLI -->
<!-- This file is licensed under the MIT License. See the LICENSE file for details. -->

# Go To-Do Application

A simple To-Do list application built with Go and PostgreSQL. This application allows users to create, view, update, and delete to-do items.

## Features

*   **Create To-Do Items**: Add new tasks to your list.
*   **View To-Do Items**: See all your pending and completed tasks.
*   **Update To-Do Items**: Mark tasks as completed.
*   **Delete To-Do Items**: Remove tasks from your list.
*   **Persistent Storage**: To-Do items are stored in a PostgreSQL database.
*   **High Availability**: Regional GKE cluster with read replicas
*   **Resilience**: Automatic retries and circuit breakers for fault tolerance
*   **Observability**: 
    - Prometheus metrics for HTTP requests and latency
    - Business metrics (todos added, updated, deleted)
    - Cloud Monitoring dashboards and SLO tracking
    - Distributed tracing with Cloud Trace

## Architecture

### Resilience Features (99.9% Availability)
- **Exponential Backoff Retries**: All database operations retry automatically on transient failures
- **Circuit Breaker**: Prevents cascading failures when database is consistently unavailable
- **Read Replica**: Read traffic distributed to replica for better performance and availability
- **Point-in-Time Recovery**: Database backups with PITR enabled

### Observability
- **Metrics**: Prometheus metrics exported at `/metrics` endpoint
  - HTTP request count and duration by endpoint
  - Business metrics: `todos_added_total`, `todos_updated_total`, `todos_deleted_total`
- **Tracing**: OpenTelemetry integration with Cloud Trace
  - Distributed tracing for all HTTP requests
  - Database query performance tracking
- **SLOs**: Availability (99.9%) and Latency (95% < 500ms) tracking
- **Dashboards**: Custom Cloud Monitoring dashboard with system overview

### Infrastructure
- **Compute**: Regional GKE cluster in `us-central1` (multi-zone) with dedicated namespaces
- **Database**: Cloud SQL PostgreSQL with HA configuration and read replica
- **Deployment**: Cloud Deploy with automated canary releases and verification
- **Monitoring**: Google Cloud Monitoring + Managed Prometheus
- **Security**: Workload Identity, IAM authentication, Content Security Policy
- **Backup**: GKE Backup for GKE enabled for cluster data and persistent volumes
- **Resource Management**: Configured resource requests and limits for pods

## Prerequisites

Before you begin, ensure you have the following installed:

*   [Go](https://go.dev/doc/install) (version 1.22.x or later)
*   [Docker](https://docs.docker.com/get-docker/) (for building container images)
*   [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) (for interacting with Kubernetes clusters)
*   [Google Cloud SDK (`gcloud`)](https://cloud.google.com/sdk/docs/install) (for interacting with Google Cloud)
*   [skaffold](https://skaffold.dev/docs/install/) (for local development and CI/CD with GKE)

## Local Development

Follow these steps to get the application up and running on your local machine for development.

### 1. Clone the Repository

```bash
git clone https://github.com/gemini/todo-app.git
cd todo-app
```

### 2. Build and Run the Go Application

To build and run the Go application directly:

```bash
go build -o main .
./main
```

The application will start on `http://localhost:8080`.

### 3. Running with Skaffold (Local Kubernetes Simulation)

You can use Skaffold to run the application locally within a Kubernetes-like environment (e.g., Minikube or Docker Desktop's Kubernetes). This builds the Docker image, deploys to your local Kubernetes context, and streams logs.

```bash
skaffold dev
```

### 4. Access the Application

Once the application is running (either directly or via Skaffold), you can access it in your web browser at:

[http://localhost:8080](http://localhost:8080)

## Deployment to GKE

The application is deployed to Google Kubernetes Engine (GKE) using Cloud Deploy. The CI/CD pipeline is configured via GitHub Actions.

### 1. Configure Google Cloud Project

Ensure your Google Cloud project is set up and authenticated with `gcloud`.

```bash
gcloud auth login
gcloud config set project [YOUR_PROJECT_ID]
```

### 2. Enable Required APIs

Enable the necessary Google Cloud APIs:

```bash
gcloud services enable \
    artifactregistry.googleapis.com \
    clouddeploy.googleapis.com \
    container.googleapis.com \
    cloudbuild.googleapis.com \
    iam.googleapis.com \
    secretmanager.googleapis.com \
    sqladmin.googleapis.com \
    cloudtrace.googleapis.com \
    monitoring.googleapis.com \
    cloudresourcemanager.googleapis.com \
    gke-backup.googleapis.com
```

### 3. Apply Terraform Configuration

The infrastructure (GKE cluster, Cloud SQL, IAM, etc.) is managed using Terraform.

```bash
cd terraform
terraform init
terraform apply
cd ..
```

### 4. Continuous Deployment via GitHub Actions

The application is continuously deployed to GKE via GitHub Actions. Pushing changes to the `main` branch or a `feature/` branch will trigger the CI/CD pipeline:

*   **Build**: Builds the Go application and Docker image.
*   **Security Scan**: Runs `gosec` and `Trivy` to check for vulnerabilities.
*   **Deploy**: Creates a Cloud Deploy release, which handles canary deployment to the GKE cluster.

### 5. Manual Release Creation

You can also manually trigger a release from the GitHub Actions UI or using the `gh` CLI:

```bash
gh workflow run .github/workflows/build-test.yml --ref [YOUR_BRANCH_NAME]
```

### 6. Monitor Deployment

Monitor your deployments through the Google Cloud Deploy console:

[Google Cloud Deploy Console](https://console.cloud.google.com/deploy)

## API Endpoints

The application exposes the following API endpoints:

*   **`GET /todos`**: Retrieve all to-do items.
*   **`POST /todos`**: Add a new to-do item.
    *   Request Body: `{"task": "New task description"}`
*   **`PUT /todos/{id}`**: Update a to-do item (e.g., mark as completed).
    *   Request Body: `{"completed": true}`
*   **`DELETE /todos/{id}`**: Delete a to-do item.
