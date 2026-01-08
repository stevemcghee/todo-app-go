# Multi-Region Expansion Plan (`milestone-13-multi-region`)

This document outlines the detailed plan to expand the `todo-app-go` implementation from a single region (`us-central1`) to a multi-region architecture (adding `us-east1`).

## Goals
- **Availability**: Increase SLI from 99.9% to 99.99%.
- **Resilience**: Survive a complete regional outage of `us-central1`.
- **Latency**: Serve users from the region closest to them.

## Architecture Overview
- **Compute**: Two GKE clusters, identical configuration.
  - Primary: `us-central1` (existing)
  - Secondary: `us-east1` (new)
- **Database**: Cloud SQL with Cross-Region Read Replica.
  - Primary: `us-central1`
  - Replica: `us-east1` (PROMOTION capable)
- **Ingress**: Global External Application Load Balancer (GCLB) with Multi-Cluster Ingress (MCI) or Multi-Cluster Gateway (MCG).
- **GitOps**: ArgoCD managing both clusters.

## Implementation Steps

### Phase 1: Preparation & Networking
- [ ] **Quota Check**: Ensure `us-east1` has sufficient CPU/IP quotas.
- [ ] **VPC Updates**: Ensure subnets exist for `us-east1` GKE and Services.
- [ ] **Terraform Refactor**: Refactor Terraform to support multi-region modules (DRY principle).

### Phase 2: Database Expansion
- [ ] **Create Replica**: Terraform changes to add `us-east1` Read Replica.
- [ ] **Verify Replication**: Check replication lag and connectivity.
- [ ] **Application Config**: Update app to be aware of read-replicas (optional optimization) or just ensure it connects to the local region's database endpoint (using Cloud SQL Proxy or internal DNS).
    - *Note*: If the app is write-heavy, writes MUST go to Primary. If using Cloud SQL Proxy, we need to ensure the proxy in `us-east1` points to the Primary in `us-central1` for writes, or we assume `us-east1` is read-only until failover.
    - *Decision*: For simplicity initially, `us-east1` app instances will connect to `us-central1` Primary for writes.

### Phase 3: Compute Replication
- [ ] **Deploy GKE Cluster**: Provision `todo-cluster-east` in `us-east1`.
- [ ] **Workload Identity**: Replicate creation of ServiceAccounts and IAM bindings.
- [ ] **ArgoCD Registration**: Register the new cluster with the existing ArgoCD (hub-and-spoke or just multi-target).
- [ ] **Deploy App**: Sync applications to the new cluster.

### Phase 4: Application Deployment & Traffic Management
- [x] **ArgoCD Registration**: Register the new cluster (Done).
- [ ] **ArgoCD Application**:
    - Update `argocd-todo-app.yaml` to include a second Application for `us-east1`.
    - Commit and sync.
- [ ] **Ingress Strategy**:
    - Current `ingress.yaml` creates a *Regional* External Load Balancer in each cluster.
    - We will have two separate IP addresses (one per region).
    - **Future/Next**: We will implement specific Multi-Cluster Ingress (MCI) or simply use DNS Geo-Routing (e.g. Google Cloud DNS with Geo Policy) to route users to the closest IP.
    - *Decision*: For this milestone, we accept two regional load balancers.
- [ ] **DNS Update**:
    - Add A records for both regional IPs (Round Robin) or use Geo DNS.

### Phase 5: Verification & Drills
- [ ] **Traffic Distribution**: Verify traffic is routed to the closest region.
- [ ] **Failover Drill**:
    1.  Simulate `us-central1` outage (drain traffic).
    2.  Verify `us-east1` handles load.
    3.  (Advanced) Promote `us-east1` DB to primary and verify write capability.

## Risks & Mitigations
- **Data Consistency**: Cross-region replication has latency. Strong consistency for writes is maintained by always writing to primary, but reads from replica might be stale.
- **Cost**: Doubling the infrastructure will double the compute/DB costs.
- **Complexity**: Debugging distributed systems is harder.

## Potential Pitfalls and Challenges Observed
During the implementation of Milestone 13, several challenges were encountered:

1.  **Binary Authorization Pattern Specificity**:
    *   **Challenge**: Gatekeeper and application images were blocked on the new cluster despite generic whitelist patterns. Patterns like `docker.io/openpolicyagent/*` failed to match when the Kubernetes event reported the image as `openpolicyagent/gatekeeper` (omitting the registry).
    *   **Solution**: Updated `binauthz-policy.yaml` to include explicit patterns matching both fully-qualified and short-name variants (e.g., `openpolicyagent/gatekeeper:*`).

2.  **Gatekeeper Installation Timeouts**:
    *   **Challenge**: Terraform's `helm_release` for Gatekeeper repeatedly timed out during the "pre-install" hook (CRD update job). This was caused by the hook job being blocked by the Binary Authorization issue mentioned above, leading to a "zombie" release state.
    *   **Solution**: Performed a manual deep cleanup of the `gatekeeper-system` namespace, manually installed the release to verify pod health, and then imported the working release into Terraform state. Set `wait = false` in `terraform/gatekeeper.tf` to prevent fragile timeout logic from breaking future applies.

3.  **ArgoCD Sync Chicken-and-Egg (CRDs)**:
    *   **Challenge**: ArgoCD failed to sync Gatekeeper `Constraints` because the `ConstraintTemplates` (which define the CRDs for those constraints) hadn't been processed by Gatekeeper yet.
    *   **Solution**: Manually seeded the `ConstraintTemplates` in the new cluster using `kubectl apply` to establish the CRDs, allowing ArgoCD to successfully sync the remaining resources in subsequent retries.
