### Azure API Management

Use of Availability Zones is recommended. The API Management service supports Zone redundancy, which provides resiliency and high availability to a service instance in a specific Azure region. With zone redundancy, the gateway and the control plane of your API Management instance (Management API, developer portal, Git configuration) are replicated across datacenters in physically separated zones, making it resilient to a zone failure.

The number of units selected must be distributed evenly across the availability zones.

[Migrate Azure API Management to availability zone support | Microsoft Learn](https://learn.microsoft.com/en-us/azure/availability-zones/migrate-api-mgt)

SKU updates can take from 15 to 45 minutes to apply. The API Management gateway can continue to handle API requests during this time.

Availability Zones are only supported when using the Premium SKU: [Feature-based comparison of the Azure API Management tiers | Microsoft Learn](https://learn.microsoft.com/en-us/azure/api-management/api-management-features)