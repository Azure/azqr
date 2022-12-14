### Container Instances

> **Note**
> This feature is currently not available for Azure portal

> **Warning**
> This feature is currently in preview

Use of Availability Zones is recommended. The Azure Container Instances (ACI) supports zonal container group deployments, meaning the instance is pinned to a specific, self-selected availability zone. The availability zone is specified at the container group level. Containers within a container group can't have unique availability zones.

[Deploy an Azure Container Instances (ACI) container group in an availability zone (preview) - Microsoft Learn](https://learn.microsoft.com/en-us/azure/container-instances/availability-zones)

The following container groups don't support availability zones, and don't offer any migration guidance:

- Container groups with GPU resources
- Virtual Network injected container groups
- Windows Server 2016 container groups

[Migrate Azure Container Instances to availability zone support - Microsoft Learn](https://learn.microsoft.com/en-us/azure/reliability/migrate-container-instances)
