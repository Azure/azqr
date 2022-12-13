### Event Hubs
Use of Availability Zones is recommended. The Azure Event Hubs service supports Zone redundancy, which provides resiliency and high availability to a service instance in a specific Azure region. With zone redundancy, the outage risk is further spread across three physically separated facilities, and the service has enough capacity reserves to instantly cope up with the complete, catastrophic loss of the entire facility. 

Availability Zones are supported when using the Standard, Premium, and Dedicated SKU: [Compare Azure Event Hubs tiers | Microsoft Learn] https://learn.microsoft.com/en-us/azure/event-hubs/compare-tiers

**[Azure Event Hubs - Geo-disaster recovery - Microsoft Learn](https://learn.microsoft.com/en-us/azure/event-hubs/event-hubs-geo-dr)**

This option allows the creation of a secondary namespace in a different region. Only the active namespace receives messages at any time. Messages and events aren't replicated to the secondary region. The RTO for the regional failover is up to 30 minutes. Confirm this RTO aligns with the requirements of the customer and fits in the broader availability strategy. If a higher RTO is required, consider implementing a client-side failover pattern.

**Partitions:**

We recommend sending events to an event hub without setting partition information to allow the Event Hubs service to balance the load across partitions.

