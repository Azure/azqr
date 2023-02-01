### App Configuration

App Configuration Free tier does no have SLA.

**Multi-region deployment in App Configuration**

App Configuration is regional service. For applications with different configurations per region, storing these configurations in one instance can create a single point of failure. Deploying one App Configuration instances per region across multiple regions may be a better option. It can help with regional disaster recovery, performance, and security siloing. Configuring by region also improves latency and uses separated throttling quotas, since throttling is per instance. To apply disaster recovery mitigation, you can use [multiple configuration stores](https://learn.microsoft.com/en-us/azure/azure-app-configuration/concept-disaster-recovery).

---