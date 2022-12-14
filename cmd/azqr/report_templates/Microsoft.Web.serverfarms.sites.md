### Azure App Service

Use of Availability Zones is recommended. The Azure App Service service supports Zone redundancy, which provides resiliency and high availability to a service instance in a specific Azure region. 

Azure Functions supports both zone-redundant and zonal instances: [What is reliability in Azure Functions? | Microsoft Learn](https://learn.microsoft.com/en-us/azure/reliability/reliability-functions)

Downtime will be dependent on how you decide to carry out the migration and how you choose to redirect traffic from your old to your new availability zone enabled App Service

Availability Zones are only supported when using the Premium SKU: [Migrate App Service to availability zone support | Microsoft Learn](https://learn.microsoft.com/en-us/azure/reliability/migrate-app-service)
