# Log Analytics

Use of Availability Zones is recommended. The Log Analytics service supports Zone redundancy, which provides resiliency and high availability to a service instance in a specific Azure region. When a workspace is linked to an availability zone, it remains active and operational even if a specific datacenter is malfunctioning or completely down, by relying on the availability of other zones in the region.

Not all dedicated clusters can use availability zones. Dedicated clusters created after mid-October 2020 can be set to support availability zones when they are created. New clusters created after that date default to be enabled for availability zones in regions where Azure Monitor supports them.

[Availability zones in Azure Monitor | Microsoft Learn](https://learn.microsoft.com/en-us/azure/azure-monitor/logs/availability-zones)

Availability Zones are supported when using the Log Analytics workspace linked to an Azure Monitor dedicated cluster: [Azure Monitor Logs Dedicated Clusters | Microsoft Learn](https://learn.microsoft.com/en-us/azure/azure-monitor/logs/logs-dedicated-clusters)

If you use the Log Analytics agent for Linux: Migrate to Azure Monitor Agent or ensure that your Linux machines only require access to a single workspace.

[Azure Monitor Agent overview - Microsoft Learn](https://learn.microsoft.com/en-us/azure/azure-monitor/agents/agents-overview)
