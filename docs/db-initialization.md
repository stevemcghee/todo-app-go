# Database Initialization - IaC Approaches

We have several options for initializing the Cloud SQL database schema in an Infrastructure-as-Code way:

## Option 1: Kubernetes Job (Recommended) ✅

**File:** `k8s/db-init-job.yaml`

A Kubernetes Job that runs once to initialize the schema. This is the most cloud-native approach.

**Pros:**
- Fully declarative and version-controlled
- Runs in the same environment as the app
- Automatically uses Cloud SQL Proxy
- Can be applied via CI/CD pipeline
- Idempotent (uses `CREATE TABLE IF NOT EXISTS`)

**Cons:**
- Requires database credentials to be available (currently hardcoded placeholders)
- Will need to be updated when we implement Secret Manager

**To apply manually:**
```bash
# First, update the placeholders in k8s/db-init-job.yaml with actual values
# Then apply:
kubectl apply -f k8s/db-init-job.yaml

# Check job status:
kubectl get jobs
kubectl logs job/db-init
```

**To apply via CI/CD:**
The job is already included in `k8s/` directory, so `kubectl apply -f k8s/` will create it.

---

## Option 2: Terraform null_resource with local-exec

Add a `null_resource` to Terraform that runs the SQL initialization after the database is created.

**Pros:**
- Runs as part of `terraform apply`
- Guaranteed to run after database creation
- Can use Terraform variables

**Cons:**
- Requires local PostgreSQL client (`psql`)
- Requires Cloud SQL Proxy to be installed locally
- Not truly declarative (uses local-exec)
- Harder to debug

**Example (not implemented):**
```hcl
resource "null_resource" "init_db" {
  depends_on = [google_sql_database.database]
  
  provisioner "local-exec" {
    command = <<-EOT
      cloud-sql-proxy ${google_sql_database_instance.main_instance.connection_name} &
      PROXY_PID=$!
      sleep 3
      PGPASSWORD=${var.db_password} psql -h 127.0.0.1 -U ${var.db_user} -d ${var.db_database_name} -f ../init.sql
      kill $PROXY_PID
    EOT
  }
}
```

---

## Option 3: Application-Level Migration (Future Enhancement)

Use a database migration tool like `golang-migrate` or `goose` in the Go application.

**Pros:**
- Application owns its schema
- Supports schema versioning and rollbacks
- No manual intervention needed

**Cons:**
- Requires code changes
- More complex setup

---

## Recommendation

Use **Option 1 (Kubernetes Job)** for now. It's the most Kubernetes-native approach and will work well with your existing setup. Once we implement Secret Manager in the HA Scalability Hardening phase, we'll update the job to pull credentials from Secret Manager instead of using placeholders.

For immediate use, you can manually apply the job with the correct credentials, or we can update the CI/CD pipeline to substitute the placeholders just like we do for the deployment.
