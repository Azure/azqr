### README - DISCLAIMER

⚠️ **IMPORTANT: This workbook does NOT represent Service Level Indicators (SLIs), Service Level Objectives (SLOs), or Service Level Agreement (SLA) compliance metrics.**

#### What This Report Shows

This workbook displays the **Event-Free %** metric, which represents the percentage of time that customer Azure resources were **not affected** by:
- Service Issue events (Azure platform incidents)
- Planned Maintenance events

#### What This Report Does NOT Show

- **Not an SLI**: These metrics are not Service Level Indicators. They do not measure actual service availability, performance, or quality.
- **Not SLO/SLA Compliance**: This data does not reflect whether services met their Service Level Objectives or Service Level Agreement commitments.
- **Not Service Availability**: A resource showing 100% time without events does not mean it was available 100% of the time. It only means it was not impacted by reported Service Health events.
- **Not All Outages**: This only tracks Azure Service Health events. It does not include:
  - Application-level failures
  - Customer-caused outages
  - Network connectivity issues outside Azure
  - Issues not reported through Service Health

#### How to Interpret the Data

The **Event-Free %** metric is calculated based on:
1. Resolved Service Health events (Service Issues and Planned Maintenance) within the selected time range
2. The duration each event impacted specific Azure resources
3. Resources with no Service Health impact show 100%

**Use this workbook to:**
- Understand which resource types were most frequently affected by Azure Service Health events
- Identify patterns in service disruptions across subscriptions and regions
- Assess the impact of Azure platform events on your resources

**Do NOT use this workbook to:**
- Measure actual service availability or uptime
- Validate SLA compliance
- Make decisions based solely on these metrics without considering actual service performance data

---

#### Support Disclaimer

**This sample workbook is not supported under any Microsoft standard support program or service.** This sample workbook and scripts are provided AS IS without warranty of any kind. Microsoft further disclaims all implied warranties including, without limitation, any implied warranties of merchantability or of fitness for a particular purpose. The entire risk arising out of the use or performance of the sample scripts and documentation remains with you. In no event shall Microsoft, its authors, or anyone else involved in the creation, production, or delivery of the scripts be liable for any damages whatsoever (including, without limitation, damages for loss of business profits, business interruption, loss of business information, or other pecuniary loss) arising out of the use of or inability to use the sample scripts or documentation, even if Microsoft has been advised of the possibility of such damages.
