### Container Apps

Use of Availability Zones is recommended. 

The Azure Container Apps service supports Zone redundancy, which provides resiliency and high availability to a service instance in a specific Azure region. 

By enabling Container Apps' zone redundancy feature, replicas are automatically randomly distributed across the zones in the region.

The number of units selected must be distributed evenly across the availability zones.

To ensure proper distribution of replicas, you should configure your app's minimum and maximum replica count with values that are divisible by three. The minimum replica count should be at least three.

[Disaster recovery guidance for Azure Container Apps | Microsoft Learn](https://learn.microsoft.com/en-us/azure/container-apps/disaster-recovery)
