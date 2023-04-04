# Azure Quick Review Rules

The following table contains the rules that are used by Azure Quick Review to scan Azure resources.

Id | Category | Subcategory | Name | Severity | More Info
---|---|---|---|---|---
aks-008 | Security | Identity and Access Control | AKS should be RBAC enabled. | Medium | https://learn.microsoft.com/azure/aks/manage-azure-rbac
aks-011 | Monitoring and Logging | Monitoring | AKS should have Container Insights enabled | Medium | https://learn.microsoft.com/azure/azure-monitor/insights/container-insights-overview
aks-003 | High Availability and Resiliency | SLA | AKS Cluster should have an SLA | High | https://learn.microsoft.com/en-us/azure/aks/free-standard-pricing-tiers#uptime-sla-terms-and-conditions
aks-006 | Governance | Naming Convention (CAF) | AKS Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
aks-007 | Security | Identity and Access Control | AKS should integrate authentication with AAD | Medium | https://learn.microsoft.com/azure/aks/manage-azure-rbac
aks-013 | Networking | Best Practices | AKS should avoid using kubenet network plugin | High | https://learn.microsoft.com/azure/aks/operator-best-practices-network
aks-009 | Security | Identity and Access Control | AKS should have local accounts disabled | Medium | https://learn.microsoft.com/azure/aks/managed-aad#disable-local-accounts
aks-012 | Security | Networking | AKS should have outbound type set to user defined routing | High | https://learn.microsoft.com/azure/aks/limit-egress-traffic
aks-014 | Operations | Scalability | AKS should have autoscaler enabled | Medium | https://learn.microsoft.com/azure/aks/concepts-scale
aks-004 | Security | Networking | AKS Cluster should be private | High | https://learn.microsoft.com/en-us/azure/aks/private-clusters
aks-005 | High Availability and Resiliency | SKU | AKS Production Cluster should use Standard SKU | High | https://learn.microsoft.com/en-us/azure/aks/free-standard-pricing-tiers
aks-010 | Security | Best Practices | AKS should have httpApplicationRouting disabled | Medium | https://learn.microsoft.com/azure/aks/http-application-routing
aks-015 | Governance | Use tags to organize your resources | AKS should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
aks-001 | Monitoring and Logging | Diagnostic Logs | AKS Cluster should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/aks/monitor-aks#collect-resource-logs
aks-002 | High Availability and Resiliency | Availability Zones | AKS Cluster should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/aks/availability-zones
apim-004 | Networking | Private Endpoint | APIM should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/api-management/private-endpoint
apim-005 | High Availability and Resiliency | SKU | Azure APIM SKU | High | https://learn.microsoft.com/en-us/azure/api-management/api-management-features
apim-006 | Governance | Naming Convention (CAF) | APIM should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
apim-007 | Governance | Use tags to organize your resources | APIM should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
apim-001 | Monitoring and Logging | Diagnostic Logs | APIM should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/api-management/api-management-howto-use-azure-monitor#resource-logs
apim-002 | High Availability and Resiliency | Availability Zones | APIM should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/reliability/migrate-api-mgt
apim-003 | High Availability and Resiliency | SLA | APIM should have a SLA | High | https://www.azure.cn/en-us/support/sla/api-management/
agw-006 | Governance | Naming Convention (CAF) | Application Gateway Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
agw-007 | Governance | Use tags to organize your resources | Application Gateway should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
agw-001 | Monitoring and Logging | Diagnostic Logs | Application Gateway should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/application-gateway/application-gateway-diagnostics#diagnostic-logging
agw-002 | High Availability and Resiliency | Availability Zones | Application Gateway should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/application-gateway/application-gateway-autoscaling-zone-redundant
agw-003 | High Availability and Resiliency | SLA | Application Gateway SLA | High | https://www.azure.cn/en-us/support/sla/application-gateway/
agw-005 | High Availability and Resiliency | SKU | Application Gateway SKU | High | https://learn.microsoft.com/en-us/azure/application-gateway/understanding-pricing
cae-004 | Security | Networking | ContainerApp should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/container-apps/vnet-custom-internal?tabs=bash&pivots=azure-portal
cae-006 | Governance | Naming Convention (CAF) | ContainerApp Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
cae-007 | Governance | Use tags to organize your resources | ContainerApp should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
cae-001 | Monitoring and Logging | Diagnostic Logs | ContainerApp should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/container-apps/log-options#diagnostic-settings
cae-002 | High Availability and Resiliency | Availability Zones | ContainerApp should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/container-apps/disaster-recovery?tabs=bash#set-up-zone-redundancy-in-your-container-apps-environment
cae-003 | High Availability and Resiliency | SLA | ContainerApp should have a SLA | High | https://azure.microsoft.com/en-us/support/legal/sla/container-apps/v1_0/
ci-007 | Governance | Use tags to organize your resources | ContainerInstance should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
ci-002 | High Availability and Resiliency | Availability Zones | ContainerInstance should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/container-instances/availability-zones
ci-003 | High Availability and Resiliency | SLA | ContainerInstance should have a SLA | High | https://www.azure.cn/en-us/support/sla/container-instances/v1_0/index.html
ci-004 | Security | Networking | ContainerInstance should use private IP addresses | High | 
ci-005 | High Availability and Resiliency | SKU | ContainerInstance SKU | High | https://azure.microsoft.com/en-us/pricing/details/container-instances/
ci-006 | Governance | Naming Convention (CAF) | ContainerInstance Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
cosmos-001 | Monitoring and Logging | Diagnostic Logs | CosmosDB should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/cosmos-db/monitor-resource-logs
cosmos-002 | High Availability and Resiliency | Availability Zones | CosmosDB should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/cosmos-db/high-availability
cosmos-003 | High Availability and Resiliency | SLA | CosmosDB should have a SLA | High | https://learn.microsoft.com/en-us/azure/cosmos-db/high-availability#slas
cosmos-004 | Security | Networking | CosmosDB should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/cosmos-db/how-to-configure-private-endpoints
cosmos-005 | High Availability and Resiliency | SKU | CosmosDB SKU | High | https://azure.microsoft.com/en-us/pricing/details/cosmos-db/autoscale-provisioned/
cosmos-006 | Governance | Naming Convention (CAF) | CosmosDB Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
cosmos-007 | Governance | Use tags to organize your resources | CosmosDB should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
cr-002 | High Availability and Resiliency | Availability Zones | ContainerRegistry should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/container-registry/zone-redundancy
cr-003 | High Availability and Resiliency | SLA | ContainerRegistry should have a SLA | High | https://www.azure.cn/en-us/support/sla/container-registry/
cr-004 | Security | Networking | ContainerRegistry should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/container-registry/container-registry-private-link
cr-006 | Governance | Naming Convention (CAF) | ContainerRegistry Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
cr-008 | Security | Identity and Access Control | ContainerRegistry should have the Administrator account disabled | Medium | https://learn.microsoft.com/azure/container-registry/container-registry-authentication-managed-identity
cr-009 | Governance | Use tags to organize your resources | ContainerRegistry should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
cr-001 | Monitoring and Logging | Diagnostic Logs | ContainerRegistry should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/container-registry/monitor-service
cr-005 | High Availability and Resiliency | SKU | ContainerRegistry SKU | High | https://learn.microsoft.com/en-us/azure/container-registry/container-registry-skus
cr-007 | Security | Identity and Access Control | ContainerRegistry should have anonymous pull access disabled | Medium | https://learn.microsoft.com/azure/container-registry/anonymous-pull-access#configure-anonymous-pull-access
cr-010 | Governance | Use retention policies | ContainerRegistry should use retention policies | Medium | https://learn.microsoft.com/en-us/azure/container-registry/container-registry-retention-policy
evh-007 | Governance | Use tags to organize your resources | Event Hub should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
evh-008 | Security | Identity and Access Control | Event Hub should have local authentication disabled | Medium | https://learn.microsoft.com/en-us/azure/event-hubs/authorize-access-event-hubs#shared-access-signatures
evh-001 | Monitoring and Logging | Diagnostic Logs | Event Hub Namespace should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/event-hubs/monitor-event-hubs#collection-and-routing
evh-002 | High Availability and Resiliency | Availability Zones | Event Hub Namespace should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/event-hubs/event-hubs-premium-overview#high-availability-with-availability-zones
evh-003 | High Availability and Resiliency | SLA | Event Hub Namespace should have a SLA | High | https://www.azure.cn/en-us/support/sla/event-hubs/
evh-004 | Security | Networking | Event Hub Namespace should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/event-hubs/network-security
evh-005 | High Availability and Resiliency | SKU | Event Hub Namespace SKU | High | https://learn.microsoft.com/en-us/azure/event-hubs/compare-tiers
evh-006 | Governance | Naming Convention (CAF) | Event Hub Namespace Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
evgd-004 | Security | Networking | Event Grid Domain should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/event-grid/configure-private-endpoints
evgd-005 | High Availability and Resiliency | SKU | Event Grid Domain SKU | High | https://azure.microsoft.com/en-gb/pricing/details/event-grid/
evgd-006 | Governance | Naming Convention (CAF) | Event Grid Domain Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
evgd-007 | Governance | Use tags to organize your resources | Event Grid Domain should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
evgd-008 | Security | Identity and Access Control | Event Grid Domain should have local authentication disabled | Medium | https://learn.microsoft.com/en-us/azure/event-grid/authenticate-with-access-keys-shared-access-signatures
evgd-001 | Monitoring and Logging | Diagnostic Logs | Event Grid Domain should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/event-grid/diagnostic-logs
evgd-003 | High Availability and Resiliency | SLA | Event Grid Domain should have a SLA | High | https://www.azure.cn/en-us/support/sla/event-grid/
kv-009 | High Availability and Resiliency | Reliability | Key Vault should have purge protection enabled | Medium | https://learn.microsoft.com/en-us/azure/key-vault/general/soft-delete-overview#purge-protection
kv-001 | Monitoring and Logging | Diagnostic Logs | Key Vault should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/key-vault/general/monitor-key-vault
kv-003 | High Availability and Resiliency | SLA | Key Vault should have a SLA | High | https://www.azure.cn/en-us/support/sla/key-vault/
kv-004 | Security | Networking | Key Vault should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/key-vault/general/private-link-service
kv-005 | High Availability and Resiliency | SKU | Key Vault SKU | High | https://azure.microsoft.com/en-us/pricing/details/key-vault/
kv-006 | Governance | Naming Convention (CAF) | Key Vault Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
kv-007 | Governance | Use tags to organize your resources | Key Vault should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
kv-008 | High Availability and Resiliency | Reliability | Key Vault should have soft delete enabled | Medium | https://learn.microsoft.com/en-us/azure/key-vault/general/soft-delete-overview
appcs-005 | High Availability and Resiliency | SKU | AppConfiguration SKU | High | https://azure.microsoft.com/en-us/pricing/details/app-configuration/
appcs-006 | Governance | Naming Convention (CAF) | AppConfiguration Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
appcs-007 | Governance | Use tags to organize your resources | AppConfiguration should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
appcs-008 | Security | Identity and Access Control | AppConfiguration should have local authentication disabled | Medium | https://learn.microsoft.com/en-us/azure/azure-app-configuration/howto-disable-access-key-authentication?tabs=portal#disable-access-key-authentication
appcs-001 | Monitoring and Logging | Diagnostic Logs | AppConfiguration should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/azure-app-configuration/monitor-app-configuration?tabs=portal
appcs-003 | High Availability and Resiliency | SLA | AppConfiguration should have a SLA | High | https://www.azure.cn/en-us/support/sla/app-configuration/
appcs-004 | Security | Networking | AppConfiguration should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/azure-app-configuration/concept-private-endpoint
plan-007 | Governance | Use tags to organize your resources | Plan should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
plan-001 | Monitoring and Logging | Diagnostic Logs | Plan should have diagnostic settings enabled | Medium | 
plan-002 | High Availability and Resiliency | Availability Zones | Plan should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/reliability/migrate-app-service
plan-003 | High Availability and Resiliency | SLA | Plan should have a SLA | High | https://www.azure.cn/en-us/support/sla/app-service/
plan-005 | High Availability and Resiliency | SKU | Plan SKU | High | https://learn.microsoft.com/en-us/azure/app-service/overview-hosting-plans
plan-006 | Governance | Naming Convention (CAF) | Plan Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
redis-002 | High Availability and Resiliency | Availability Zones | Redis should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-high-availability
redis-003 | High Availability and Resiliency | SLA | Redis should have a SLA | High | https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services?lang=1
redis-004 | Security | Networking | Redis should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-private-link
redis-006 | Governance | Naming Convention (CAF) | Redis Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
redis-009 | Security | Networking | Redis should enforce TLS >= 1.2 | Low | https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-remove-tls-10-11
redis-001 | Monitoring and Logging | Diagnostic Logs | Redis should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-monitor-diagnostic-settings
redis-005 | High Availability and Resiliency | SKU | Redis SKU | High | https://azure.microsoft.com/en-gb/pricing/details/cache/
redis-007 | Governance | Use tags to organize your resources | Redis should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
redis-008 | Security | Networking | Redis should not enable non SSL ports | High | https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-configure#access-ports
sb-003 | High Availability and Resiliency | SLA | Service Bus should have a SLA | High | https://www.azure.cn/en-us/support/sla/service-bus/
sb-004 | Security | Networking | Service Bus should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/service-bus-messaging/network-security
sb-005 | High Availability and Resiliency | SKU | Service Bus SKU | High | https://azure.microsoft.com/en-us/pricing/details/service-bus/
sb-006 | Governance | Naming Convention (CAF) | Service Bus Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
sb-007 | Governance | Use tags to organize your resources | Service Bus should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
sb-008 | Security | Identity and Access Control | Service Bus should have local authentication disabled | Medium | https://learn.microsoft.com/en-us/azure/service-bus-messaging/service-bus-sas
sb-001 | Monitoring and Logging | Diagnostic Logs | Service Bus should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/service-bus-messaging/monitor-service-bus#collection-and-routing
sb-002 | High Availability and Resiliency | Availability Zones | Service Bus should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/service-bus-messaging/service-bus-outages-disasters#availability-zones
sigr-005 | High Availability and Resiliency | SKU | SignalR SKU | High | https://azure.microsoft.com/en-us/pricing/details/signalr-service/
sigr-006 | Governance | Naming Convention (CAF) | SignalR Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
sigr-007 | Governance | Use tags to organize your resources | SignalR should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
sigr-001 | Monitoring and Logging | Diagnostic Logs | SignalR should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/azure-signalr/signalr-howto-diagnostic-logs
sigr-002 | High Availability and Resiliency | Availability Zones | SignalR should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/azure-signalr/availability-zones
sigr-003 | High Availability and Resiliency | SLA | SignalR should have a SLA | High | https://www.azure.cn/en-us/support/sla/signalr-service/
sigr-004 | Security | Networking | SignalR should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/azure-signalr/howto-private-endpoints
wps-006 | Governance | Naming Convention (CAF) | Web Pub Sub Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
wps-007 | Governance | Use tags to organize your resources | Web Pub Sub should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
wps-001 | Monitoring and Logging | Diagnostic Logs | Web Pub Sub should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/azure-web-pubsub/howto-troubleshoot-resource-logs
wps-002 | High Availability and Resiliency | Availability Zones | Web Pub Sub should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/azure-web-pubsub/concept-availability-zones
wps-003 | High Availability and Resiliency | SLA | Web Pub Sub should have a SLA | High | https://azure.microsoft.com/en-gb/support/legal/sla/web-pubsub/
wps-004 | Security | Networking | Web Pub Sub should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/azure-web-pubsub/howto-secure-private-endpoints
wps-005 | High Availability and Resiliency | SKU | Web Pub Sub SKU | High | https://azure.microsoft.com/en-us/pricing/details/web-pubsub/
st-001 | Monitoring and Logging | Diagnostic Logs | Storage should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/storage/blobs/monitor-blob-storage
st-004 | Security | Networking | Storage should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/storage/common/storage-private-endpoints
st-006 | Governance | Naming Convention (CAF) | Storage Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
st-007 | Security | Network Security | Storage Account should use HTTPS only | High | https://learn.microsoft.com/en-us/azure/storage/common/storage-require-secure-transfer
st-008 | Governance | Use tags to organize your resources | Storage Account should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
st-002 | High Availability and Resiliency | Availability Zones | Storage should have availability zones enabled | High | https://learn.microsoft.com/EN-US/azure/reliability/migrate-storage
st-003 | High Availability and Resiliency | SLA | Storage should have a SLA | High | https://www.azure.cn/en-us/support/sla/storage/
st-005 | High Availability and Resiliency | SKU | Storage SKU | High | https://learn.microsoft.com/en-us/rest/api/storagerp/srp_sku_types
st-009 | Security | Networking | Storage Account should enforce TLS >= 1.2 | Low | https://learn.microsoft.com/en-us/azure/storage/common/transport-layer-security-configure-minimum-version?tabs=portal
psql-004 | Security | Networking | PostgreSQL should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/postgresql/single-server/concepts-data-access-and-security-private-link
psql-005 | High Availability and Resiliency | SKU | PostgreSQL SKU | High | https://learn.microsoft.com/en-us/azure/postgresql/single-server/concepts-pricing-tiers
psql-006 | Governance | Naming Convention (CAF) | PostgreSQL Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
psql-007 | Governance | Use tags to organize your resources | PostgreSQL should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
psql-008 | Security | Networking | PostgreSQL should enforce SSL | High | https://learn.microsoft.com/en-us/azure/postgresql/single-server/concepts-ssl-connection-security#enforcing-tls-connections
psql-009 | Security | Networking | PostgreSQL should enforce TLS >= 1.2 | Low | https://learn.microsoft.com/en-us/azure/postgresql/single-server/how-to-tls-configurations
psql-001 | Monitoring and Logging | Diagnostic Logs | PostgreSQL should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/postgresql/single-server/concepts-server-logs#resource-logs
psql-003 | High Availability and Resiliency | SLA | PostgreSQL should have a SLA | High | https://www.azure.cn/en-us/support/sla/postgresql/
psqlf-003 | High Availability and Resiliency | SLA | PostgreSQL should have a SLA | High | https://learn.microsoft.com/en-us/azure/postgresql/flexible-server/concepts-compare-single-server-flexible-server
psqlf-004 | Security | Private Access | PostgreSQL should have private access enabled | High | https://learn.microsoft.com/en-us/azure/postgresql/flexible-server/concepts-networking#private-access-vnet-integration
psqlf-005 | High Availability and Resiliency | SKU | PostgreSQL SKU | High | https://azure.microsoft.com/en-gb/pricing/details/postgresql/flexible-server/
psqlf-006 | Governance | Naming Convention (CAF) | PostgreSQL Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
psqlf-007 | Governance | Use tags to organize your resources | PostgreSQL should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
psqlf-001 | Monitoring and Logging | Diagnostic Logs | PostgreSQL should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/postgresql/flexible-server/howto-configure-and-access-logs
psqlf-002 | High Availability and Resiliency | Availability Zones | PostgreSQL should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/postgresql/flexible-server/overview#architecture-and-high-availability
sql-001 | Monitoring and Logging | Diagnostic Logs | SQL should have diagnostic settings enabled | Medium | 
sql-004 | Security | Networking | SQL should have private endpoints enabled | High | 
sql-006 | Governance | Naming Convention (CAF) | SQL Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
sql-007 | Governance | Use tags to organize your resources | SQL should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
sql-008 | Security | Networking | SQL should enforce TLS >= 1.2 | Low | https://learn.microsoft.com/en-us/azure/azure-sql/database/connectivity-settings?view=azuresql&tabs=azure-portal#minimal-tls-version
afd-001 | Monitoring and Logging | Diagnostic Logs | Azure FrontDoor should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/frontdoor/standard-premium/how-to-logs
afd-003 | High Availability and Resiliency | SLA | Azure FrontDoor SLA | High | https://www.azure.cn/en-us/support/sla/cdn/
afd-005 | High Availability and Resiliency | SKU | Azure FrontDoor SKU | High | https://learn.microsoft.com/en-us/azure/frontdoor/standard-premium/tier-comparison
afd-006 | Governance | Naming Convention | Azure FrontDoor Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
afd-007 | Governance | Use tags to organize your resources | Azure FrontDoor should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
afw-006 | Governance | Naming Convention | Azure Firewall Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
afw-007 | Governance | Use tags to organize your resources | Azure Firewall should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
afw-001 | Monitoring and Logging | Diagnostic Logs | Azure Firewall should have diagnostic settings enabled | Medium | https://docs.microsoft.com/en-us/azure/firewall/logs-and-metrics
afw-002 | High Availability and Resiliency | Availability Zones | Azure Firewall should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/firewall/features#availability-zones
afw-003 | High Availability and Resiliency | SLA | Azure Firewall SLA | High | https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services
afw-005 | High Availability and Resiliency | SKU | Azure Firewall SKU | High | https://learn.microsoft.com/en-us/azure/firewall/choose-firewall-sku
mysql-003 | High Availability and Resiliency | SLA | Azure Database for MySQL - Flexible Server should have a SLA | High | https://www.azure.cn/en-us/support/sla/mysql/
mysql-004 | Security | Networking | Azure Database for MySQL - Flexible Server should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/mysql/single-server/concepts-data-access-security-private-link
mysql-005 | High Availability and Resiliency | SKU | Azure Database for MySQL - Flexible Server SKU | High | https://learn.microsoft.com/en-us/azure/mysql/single-server/concepts-pricing-tiers
mysql-006 | Governance | Naming Convention (CAF) | Azure Database for MySQL - Flexible Server Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
mysql-007 | Operations | Best Practices | Azure Database for MySQL - Single Server is on the retirement path | High | https://learn.microsoft.com/en-us/azure/mysql/single-server/whats-happening-to-mysql-single-server
mysql-008 | Governance | Use tags to organize your resources | Azure Database for MySQL - Single Server should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
mysql-001 | Monitoring and Logging | Diagnostic Logs | Azure Database for MySQL - Flexible Server should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/mysql/single-server/concepts-monitoring#server-logs
mysqlf-001 | Monitoring and Logging | Diagnostic Logs | Azure Database for MySQL - Flexible Server should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/mysql/flexible-server/tutorial-query-performance-insights#set-up-diagnostics
mysqlf-002 | High Availability and Resiliency | Availability Zones | Azure Database for MySQL - Flexible Server should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/mysql/flexible-server/how-to-configure-high-availability-cli
mysqlf-003 | High Availability and Resiliency | SLA | Azure Database for MySQL - Flexible Server should have a SLA | High | hhttps://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services?lang=1
mysqlf-004 | Security | Private Access | Azure Database for MySQL - Flexible Server should have private access enabled | High | https://learn.microsoft.com/en-us/azure/mysql/flexible-server/how-to-manage-virtual-network-cli
mysqlf-005 | High Availability and Resiliency | SKU | Azure Database for MySQL - Flexible Server SKU | High | https://learn.microsoft.com/en-us/azure/mysql/flexible-server/concepts-service-tiers-storage
mysqlf-006 | Governance | Naming Convention (CAF) | Azure Database for MySQL - Flexible Server Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
mysqlf-007 | Governance | Use tags to organize your resources | Azure Database for MySQL - Flexible Server should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
