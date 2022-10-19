[![build](https://github.com/cmendible/azqr/actions/workflows/build.yaml/badge.svg)](https://github.com/cmendible/azqr/actions/workflows/build.yaml)

# Azure Quick Review

Azure Quick Review (azqr) goal is to produce a high level assemesment of an Azure Subscription or Resource Group providing the following information for each Azure Service:

* SLA: current expected SLA
* Availability Zones: checks if the service is proytected against Zone failures. 
* Private Endpoints: checks if the service uses Private Endpoints.
* Diagnostic Settings: checks if there are Diagnostic Settings configured for the service. 
* CAF Naming convention: checks if the service follows CAF Naming convention.
