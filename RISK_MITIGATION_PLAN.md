# Risk Mitigation Plan for todo-app-go

This document outlines identified infrastructure and dependency risks for the `todo-app-go` project and proposes a mitigation plan to enhance its reliability, scalability, and observability.

## Mitigations in Place

1.  **No Automated Health Checks or Recovery:**
    *   **Mitigation:** An HTTP `/healthz` endpoint has been added to the Go application. This endpoint checks the database connection and returns a 200 OK status if the database is reachable. The `docker-compose.yml` file has been updated to use this health check, allowing Docker to automatically restart unhealthy containers.

2.  **Containerization for Production:**
    *   **Mitigation:** The `Dockerfile` has been improved for production use. It now uses a multi-stage build to create a small final image based on `alpine:latest`. The application is run as a non-root user (`appuser`) to enhance security. `curl` is also installed in the final image to support the health check.

3.  **Manual and Error-Prone Deployments (CI/CD):**
    *   **Mitigation:** A basic CI/CD pipeline has been created using GitHub Actions (`.github/workflows/build-test.yml`). This pipeline automatically builds and tests the application on every push or pull request to the `main` branch. This reduces the risk of manual deployment errors and ensures a consistent build process.

4.  **Observability Failures:**
    *   **Mitigation:** Observability has been enhanced by:
        *   **Structured Logging:** The application now uses the `log/slog` library for structured JSON logging, making logs easier to parse and analyze.
        *   **Metrics Endpoint:** A `/metrics` endpoint has been added to expose Prometheus-compatible metrics, providing visibility into the application's performance and health.

## Remaining Risks

1.  **Single Point of Failure (SPOF) for Application:** The application still runs as a single container in the local Docker environment. If this container crashes, the service will be unavailable.
    *   **Proposed Mitigation:** Deploy the application to a container orchestration platform like Kubernetes, which can manage multiple replicas of the application and automatically restart failed instances.

2.  **Single Point of Failure (SPOF) for Database:** The database still runs as a single container. If the database fails, the application will be unable to read or write data.
    *   **Proposed Mitigation:** Use a managed, high-availability database service (like Google Cloud SQL or Amazon RDS) with automated failover and replication.

3.  **Data Loss and Durability:** The database relies on a local Docker volume. This data is not backed up, is not replicated, and could be lost if the host machine fails.
    *   **Proposed Mitigation:** Use a managed database service with automated backups and point-in-time recovery.

4.  **Lack of Scalability:** The current single-instance architecture cannot handle any significant load.
    *   **Proposed Mitigation:** Deploy the application to a container orchestrator (like Kubernetes) that can automatically scale the number of application instances based on load.

5.  **Zonal and Regional Failure:** The entire application stack runs on a single machine, making it completely vulnerable to failures of that machine or its location.
    *   **Proposed Mitigation:** Deploy the application and database across multiple availability zones within a region for zonal redundancy. For higher availability, consider a multi-region deployment.

6.  **Insecure and Inflexible Configuration:** Database credentials are still managed via environment variables and `.env` files, which is not secure for production.
    *   **Proposed Mitigation:** Use a dedicated secrets management solution like Google Secret Manager or HashiCorp Vault.

These remaining risks are addressed by the "Infrastructure as Code (IaC) for Cloud Deployment" and "Deploy to the Cloud" steps in the original mitigation plan.

