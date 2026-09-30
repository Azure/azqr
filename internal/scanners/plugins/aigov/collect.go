// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package aigov

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azqr/internal/scanners/plugins/aigov/metrics"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/monitor/query/azmetrics"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cognitiveservices/armcognitiveservices/v2"
	"github.com/rs/zerolog/log"
)

// collectedAccount is one account with everything collected for it.
type collectedAccount struct {
	Resource         *models.Resource
	SubscriptionName string
	Data             *metrics.AccountData
	Deployments      map[string]deploymentInfo
}

func (c *collectedAccount) deploymentInfo(name string) deploymentInfo {
	if di, ok := c.Deployments[name]; ok {
		return di
	}
	return newDefaultDeploymentInfo()
}

// collectAccounts resolves each account's metric plan from its own metric
// definitions, collects the window for accounts that share a plan in one batch
// query set, and lists deployments for the accounts that returned data.
func (s *AIGovScanner) collectAccounts(ctx context.Context, cred azcore.TokenCredential, deploymentsClient *armcognitiveservices.DeploymentsClient, subscriptionName, subscriptionID, region string, resources []*models.Resource) []*collectedAccount {
	if len(resources) == 0 {
		return nil
	}

	plans := resolvePlans(ctx, cred, subscriptionID, resources)

	endpoint := fmt.Sprintf("https://%s.metrics.monitor.azure.com", region)
	client, err := azmetrics.NewClient(endpoint, cred, nil)
	if err != nil {
		log.Debug().Err(err).Str("region", region).Msg("failed to create metrics client")
		return nil
	}

	bySignature := map[string][]string{}
	signaturePlan := map[string]metrics.Plan{}
	for _, r := range resources {
		p := plans[r.ID]
		sig := p.Signature()
		bySignature[sig] = append(bySignature[sig], r.ID)
		if _, ok := signaturePlan[sig]; !ok {
			signaturePlan[sig] = p
		}
	}

	end := time.Now().UTC().Truncate(5 * time.Minute)
	data := map[string]*metrics.AccountData{}
	for sig, ids := range bySignature {
		for id, d := range metrics.CollectBatch(ctx, client, subscriptionID, ids, signaturePlan[sig], end, metrics.MaxWindowDays) {
			// Per-account notes (e.g. metrics excluded) come from its own plan.
			d.Plan = plans[id]
			data[id] = d
		}
	}

	accounts := make([]*collectedAccount, 0, len(resources))
	byID := map[string]*collectedAccount{}
	for _, r := range resources {
		d := data[r.ID]
		if d == nil {
			continue
		}
		acct := &collectedAccount{Resource: r, SubscriptionName: subscriptionName, Data: d}
		accounts = append(accounts, acct)
		byID[r.ID] = acct
	}

	var withData []*collectedAccount
	for _, acct := range accounts {
		if len(acct.Data.Buckets) > 0 {
			withData = append(withData, acct)
		}
	}
	var mu sync.Mutex
	forEachBounded(withData, func(a *collectedAccount) {
		deployments, fetchErr := s.getDeployments(ctx, deploymentsClient, a.Resource.ResourceGroup, a.Resource.Name)
		if fetchErr != nil {
			log.Debug().Err(fetchErr).Str("resourceID", a.Resource.ID).Msg("Failed to get deployments")
			return
		}
		m := make(map[string]deploymentInfo, len(deployments))
		for _, d := range deployments {
			if d.Name != nil {
				m[*d.Name] = extractDeploymentInfo(d)
			}
		}
		mu.Lock()
		a.Deployments = m
		mu.Unlock()
	})
	return accounts
}

// forEachBounded runs fn for every item with at most
// maxConcurrentDeploymentFetches in flight and waits for all to finish.
func forEachBounded[T any](items []T, fn func(T)) {
	sem := make(chan struct{}, maxConcurrentDeploymentFetches)
	var wg sync.WaitGroup
	for _, item := range items {
		wg.Add(1)
		go func(it T) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			fn(it)
		}(item)
	}
	wg.Wait()
}

// resolvePlans lists metric definitions per account (bounded concurrency) and
// falls back to the documented default plan when listing fails.
func resolvePlans(ctx context.Context, cred azcore.TokenCredential, subscriptionID string, resources []*models.Resource) map[string]metrics.Plan {
	plans := make(map[string]metrics.Plan, len(resources))
	var mu sync.Mutex
	forEachBounded(resources, func(res *models.Resource) {
		defs, err := metrics.FetchDefinitions(ctx, cred, subscriptionID, res.ID)
		var plan metrics.Plan
		if err != nil || len(defs) == 0 {
			log.Debug().Err(err).Str("resourceID", res.ID).Msg("metric definitions unavailable; using default plan")
			plan = metrics.DefaultPlan()
		} else {
			plan = metrics.ResolvePlan(defs)
		}
		mu.Lock()
		plans[res.ID] = plan
		mu.Unlock()
	})
	return plans
}
