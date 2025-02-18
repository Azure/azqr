#!/bin/bash

# Function to run KQL queries using Azure CLI
# Arguments:
#   $1 - The KQL query to run
runKqlQuery() {
  local kqlQuery="$1"
  echo "Running query: $kqlQuery"
  # Run the KQL query using the Azure CLI and stop on error
  az graph query -q "$kqlQuery" || { echo "Error running query: $kqlQuery"; exit 1; }
}

# Find all .kql files in the specified directory
kqlFiles=$(find internal/azure-orphan-resources -type f -name "*.kql")

# Loop through each .kql file
for kqlFile in $kqlFiles; do
    echo "Processing file: $kqlFile"

    # Read the entire content of the .kql file
    kqlQuery=$(<"$kqlFile")

    # Run the KQL query using the Azure CLI
    runKqlQuery "$kqlQuery"
done
