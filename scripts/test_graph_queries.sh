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
kqlFiles=$(find internal/graph/azure-orphan-resources -type f -name "*.kql")

# Find all .kql files in the internal/graph/aprl/azure-resources directory and append them to kqlFiles
aprlKqlFiles=$(find internal/graph/aprl/azure-resources -type f -name "*.kql")

# Combine kqlFiles and aprlKqlFiles into a single variable
kqlFiles="$kqlFiles $aprlKqlFiles"

# Loop through each .kql file
for kqlFile in $kqlFiles; do
    echo "Processing file: $kqlFile"

    # Read the entire content of the .kql file
    kqlQuery=$(<"$kqlFile")

    # dot not attempt to run th equery if it contains 
    # "cannot-be-validated-with-arg, "under-development" or under development"
    if [[ "$kqlQuery" == *"cannot-be-validated-with-arg"* ]] || \
       [[ "$kqlQuery" == *"under-development"* ]] || \
       [[ "$kqlQuery" == *"under development"* ]]; then
        echo "Skipping query in $kqlFile due to manual validation or development status."
        continue
    fi

    # Run the KQL query using the Azure CLI
    runKqlQuery "$kqlQuery"
done
