#!/bin/bash
# Script to initialize the Cloud SQL database schema
# This script connects to Cloud SQL via the Cloud SQL Proxy and runs init.sql

set -e

# Get the Cloud SQL instance connection name from Terraform output
INSTANCE_CONNECTION_NAME=$(cd terraform && terraform output -raw cloudsql_instance_connection_name)
DB_NAME="todoapp_db"
DB_USER="todoappuser"

echo "Connecting to Cloud SQL instance: $INSTANCE_CONNECTION_NAME"
echo "Database: $DB_NAME"
echo "User: $DB_USER"
echo ""

# Check if cloud-sql-proxy is installed
if ! command -v cloud-sql-proxy &> /dev/null; then
    echo "cloud-sql-proxy not found. Installing..."
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        curl -o cloud-sql-proxy https://storage.googleapis.com/cloud-sql-connectors/cloud-sql-proxy/v2.8.0/cloud-sql-proxy.darwin.amd64
        chmod +x cloud-sql-proxy
        sudo mv cloud-sql-proxy /usr/local/bin/
    else
        # Linux
        curl -o cloud-sql-proxy https://storage.googleapis.com/cloud-sql-connectors/cloud-sql-proxy/v2.8.0/cloud-sql-proxy.linux.amd64
        chmod +x cloud-sql-proxy
        sudo mv cloud-sql-proxy /usr/local/bin/
    fi
fi

# Start the Cloud SQL Proxy in the background
echo "Starting Cloud SQL Proxy..."
cloud-sql-proxy --port 5432 $INSTANCE_CONNECTION_NAME &
PROXY_PID=$!

# Wait for the proxy to be ready
sleep 3

# Prompt for the database password
echo ""
read -sp "Enter database password for user '$DB_USER': " DB_PASSWORD
echo ""

# Run the init.sql script
echo "Running init.sql..."
PGPASSWORD=$DB_PASSWORD psql -h 127.0.0.1 -p 5432 -U $DB_USER -d $DB_NAME -f init.sql

echo ""
echo "Database schema initialized successfully!"

# Kill the Cloud SQL Proxy
kill $PROXY_PID

echo "Cloud SQL Proxy stopped."
