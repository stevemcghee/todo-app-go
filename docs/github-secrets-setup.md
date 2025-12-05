# Database Initialization Setup

## GitHub Secrets and Variables Required

To enable automated database initialization via the Kubernetes Job, you need to add the following to your GitHub repository:

### Secrets (Settings → Secrets and variables → Actions → Secrets)

1. **DB_PASSWORD**: The password for the `todoappuser` database user
   - Value: (the password you set in `terraform/terraform.tfvars`)

### Variables (Settings → Secrets and variables → Actions → Variables)

1. **DB_USER**: `todoappuser`
2. **DB_NAME**: `todoapp_db`

## How to Add Them

### Via GitHub Web UI:
1. Go to your repository on GitHub
2. Click **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret** to add `DB_PASSWORD`
4. Click the **Variables** tab → **New repository variable** to add `DB_USER` and `DB_NAME`

### Via GitHub CLI:
```bash
# Add secret
gh secret set DB_PASSWORD --body "your-db-password-here"

# Add variables
gh variable set DB_USER --body "todoappuser"
gh variable set DB_NAME --body "todoapp_db"
```

## How It Works

Once these are configured:

1. The CI/CD pipeline will automatically substitute the placeholders in `k8s/db-init-job.yaml`
2. The Kubernetes Job will run and create the `todos` table if it doesn't exist
3. The application pods will start successfully

## Manual Initialization (Alternative)

If you prefer to initialize the database manually first, you can use the script:

```bash
./scripts/init-db.sh
```

This is useful for:
- Initial setup before configuring GitHub secrets
- Local development
- Troubleshooting
