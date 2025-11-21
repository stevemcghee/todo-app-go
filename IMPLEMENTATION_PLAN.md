# Implementation Plan for Production Readiness

This document outlines the branching strategy and step-by-step implementation plan for making the `todo-app-go` application a production-ready, resilient, and secure service.

The strategy is to use a series of feature branches, where each branch represents a major architectural evolution of the system. This allows for clear, incremental progress that can be reviewed and merged via pull requests.

---

### 1. Branch: `2-gke-cicd-base` (Completed)

*   **Goal:** Get the application running on a basic GKE cluster with a managed database. This is the foundational step to move away from the local Docker environment.
*   **Tasks:**
    1.  [x] Update Terraform scripts to provision a basic GKE cluster and a single-zone Cloud SQL instance.
    2.  [x] Create the initial Kubernetes manifests (`deployment.yaml`, `service.yaml`) for the todo-app.
    3.  [x] Update the GitHub Actions CI/CD pipeline to:
        *   Build and publish the Docker image to Google Artifact Registry.
        *   Authenticate to GKE and apply the Kubernetes manifests.
    4.  [x] Implement Database Initialization:
        *   Create a Kubernetes Job (`k8s/db-init-job.yaml`) to initialize the schema using Cloud SQL Proxy.
        *   Update CI/CD to handle secret substitution and job cleanup.
        *   Document secret setup (`docs/github-secrets-setup.md`).
*   **Risks Addressed:** This branch lays the groundwork but doesn't fully mitigate the major risks yet. It serves as a "walking skeleton" for the production environment.

---

### 2. Branch: `3-ha-scalability` (Current Focus)

*   **Goal:** Make the base deployment highly available and scalable.
*   **Tasks:**
    1.  Upgrade the GKE cluster in Terraform to be a *regional* cluster (spanning multiple zones).
    2.  Upgrade the Cloud SQL instance in Terraform to use the High Availability (HA) configuration.
    3.  Implement a Horizontal Pod Autoscaler (HPA) in Kubernetes to automatically scale the application.
*   **Risks Addressed:**
    *   ✅ Single Point of Failure (Application & Database)
    *   ✅ Lack of Scalability
    *   ✅ Zonal Failure

---

### 3. Branch: `4-security-observability` (Branches from `3-ha-scalability`)

*   **Goal:** Protect the application from external threats and improve monitoring.
*   **Tasks:**
    1.  Implement Workload Identity and use Secret Manager for database credentials, removing them from environment variables.
    2.  Update the Go application to fetch credentials from Secret Manager.
    3.  Use Terraform to provision a Google Cloud Load Balancer with Google Cloud Armor (WAF) to protect against DDoS and other attacks.
    4.  Implement a Content Security Policy (CSP) in the `index.html` template.
    5.  Integrate a security scanner (like `gosec` or `Trivy`) into the CI/CD pipeline.
*   **Risks Addressed:**
    *   ✅ Insecure and Inflexible Configuration
    *   ✅ DDoS attacks or other security concerns

---

### 3.5. Branch: `4.5-advanced-deployments` (Branches from `4-security-observability`)

*   **Goal:** Implement multi-stage CI/CD and progressive delivery strategies for safer, more controlled deployments.
*   **Tasks:**
    1.  Create a staging environment (separate GKE namespace or cluster) in Terraform.
    2.  Update CI/CD pipeline to deploy to staging first, then production.
    3.  Implement canary deployments using Flagger or native Kubernetes strategies.
    4.  Add deployment approval gates for production deployments.
    5.  Implement automated rollback on deployment failures or error rate spikes.
    6.  Add smoke tests and integration tests that run post-deployment.
*   **Risks Addressed:**
    *   ✅ Deployment-induced outages
    *   ✅ Inability to quickly rollback bad deployments
    *   ✅ Lack of deployment confidence and testing

---

### 4. Branch: `5-disaster-recovery` (Branches from `4.5-advanced-deployments`)

*   **Goal:** Prepare for a full regional outage with multi-region deployment.
*   **Tasks:**
    1.  Update Terraform to be able to replicate the entire GKE and Cloud SQL setup in a second region.
    2.  Configure the Cloud Load Balancer to manage traffic between the two regions, failing over if one region becomes unhealthy.
    3.  Configure cross-region replication for the Cloud SQL instance.
*   **Risks Addressed:**
    *   ✅ Compute Service or Regional Failure within GCP
