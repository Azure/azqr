# Azure Quick Review Rules

Azure Quick Review uses the following rules to identify Azure resources that may be or not be compliant with Azure best practices and recommendations:

\#  | Id | Category | Subcategory | Name | Severity | More Info
---|---|---|---|---|---|---
1 | adf-001 | Reliability | Diagnostic Logs | Azure Data Factory should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/data-factory/monitor-configure-diagnostics
2 | adf-002 | Security | Private Endpoint | Azure Data Factory should have private endpoints enabled | High | 
3 | adf-003 | Reliability | SLA | Azure Data Factory SLA | High | https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services
4 | adf-004 | Operational Excellence | Naming Convention (CAF) | Azure Data Factory Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
5 | adf-005 | Operational Excellence | Tags | Azure Data Factory should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
6 | afd-001 | Reliability | Diagnostic Logs | Azure FrontDoor should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/frontdoor/standard-premium/how-to-logs
7 | afd-003 | Reliability | SLA | Azure FrontDoor SLA | High | https://www.azure.cn/en-us/support/sla/cdn/
8 | afd-005 | Reliability | SKU | Azure FrontDoor SKU | High | https://learn.microsoft.com/en-us/azure/frontdoor/standard-premium/tier-comparison
9 | afd-006 | Operational Excellence | Naming Convention (CAF) | Azure FrontDoor Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
10 | afd-007 | Operational Excellence | Tags | Azure FrontDoor should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
11 | afw-001 | Reliability | Diagnostic Logs | Azure Firewall should have diagnostic settings enabled | Medium | https://docs.microsoft.com/en-us/azure/firewall/logs-and-metrics
12 | afw-002 | Reliability | Availability Zones | Azure Firewall should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/firewall/features#availability-zones
13 | afw-003 | Reliability | SLA | Azure Firewall SLA | High | https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services
14 | afw-005 | Reliability | SKU | Azure Firewall SKU | High | https://learn.microsoft.com/en-us/azure/firewall/choose-firewall-sku
15 | afw-006 | Operational Excellence | Naming Convention (CAF) | Azure Firewall Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
16 | afw-007 | Operational Excellence | Tags | Azure Firewall should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
17 | agw-001 | Reliability | Scaling | Application Gateway: Ensure autoscaling is used with a minimum of 2 instances | High | https://learn.microsoft.com/en-us/azure/application-gateway/application-gateway-autoscaling-zone-redundant
18 | agw-002 | Security | SSL | Application Gateway: Secure all incoming connections with SSL | High | https://learn.microsoft.com/en-us/azure/well-architected/services/networking/azure-application-gateway#security
19 | agw-003 | Security | Firewall | Application Gateway: Enable WAF policies | High | https://learn.microsoft.com/en-us/azure/application-gateway/features#web-application-firewall
20 | agw-004 | Reliability | SKU | Application Gateway: Use Application GW V2 instead of V1 | High | https://azure.microsoft.com/en-us/updates/application-gateway-v1-will-be-retired-on-28-april-2026-transition-to-application-gateway-v2/
21 | agw-005 | Reliability | Diagnostic Logs | Application Gateway: Monitor and Log the configurations and traffic | Medium | https://learn.microsoft.com/en-us/azure/application-gateway/application-gateway-diagnostics#diagnostic-logging
22 | agw-007 | Reliability | Availability Zones | Application Gateway should have availability zones enabled | Medium | https://learn.microsoft.com/en-us/azure/application-gateway/application-gateway-autoscaling-zone-redundant
23 | agw-008 | Reliability | Maintenance | Application Gateway: Plan for backend maintenance by using connection draining | Medium | https://learn.microsoft.com/en-us/azure/application-gateway/features#connection-draining
24 | agw-103 | Reliability | SLA | Application Gateway SLA | High | https://www.azure.cn/en-us/support/sla/application-gateway/
25 | agw-104 | Reliability | SKU | Application Gateway SKU | High | https://learn.microsoft.com/en-us/azure/application-gateway/understanding-pricing
26 | agw-105 | Operational Excellence | Naming Convention (CAF) | Application Gateway Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
27 | agw-106 | Operational Excellence | Tags | Application Gateway should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
28 | aks-001 | Reliability | Diagnostic Logs | AKS Cluster should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/aks/monitor-aks#collect-resource-logs
29 | aks-002 | Reliability | Availability Zones | AKS Cluster should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/aks/availability-zones
30 | aks-003 | Reliability | SLA | AKS Cluster should have an SLA | High | https://learn.microsoft.com/en-us/azure/aks/free-standard-pricing-tiers#uptime-sla-terms-and-conditions
31 | aks-004 | Security | Private Endpoint | AKS Cluster should be private | High | https://learn.microsoft.com/en-us/azure/aks/private-clusters
32 | aks-005 | Reliability | SKU | AKS Production Cluster should use Standard SKU | High | https://learn.microsoft.com/en-us/azure/aks/free-standard-pricing-tiers
33 | aks-006 | Operational Excellence | Naming Convention (CAF) | AKS Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
34 | aks-007 | Security | Identity and Access Control | AKS should integrate authentication with AAD (Managed) | Medium | https://learn.microsoft.com/en-us/azure/aks/managed-azure-ad
35 | aks-008 | Security | Identity and Access Control | AKS should be RBAC enabled. | Medium | https://learn.microsoft.com/azure/aks/manage-azure-rbac
36 | aks-009 | Security | Identity and Access Control | AKS should have local accounts disabled | Medium | https://learn.microsoft.com/azure/aks/managed-aad#disable-local-accounts
37 | aks-010 | Security | Best Practices | AKS should have httpApplicationRouting disabled | Medium | https://learn.microsoft.com/azure/aks/http-application-routing
38 | aks-011 | Reliability | Monitoring | AKS should have Container Insights enabled | Medium | https://learn.microsoft.com/azure/azure-monitor/insights/container-insights-overview
39 | aks-012 | Security | Networking | AKS should have outbound type set to user defined routing | High | https://learn.microsoft.com/azure/aks/limit-egress-traffic
40 | aks-013 | Performance Efficiency | Networking | AKS should avoid using kubenet network plugin | Medium | https://learn.microsoft.com/azure/aks/operator-best-practices-network
41 | aks-014 | Operational Excellence | Scaling | AKS should have autoscaler enabled | Medium | https://learn.microsoft.com/azure/aks/concepts-scale
42 | aks-015 | Operational Excellence | Tags | AKS should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
43 | apim-001 | Reliability | Diagnostic Logs | APIM should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/api-management/api-management-howto-use-azure-monitor#resource-logs
44 | apim-002 | Reliability | Availability Zones | APIM should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/reliability/migrate-api-mgt
45 | apim-003 | Reliability | SLA | APIM should have a SLA | High | https://www.azure.cn/en-us/support/sla/api-management/
46 | apim-004 | Security | Private Endpoint | APIM should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/api-management/private-endpoint
47 | apim-005 | Reliability | SKU | Azure APIM SKU | High | https://learn.microsoft.com/en-us/azure/api-management/api-management-features
48 | apim-006 | Operational Excellence | Naming Convention (CAF) | APIM should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
49 | apim-007 | Operational Excellence | Tags | APIM should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
50 | appcs-001 | Reliability | Diagnostic Logs | AppConfiguration should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/azure-app-configuration/monitor-app-configuration?tabs=portal
51 | appcs-003 | Reliability | SLA | AppConfiguration should have a SLA | High | https://www.azure.cn/en-us/support/sla/app-configuration/
52 | appcs-004 | Security | Private Endpoint | AppConfiguration should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/azure-app-configuration/concept-private-endpoint
53 | appcs-005 | Reliability | SKU | AppConfiguration SKU | High | https://azure.microsoft.com/en-us/pricing/details/app-configuration/
54 | appcs-006 | Operational Excellence | Naming Convention (CAF) | AppConfiguration Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
55 | appcs-007 | Operational Excellence | Tags | AppConfiguration should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
56 | appcs-008 | Security | Identity and Access Control | AppConfiguration should have local authentication disabled | Medium | https://learn.microsoft.com/en-us/azure/azure-app-configuration/howto-disable-access-key-authentication?tabs=portal#disable-access-key-authentication
57 | appi-001 | Reliability | SLA | Azure Application Insights SLA | High | https://www.azure.cn/en-us/support/sla/application-insights/index.html
58 | appi-002 | Operational Excellence | Naming Convention (CAF) | Azure Application Insights Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
59 | appi-003 | Operational Excellence | Tags | Azure Application Insights should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
60 | appi-004 | Operational Excellence | Tags | Azure Application Insights should store data in a Log Analytics Workspace | Low | https://learn.microsoft.com/en-us/azure/azure-monitor/app/create-workspace-resource
61 | cae-001 | Reliability | Diagnostic Logs | ContainerApp should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/container-apps/log-options#diagnostic-settings
62 | cae-002 | Reliability | Availability Zones | ContainerApp should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/container-apps/disaster-recovery?tabs=bash#set-up-zone-redundancy-in-your-container-apps-environment
63 | cae-003 | Reliability | SLA | ContainerApp should have a SLA | High | https://azure.microsoft.com/en-us/support/legal/sla/container-apps/v1_0/
64 | cae-004 | Security | Private Endpoint | ContainerApp should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/container-apps/vnet-custom-internal?tabs=bash&pivots=azure-portal
65 | cae-006 | Operational Excellence | Naming Convention (CAF) | ContainerApp Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
66 | cae-007 | Operational Excellence | Tags | ContainerApp should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
67 | ci-002 | Reliability | Availability Zones | ContainerInstance should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/container-instances/availability-zones
68 | ci-003 | Reliability | SLA | ContainerInstance should have a SLA | High | https://www.azure.cn/en-us/support/sla/container-instances/v1_0/index.html
69 | ci-004 | Security | Private IP Address | ContainerInstance should use private IP addresses | High | 
70 | ci-005 | Reliability | SKU | ContainerInstance SKU | High | https://azure.microsoft.com/en-us/pricing/details/container-instances/
71 | ci-006 | Operational Excellence | Naming Convention (CAF) | ContainerInstance Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
72 | ci-007 | Operational Excellence | Tags | ContainerInstance should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
73 | cog-001 | Reliability | Diagnostic Logs | Cognitive Service Account should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/event-hubs/monitor-event-hubs#collection-and-routing
74 | cog-003 | Reliability | SLA | Cognitive Service Account should have a SLA | High | https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services?lang=1
75 | cog-004 | Security | Private Endpoint | Cognitive Service Account should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/cognitive-services/cognitive-services-virtual-networks
76 | cog-005 | Reliability | SKU | Cognitive Service Account SKU | High | https://learn.microsoft.com/en-us/azure/templates/microsoft.cognitiveservices/accounts?pivots=deployment-language-bicep#sku
77 | cog-006 | Operational Excellence | Naming Convention (CAF) | Cognitive Service Account Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
78 | cog-007 | Operational Excellence | Tags | Cognitive Service Account should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
79 | cog-008 | Security | Identity and Access Control | Cognitive Service Account should have local authentication disabled | Medium | https://learn.microsoft.com/en-us/azure/ai-services/policy-reference#azure-ai-services
80 | cosmos-001 | Reliability | Diagnostic Logs | CosmosDB should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/cosmos-db/monitor-resource-logs
81 | cosmos-002 | Reliability | Availability Zones | CosmosDB should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/cosmos-db/high-availability
82 | cosmos-003 | Reliability | SLA | CosmosDB should have a SLA | High | https://learn.microsoft.com/en-us/azure/cosmos-db/high-availability#slas
83 | cosmos-004 | Security | Private Endpoint | CosmosDB should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/cosmos-db/how-to-configure-private-endpoints
84 | cosmos-005 | Reliability | SKU | CosmosDB SKU | High | https://azure.microsoft.com/en-us/pricing/details/cosmos-db/autoscale-provisioned/
85 | cosmos-006 | Operational Excellence | Naming Convention (CAF) | CosmosDB Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
86 | cosmos-007 | Operational Excellence | Tags | CosmosDB should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
87 | cr-001 | Reliability | Diagnostic Logs | ContainerRegistry should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/container-registry/monitor-service
88 | cr-002 | Reliability | Availability Zones | ContainerRegistry should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/container-registry/zone-redundancy
89 | cr-003 | Reliability | SLA | ContainerRegistry should have a SLA | High | https://www.azure.cn/en-us/support/sla/container-registry/
90 | cr-004 | Security | Private Endpoint | ContainerRegistry should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/container-registry/container-registry-private-link
91 | cr-005 | Reliability | SKU | ContainerRegistry SKU | High | https://learn.microsoft.com/en-us/azure/container-registry/container-registry-skus
92 | cr-006 | Operational Excellence | Naming Convention (CAF) | ContainerRegistry Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
93 | cr-007 | Security | Identity and Access Control | ContainerRegistry should have anonymous pull access disabled | Medium | https://learn.microsoft.com/azure/container-registry/anonymous-pull-access#configure-anonymous-pull-access
94 | cr-008 | Security | Identity and Access Control | ContainerRegistry should have the Administrator account disabled | Medium | https://learn.microsoft.com/azure/container-registry/container-registry-authentication-managed-identity
95 | cr-009 | Operational Excellence | Tags | ContainerRegistry should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
96 | cr-010 | Operational Excellence | Retention Policies | ContainerRegistry should use retention policies | Medium | https://learn.microsoft.com/en-us/azure/container-registry/container-registry-retention-policy
97 | dec-001 | Reliability | Diagnostic Logs | Azure Data Explorer should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/data-explorer/using-diagnostic-logs
98 | dec-002 | Reliability | SLA | Azure Data Explorer SLA | High | https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services
99 | dec-003 | Reliability | SKU | Azure Data Explorer SKU | High | https://learn.microsoft.com/en-us/azure/data-explorer/manage-cluster-choose-sku
100 | dec-004 | Operational Excellence | Naming Convention (CAF) | Azure Data Explorer Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
101 | dec-005 | Operational Excellence | Tags | Azure Data Explorer should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
102 | evgd-001 | Reliability | Diagnostic Logs | Event Grid Domain should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/event-grid/diagnostic-logs
103 | evgd-003 | Reliability | SLA | Event Grid Domain should have a SLA | High | https://www.azure.cn/en-us/support/sla/event-grid/
104 | evgd-004 | Security | Private Endpoint | Event Grid Domain should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/event-grid/configure-private-endpoints
105 | evgd-005 | Reliability | SKU | Event Grid Domain SKU | High | https://azure.microsoft.com/en-gb/pricing/details/event-grid/
106 | evgd-006 | Operational Excellence | Naming Convention (CAF) | Event Grid Domain Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
107 | evgd-007 | Operational Excellence | Tags | Event Grid Domain should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
108 | evgd-008 | Security | Identity and Access Control | Event Grid Domain should have local authentication disabled | Medium | https://learn.microsoft.com/en-us/azure/event-grid/authenticate-with-access-keys-shared-access-signatures
109 | evh-001 | Reliability | Diagnostic Logs | Event Hub Namespace should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/event-hubs/monitor-event-hubs#collection-and-routing
110 | evh-002 | Reliability | Availability Zones | Event Hub Namespace should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/event-hubs/event-hubs-premium-overview#high-availability-with-availability-zones
111 | evh-003 | Reliability | SLA | Event Hub Namespace should have a SLA | High | https://www.azure.cn/en-us/support/sla/event-hubs/
112 | evh-004 | Security | Private Endpoint | Event Hub Namespace should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/event-hubs/network-security
113 | evh-005 | Reliability | SKU | Event Hub Namespace SKU | High | https://learn.microsoft.com/en-us/azure/event-hubs/compare-tiers
114 | evh-006 | Operational Excellence | Naming Convention (CAF) | Event Hub Namespace Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
115 | evh-007 | Operational Excellence | Tags | Event Hub should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
116 | evh-008 | Security | Identity and Access Control | Event Hub should have local authentication disabled | Medium | https://learn.microsoft.com/en-us/azure/event-hubs/authorize-access-event-hubs#shared-access-signatures
117 | kv-001 | Reliability | Diagnostic Logs | Key Vault should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/key-vault/general/monitor-key-vault
118 | kv-003 | Reliability | SLA | Key Vault should have a SLA | High | https://www.azure.cn/en-us/support/sla/key-vault/
119 | kv-004 | Security | Private Endpoint | Key Vault should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/key-vault/general/private-link-service
120 | kv-005 | Reliability | SKU | Key Vault SKU | High | https://azure.microsoft.com/en-us/pricing/details/key-vault/
121 | kv-006 | Operational Excellence | Naming Convention (CAF) | Key Vault Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
122 | kv-007 | Operational Excellence | Tags | Key Vault should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
123 | kv-008 | Reliability | Reliability | Key Vault should have soft delete enabled | Medium | https://learn.microsoft.com/en-us/azure/key-vault/general/soft-delete-overview
124 | kv-009 | Reliability | Reliability | Key Vault should have purge protection enabled | Medium | https://learn.microsoft.com/en-us/azure/key-vault/general/soft-delete-overview#purge-protection
125 | lb-001 | Reliability | Diagnostic Logs | Load Balancer should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/load-balancer/monitor-load-balancer#creating-a-diagnostic-setting
126 | lb-002 | Reliability | Availability Zones | Load Balancer should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/load-balancer/load-balancer-standard-availability-zones#zone-redundant
127 | lb-003 | Reliability | SLA | Load Balancer should have a SLA | High | https://learn.microsoft.com/en-us/azure/load-balancer/skus
128 | lb-005 | Reliability | SKU | Load Balancer SKU | High | https://learn.microsoft.com/en-us/azure/load-balancer/skus
129 | lb-006 | Operational Excellence | Naming Convention (CAF) | Load Balancer Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
130 | lb-007 | Operational Excellence | Tags | Load Balancer should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
131 | maria-001 | Reliability | Diagnostic Logs | MariaDB should have diagnostic settings enabled | Medium | 
132 | maria-002 | Security | Private Endpoint | MariaDB should have private endpoints enabled | High | 
133 | maria-003 | Operational Excellence | Naming Convention (CAF) | MariaDB server Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
134 | maria-004 | Reliability | SLA | MariaDB server should have a SLA | High | 
135 | maria-005 | Operational Excellence | Tags | MariaDB should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
136 | maria-006 | Security | TLS | MariaDB should enforce TLS >= 1.2 | Low | https://learn.microsoft.com/en-us/azure/mariadb/howto-tls-configurations
137 | mysqlf-001 | Reliability | Diagnostic Logs | Azure Database for MySQL - Flexible Server should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/mysql/flexible-server/tutorial-query-performance-insights#set-up-diagnostics
138 | mysqlf-002 | Reliability | Availability Zones | Azure Database for MySQL - Flexible Server should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/mysql/flexible-server/how-to-configure-high-availability-cli
139 | mysqlf-003 | Reliability | SLA | Azure Database for MySQL - Flexible Server should have a SLA | High | hhttps://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services?lang=1
140 | mysqlf-004 | Security | Private IP Address | Azure Database for MySQL - Flexible Server should have private access enabled | High | https://learn.microsoft.com/en-us/azure/mysql/flexible-server/how-to-manage-virtual-network-cli
141 | mysqlf-005 | Reliability | SKU | Azure Database for MySQL - Flexible Server SKU | High | https://learn.microsoft.com/en-us/azure/mysql/flexible-server/concepts-service-tiers-storage
142 | mysqlf-006 | Operational Excellence | Naming Convention (CAF) | Azure Database for MySQL - Flexible Server Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
143 | mysqlf-007 | Operational Excellence | Tags | Azure Database for MySQL - Flexible Server should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
144 | mysql-001 | Reliability | Diagnostic Logs | Azure Database for MySQL - Flexible Server should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/mysql/single-server/concepts-monitoring#server-logs
145 | mysql-003 | Reliability | SLA | Azure Database for MySQL - Flexible Server should have a SLA | High | https://www.azure.cn/en-us/support/sla/mysql/
146 | mysql-004 | Security | Private Endpoint | Azure Database for MySQL - Flexible Server should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/mysql/single-server/concepts-data-access-security-private-link
147 | mysql-005 | Reliability | SKU | Azure Database for MySQL - Flexible Server SKU | High | https://learn.microsoft.com/en-us/azure/mysql/single-server/concepts-pricing-tiers
148 | mysql-006 | Operational Excellence | Naming Convention (CAF) | Azure Database for MySQL - Flexible Server Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
149 | mysql-007 | Reliability | SKU | Azure Database for MySQL - Single Server is on the retirement path | High | https://learn.microsoft.com/en-us/azure/mysql/single-server/whats-happening-to-mysql-single-server
150 | mysql-008 | Operational Excellence | Tags | Azure Database for MySQL - Single Server should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
151 | app-001 | Reliability | Diagnostic Logs | App Service should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/app-service/troubleshoot-diagnostic-logs#send-logs-to-azure-monitor
152 | app-004 | Security | Private Endpoint | App Service should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/app-service/networking/private-endpoint
153 | app-006 | Operational Excellence | Naming Convention (CAF) | App Service Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
154 | app-007 | Security | HTTPS Only | App Service should use HTTPS only | High | https://learn.microsoft.com/azure/app-service/configure-ssl-bindings#enforce-https
155 | app-008 | Operational Excellence | Tags | App Service should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
156 | func-001 | Reliability | Diagnostic Logs | Function should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/azure-functions/functions-monitor-log-analytics?tabs=csharp
157 | func-004 | Security | Private Endpoint | Function should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/azure-functions/functions-create-vnet
158 | func-006 | Operational Excellence | Naming Convention (CAF) | Function Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
159 | func-007 | Security | HTTPS Only | Function should use HTTPS only | High | https://learn.microsoft.com/azure/app-service/configure-ssl-bindings#enforce-https
160 | func-008 | Operational Excellence | Tags | Function should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
161 | plan-001 | Reliability | Diagnostic Logs | Plan should have diagnostic settings enabled | Medium | 
162 | plan-002 | Reliability | Availability Zones | Plan should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/reliability/migrate-app-service
163 | plan-003 | Reliability | SLA | Plan should have a SLA | High | https://www.azure.cn/en-us/support/sla/app-service/
164 | plan-005 | Reliability | SKU | Plan SKU | High | https://learn.microsoft.com/en-us/azure/app-service/overview-hosting-plans
165 | plan-006 | Operational Excellence | Naming Convention (CAF) | Plan Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
166 | plan-007 | Operational Excellence | Tags | Plan should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
167 | psqlf-001 | Reliability | Diagnostic Logs | PostgreSQL should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/postgresql/flexible-server/howto-configure-and-access-logs
168 | psqlf-002 | Reliability | Availability Zones | PostgreSQL should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/postgresql/flexible-server/overview#architecture-and-high-availability
169 | psqlf-003 | Reliability | SLA | PostgreSQL should have a SLA | High | https://learn.microsoft.com/en-us/azure/postgresql/flexible-server/concepts-compare-single-server-flexible-server
170 | psqlf-004 | Security | Private IP Address | PostgreSQL should have private access enabled | High | https://learn.microsoft.com/en-us/azure/postgresql/flexible-server/concepts-networking#private-access-vnet-integration
171 | psqlf-005 | Reliability | SKU | PostgreSQL SKU | High | https://azure.microsoft.com/en-gb/pricing/details/postgresql/flexible-server/
172 | psqlf-006 | Operational Excellence | Naming Convention (CAF) | PostgreSQL Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
173 | psqlf-007 | Operational Excellence | Tags | PostgreSQL should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
174 | psql-001 | Reliability | Diagnostic Logs | PostgreSQL should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/postgresql/single-server/concepts-server-logs#resource-logs
175 | psql-003 | Reliability | SLA | PostgreSQL should have a SLA | High | https://www.azure.cn/en-us/support/sla/postgresql/
176 | psql-004 | Security | Private Endpoint | PostgreSQL should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/postgresql/single-server/concepts-data-access-and-security-private-link
177 | psql-005 | Reliability | SKU | PostgreSQL SKU | High | https://learn.microsoft.com/en-us/azure/postgresql/single-server/concepts-pricing-tiers
178 | psql-006 | Operational Excellence | Naming Convention (CAF) | PostgreSQL Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
179 | psql-007 | Operational Excellence | Tags | PostgreSQL should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
180 | psql-008 | Security | SSL | PostgreSQL should enforce SSL | High | https://learn.microsoft.com/en-us/azure/postgresql/single-server/concepts-ssl-connection-security#enforcing-tls-connections
181 | psql-009 | Security | TLS | PostgreSQL should enforce TLS >= 1.2 | Low | https://learn.microsoft.com/en-us/azure/postgresql/single-server/how-to-tls-configurations
182 | redis-001 | Reliability | Diagnostic Logs | Redis should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-monitor-diagnostic-settings
183 | redis-002 | Reliability | Availability Zones | Redis should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-high-availability
184 | redis-003 | Reliability | SLA | Redis should have a SLA | High | https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services?lang=1
185 | redis-004 | Security | Private Endpoint | Redis should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-private-link
186 | redis-005 | Reliability | SKU | Redis SKU | High | https://azure.microsoft.com/en-gb/pricing/details/cache/
187 | redis-006 | Operational Excellence | Naming Convention (CAF) | Redis Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
188 | redis-007 | Operational Excellence | Tags | Redis should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
189 | redis-008 | Security | SSL | Redis should not enable non SSL ports | High | https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-configure#access-ports
190 | redis-009 | Security | TLS | Redis should enforce TLS >= 1.2 | Low | https://learn.microsoft.com/en-us/azure/azure-cache-for-redis/cache-remove-tls-10-11
191 | sb-001 | Reliability | Diagnostic Logs | Service Bus should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/service-bus-messaging/monitor-service-bus#collection-and-routing
192 | sb-002 | Reliability | Availability Zones | Service Bus should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/service-bus-messaging/service-bus-outages-disasters#availability-zones
193 | sb-003 | Reliability | SLA | Service Bus should have a SLA | High | https://www.azure.cn/en-us/support/sla/service-bus/
194 | sb-004 | Security | Private Endpoint | Service Bus should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/service-bus-messaging/network-security
195 | sb-005 | Reliability | SKU | Service Bus SKU | High | https://azure.microsoft.com/en-us/pricing/details/service-bus/
196 | sb-006 | Operational Excellence | Naming Convention (CAF) | Service Bus Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
197 | sb-007 | Operational Excellence | Tags | Service Bus should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
198 | sb-008 | Security | Identity and Access Control | Service Bus should have local authentication disabled | Medium | https://learn.microsoft.com/en-us/azure/service-bus-messaging/service-bus-sas
199 | sigr-001 | Reliability | Diagnostic Logs | SignalR should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/azure-signalr/signalr-howto-diagnostic-logs
200 | sigr-002 | Reliability | Availability Zones | SignalR should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/azure-signalr/availability-zones
201 | sigr-003 | Reliability | SLA | SignalR should have a SLA | High | https://www.azure.cn/en-us/support/sla/signalr-service/
202 | sigr-004 | Security | Private Endpoint | SignalR should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/azure-signalr/howto-private-endpoints
203 | sigr-005 | Reliability | SKU | SignalR SKU | High | https://azure.microsoft.com/en-us/pricing/details/signalr-service/
204 | sigr-006 | Operational Excellence | Naming Convention (CAF) | SignalR Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
205 | sigr-007 | Operational Excellence | Tags | SignalR should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
206 | sql-001 | Reliability | Diagnostic Logs | SQL should have diagnostic settings enabled | Medium | 
207 | sql-004 | Security | Private Endpoint | SQL should have private endpoints enabled | High | 
208 | sql-006 | Operational Excellence | Naming Convention (CAF) | SQL Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
209 | sql-007 | Operational Excellence | Tags | SQL should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
210 | sql-008 | Security | TLS | SQL should enforce TLS >= 1.2 | Low | https://learn.microsoft.com/en-us/azure/azure-sql/database/connectivity-settings?view=azuresql&tabs=azure-portal#minimal-tls-version
211 | sqldb-001 | Reliability | Diagnostic Logs | SQL Database should have diagnostic settings enabled | Medium | 
212 | sqldb-002 | Reliability | Availability Zones | SQL Database should have availability zones enabled | High | 
213 | sqldb-003 | Reliability | SLA | SQL Database should have a SLA | High | 
214 | sqldb-005 | Reliability | SKU | SQL Database SKU | High | https://docs.microsoft.com/en-us/azure/azure-sql/database/service-tiers-vcore?tabs=azure-portal
215 | sqldb-006 | Operational Excellence | Naming Convention (CAF) | SQL Database Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
216 | sqldb-007 | Operational Excellence | Tags | SQL Database should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
217 | st-001 | Reliability | Diagnostic Logs | Storage should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/storage/blobs/monitor-blob-storage
218 | st-002 | Reliability | Availability Zones | Storage should have availability zones enabled | High | https://learn.microsoft.com/EN-US/azure/reliability/migrate-storage
219 | st-003 | Reliability | SLA | Storage should have a SLA | High | https://www.azure.cn/en-us/support/sla/storage/
220 | st-004 | Security | Private Endpoint | Storage should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/storage/common/storage-private-endpoints
221 | st-005 | Reliability | SKU | Storage SKU | High | https://learn.microsoft.com/en-us/rest/api/storagerp/srp_sku_types
222 | st-006 | Operational Excellence | Naming Convention (CAF) | Storage Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
223 | st-007 | Security | HTTPS Only | Storage Account should use HTTPS only | High | https://learn.microsoft.com/en-us/azure/storage/common/storage-require-secure-transfer
224 | st-008 | Operational Excellence | Tags | Storage Account should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
225 | st-009 | Security | TLS | Storage Account should enforce TLS >= 1.2 | Low | https://learn.microsoft.com/en-us/azure/storage/common/transport-layer-security-configure-minimum-version?tabs=portal
226 | vm-001 | Reliability | Diagnostic Logs | Virtual Machine should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/azure-monitor/agents/diagnostics-extension-windows-install
227 | vm-002 | Reliability | Availability Zones | Virtual Machine should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/virtual-machines/availability#availability-zones
228 | vm-003 | Reliability | SLA | Virtual Machine should have a SLA | High | https://www.microsoft.com/licensing/docs/view/Service-Level-Agreements-SLA-for-Online-Services?lang=1
229 | vm-006 | Operational Excellence | Naming Convention (CAF) | Virtual Machine Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
230 | vm-007 | Operational Excellence | Tags | Virtual Machine should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
231 | vm-008 | Reliability | Reliability | Virtual Machine should use managed disks | High | https://learn.microsoft.com/en-us/azure/architecture/checklist/resiliency-per-service#virtual-machines
232 | vm-009 | Reliability | Reliability | Virtual Machine should host application or database data on a data disk | Low | https://learn.microsoft.com/azure/virtual-machines/managed-disks-overview#data-disk
233 | vnet-001 | Reliability | Diagnostic Logs | Virtual Network should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/virtual-network/monitor-virtual-network#collection-and-routing
234 | vnet-002 | Reliability | Availability Zones | Virtual Network should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/virtual-network/virtual-networks-overview#virtual-networks-and-availability-zones
235 | vnet-006 | Operational Excellence | Naming Convention (CAF) | Virtual Network Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
236 | vnet-007 | Operational Excellence | Tags | Virtual Network should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
237 | vnet-008 | Security | Networking | Virtual Network: All Subnets should have a Network Security Group associated | High | https://learn.microsoft.com/azure/virtual-network/concepts-and-best-practices
238 | vnet-009 | Reliability | Reliability | Virtual NetworK should have at least two DNS servers assigned | High | https://learn.microsoft.com/en-us/azure/virtual-network/virtual-networks-name-resolution-for-vms-and-role-instances?tabs=redhat#specify-dns-servers
239 | wps-001 | Reliability | Diagnostic Logs | Web Pub Sub should have diagnostic settings enabled | Medium | https://learn.microsoft.com/en-us/azure/azure-web-pubsub/howto-troubleshoot-resource-logs
240 | wps-002 | Reliability | Availability Zones | Web Pub Sub should have availability zones enabled | High | https://learn.microsoft.com/en-us/azure/azure-web-pubsub/concept-availability-zones
241 | wps-003 | Reliability | SLA | Web Pub Sub should have a SLA | High | https://azure.microsoft.com/en-gb/support/legal/sla/web-pubsub/
242 | wps-004 | Security | Private Endpoint | Web Pub Sub should have private endpoints enabled | High | https://learn.microsoft.com/en-us/azure/azure-web-pubsub/howto-secure-private-endpoints
243 | wps-005 | Reliability | SKU | Web Pub Sub SKU | High | https://azure.microsoft.com/en-us/pricing/details/web-pubsub/
244 | wps-006 | Operational Excellence | Naming Convention (CAF) | Web Pub Sub Name should comply with naming conventions | Low | https://learn.microsoft.com/en-us/azure/cloud-adoption-framework/ready/azure-best-practices/resource-abbreviations
245 | wps-007 | Operational Excellence | Tags | Web Pub Sub should have tags | Low | https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/tag-resources?tabs=json
