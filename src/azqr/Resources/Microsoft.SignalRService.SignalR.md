### SignalR

Azure SignalR Service uses Azure availability zones to provide high availability and fault tolerance within an Azure region.  

Zone redundancy is a Premium tier feature. It is implicitly enabled when you create or upgrade to a Premium tier resource. Standard tier resources can be upgraded to Premium tier without downtime. 

[Availability zones support in Azure SignalR Service | Microsoft Learn](https://learn.microsoft.com/en-us/azure/azure-signalr/availability-zones)

Resiliency and disaster recovery is a common need for online systems. Azure SignalR Service already guarantees 99.9% availability, but it's still a regional service. Your service instance is always running in one region and won't fail-over to another region when there is a region-wide outage. 

Instead, our service SDK provides a functionality to support multiple SignalR service instances and automatically switch to other instances when some of them are not available. With this feature, you'll be able to recover when a disaster takes place, but you will need to set up the right system topology by yourself.  

Diagram shows two regions each with an app server and a SignalR service, where each server is associated with the SignalR service in its region as primary and with the service in the other region as secondary. 

In order to have cross region resiliency for SignalR service, you need to set up multiple service instances in different regions, check the following document to learn more: [Resiliency and disaster recovery in Azure SignalR Service | Microsoft Docs.](https://learn.microsoft.com/en-us/azure/azure-signalr/signalr-concept-disaster-recovery)