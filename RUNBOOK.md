# Maintenance Runbook

## Rollback Procedures

### Cloud Deploy Rollback (Preferred)
If a bad release is detected, use Cloud Deploy to rollback to the previous release.

1. **Identify the previous release**:
   ```bash
   gcloud deploy releases list --delivery-pipeline=todo-app-pipeline --region=us-central1
   ```
2. **Promote the previous release**:
   ```bash
   gcloud deploy releases promote --release=[PREVIOUS_RELEASE_NAME] \
     --delivery-pipeline=todo-app-pipeline \
     --region=us-central1 \
     --to-target=production
   ```

### Manual Rollback (Emergency)
If Cloud Deploy is unavailable, manually apply the previous Kubernetes manifests.

1. **Checkout the previous stable commit**:
   ```bash
   git checkout [PREVIOUS_COMMIT_SHA]
   ```
2. **Apply manifests**:
   ```bash
   kubectl apply -f k8s/
   ```

## Application Resilience Features

### Automatic Retries
The application implements exponential backoff retries for all database operations.

**Configuration**:
- Initial interval: 100ms
- Max interval: 2s
- Max elapsed time: 5s

**Behavior**: Transient database errors (network blips, connection pool exhaustion) are automatically retried. Check logs for retry warnings:
```bash
kubectl logs -l app=todo-app-go | grep "retrying"
```

### Circuit Breaker
A circuit breaker protects against cascading failures when the database is consistently unavailable.

**States**:
- **Closed**: Normal operation, all requests pass through
- **Open**: After 60% failure rate (min 3 requests), requests fail immediately with `503 Service Unavailable`
- **Half-Open**: After 30s, allows 1 request to test if service recovered

**Monitoring**:
Check circuit breaker state changes:
```bash
kubectl logs -l app=todo-app-go | grep "Circuit Breaker state changed"
```

**Recovery**: Circuit breaker auto-recovers when database becomes healthy. No manual intervention needed.

### Read Replica
Read queries (`GET /todos`) are automatically routed to a read replica for improved performance and availability.

**Failover**: If read replica is unavailable, application falls back to primary database automatically.

**Verify Connection**:
```bash
# Check both connections are active
kubectl logs -l app=todo-app-go | grep "Successfully connected"
# Should see: "Successfully connected to PRIMARY database"
# AND: "Successfully connected to READ REPLICA"
```

## Troubleshooting

### Database Connectivity Issues
**Symptoms**: HTTP 500 errors, "password authentication failed" logs.

1. **Check Cloud SQL Proxy**:
   ```bash
   kubectl logs -l app=todo-app-go -c cloudsql-proxy
   ```
2. **Verify Workload Identity**:
   Ensure the Kubernetes ServiceAccount is annotated correctly:
   ```bash
   kubectl describe sa todo-app-sa
   ```
3. **Check IAM Permissions**:
   Ensure the Google Service Account has `roles/cloudsql.instanceUser`.

### High Load / Scaling Issues
**Symptoms**: High latency, HPA maxed out.

1. **Check HPA Status**:
   ```bash
   kubectl get hpa
   ```
2. **Increase Max Replicas** (if needed):
   Edit `k8s/hpa.yaml` and increase `maxReplicas`.
   ```bash
   kubectl apply -f k8s/hpa.yaml
   ```
3. **Check Database Load**:
   Check Cloud SQL CPU utilization in Cloud Console. If high, consider upgrading the instance tier (requires downtime).

## Disaster Recovery

### Cluster Failure Scenarios

#### Zone Failure
**Risk**: One zone in `us-central1` becomes unavailable.
**Mitigation**: We use a **Regional GKE Cluster**. The control plane is replicated across zones, and nodes are distributed.
**Action**: Kubernetes will automatically reschedule pods to healthy zones. No manual intervention required, but capacity might be reduced.

#### Region Failure
**Risk**: The entire `us-central1` region becomes unavailable.
**Mitigation**: Currently **UNMITIGATED**. The application resides only in `us-central1`.
**Recovery**:
1.  Spin up infrastructure in a new region (e.g., `us-east1`) using Terraform (update `region` variable).
2.  Restore Cloud SQL database from backup to the new region (Cross-Region Restore).
3.  Update DNS to point to the new Load Balancer IP.

### Database Restore

#### Point-in-Time Recovery (PITR)
To restore the database to a specific timestamp (e.g., before an accidental deletion):

1.  **Identify Timestamp**: Determine the exact time of the incident (in RFC 3339 format, e.g., `2025-12-03T12:00:00Z`).
2.  **Clone Instance**: Create a new instance from the backup (safer than overwriting).
    ```bash
    gcloud sql instances clone todo-app-db-instance todo-app-db-recovered \
      --point-in-time="2025-12-03T12:00:00Z"
    ```
3.  **Verify Data**: Connect to `todo-app-db-recovered` and verify the data.
4.  **Promote**: Update the application to use the new instance IP/connection name.

#### Full Backup Restore
To restore from a specific daily backup (overwrites current data):

1.  **List Backups**:
    ```bash
    gcloud sql backups list --instance=todo-app-db-instance
    ```
2.  **Restore**:
    ```bash
    gcloud sql backups restore [BACKUP_ID] --restore-instance=todo-app-db-instance
    ```
    *Warning: This will overwrite the current database state.*
