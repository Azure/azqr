---
title: Releated Projects
description: Azure Quick Review difference to Azure Review Checklists and PSRule.Rules.Azure
weight: 6
---

**Azure Quick Review (azqr)** was created to address a very specific need we had back in 2022. Initially, we had to run three assessments to get a clear picture of various solutions in terms of SLAs, use of Availability Zones, and Diagnostic Settings. At the time, we were not aware of the existence of the [`review-checklist`](https://github.com/Azure/review-checklists) or [`PSRule.Rules.Azure`](https://github.com/Azure/PSRule.Rules.Azure).

When some of our peers saw the assessments we were able to deliver with the early bits of **Azure Quick Review (azqr)**, they asked us to add more checks (rules) and change the output format from markdown to Excel.

Also, many of our customers work in very restrictive environments. Therefore, the ability to run a self-contained binary, in any OS, that requires just read-only permissions was very important.

Moving forward to 2023, based on great feedback from both peers and customers, we moved the original repo to the Azure organization, added support for more services, fixed some issues and even added a Power BI template.

To be open we are giving **Azure Quick Review (azqr)** a "best effort" approach, and since we learned about [`PSRule.Rules.Azure`](https://github.com/Azure/PSRule.Rules.Azure), we are slowly trying to catch up with their great set of rules, with no intention to replace Bernie White's and collaborators outstanding work.

Long Story short, given the fact that each solution is built using different technologies and have a different set of features, there is room for both.
