# HOWTO: Manual Steps for `feature/gke-base-deployment`

This document outlines manual steps required after running `terraform apply` for the `feature/gke-base-deployment` branch. These steps involve generating a service account key and configuring GitHub secrets and variables, which cannot be automated directly by Terraform or committed to the repository for security reasons.

---

### Step 1: Run `terraform apply`

Execute `terraform apply` in the `terraform/` directory. This will provision the GKE cluster, Cloud SQL instance, and the `github-actions-deployer` service account with the necessary IAM roles.

```bash
cd terraform/
terraform init
terraform apply
```

After successful application, Terraform will output the email of the newly created service account (e.g., `github_actions_deployer_email`).

---

### Step 2: Create the Service Account Key

Using the service account email obtained from `terraform output`, generate a JSON key file for the service account.

1.  **Retrieve Service Account Email:**
    ```bash
    export SA_EMAIL=$(terraform output -raw github_actions_deployer_email)
    echo "Service Account Email: $SA_EMAIL"
    ```

2.  **Generate Key File:**
    ```bash
    gcloud iam service-accounts keys create "gcp-sa-key.json" \
      --iam-account="$SA_EMAIL"
    ```
    This command will create a file named `gcp-sa-key.json` in your current directory.

---

### Step 3: Configure GitHub Secrets and Variables

You need to configure the following in your GitHub repository's settings (`Settings > Secrets and variables > Actions`):

#### GitHub Secrets

*   **`GCP_SA_KEY`**:
    *   **Value:** The **entire content** of the `gcp-sa-key.json` file generated in Step 2.
    *   **Purpose:** Allows GitHub Actions to authenticate with Google Cloud to push Docker images and deploy to GKE.

#### GitHub Variables

*   **`GCP_PROJECT_ID`**:
    *   **Value:** Your Google Cloud Project ID (e.g., `my-gcp-project-12345`).
    *   **Purpose:** Used by CI/CD to identify the target project.
*   **`GCR_HOSTNAME`**:
    *   **Value:** The hostname for your Google Artifact Registry (e.g., `us-central1-docker.pkg.dev` or `gcr.io` if using legacy GCR).
    *   **Purpose:** Specifies where to push Docker images.
*   **`GKE_CLUSTER_NAME`**:
    *   **Value:** The name of your GKE cluster (default: `todo-app-cluster`).
    *   **Purpose:** Used by CI/CD to authenticate and deploy to the correct GKE cluster.
*   **`GKE_CLUSTER_LOCATION`**:
    *   **Value:** The zone where your GKE cluster is located (default: `us-central1-a`).
    *   **Purpose:** Used by CI/CD to authenticate and deploy to the correct GKE cluster.

---

**Security Reminder:**
*   Treat your `gcp-sa-key.json` file as highly sensitive. Never commit it to git.
*   After securely configuring `GCP_SA_KEY` in GitHub, consider deleting the local `gcp-sa-key.json` file.
