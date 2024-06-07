// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package graph

import (
	"context"
	"time"

	"github.com/Azure/azqr/internal/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	arg "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph"
	"github.com/rs/zerolog/log"
)

type (
	GraphQuery struct {
		client *arg.Client
	}

	GraphResult struct {
		Count int64
		Data  []interface{}
	}
)

func NewGraphQuery(cred azcore.TokenCredential) *GraphQuery {
	client, err := arg.NewClient(cred, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Resource Graph client")
		return nil
	}
	return &GraphQuery{
		client: client,
	}
}

func (q *GraphQuery) Query(ctx context.Context, query string, subscriptionIDs []*string) *GraphResult {
	format := arg.ResultFormatObjectArray
	request := arg.QueryRequest{
		Subscriptions: subscriptionIDs,
		Query:         &query,
		Options: &arg.QueryRequestOptions{
			ResultFormat: &format,
			Top:          to.Ptr(int32(1000)),
		},
	}

	if q.client == nil {
		log.Fatal().Msg("Resource Graph client not initialized")
	}

	result := GraphResult{}
	result.Data = make([]interface{}, 0)
	var skipToken *string = nil
	for ok := true; ok; ok = skipToken != nil {
		request.Options.SkipToken = skipToken
		// Run the query and get the results
		results, err := q.retry(ctx, 3, 10*time.Second, request)
		if err == nil {
			result.Count = *results.TotalRecords
			result.Data = append(result.Data, results.Data.([]interface{})...)
			skipToken = results.SkipToken
		} else {
			log.Fatal().Err(err).Msgf("Failed to run Resource Graph query: %s", query)
			return nil
		}
	}
	return &result
}

func (q *GraphQuery) retry(ctx context.Context, attempts int, sleep time.Duration, request arg.QueryRequest) (arg.ClientResourcesResponse, error) {
	var err error
	for i := 0; ; i++ {
		res, err := q.client.Resources(ctx, request, nil)
		if err == nil {
			return res, nil
		}

		// if shouldSkipError(err) {
		// 	return []scanners.AzureServiceResult{}, nil
		// }

		errAsString := err.Error()

		if i >= (attempts - 1) {
			log.Info().Msgf("Retry limit reached. Error: %s", errAsString)
			break
		}

		log.Debug().Msgf("Retrying after error: %s", errAsString)

		time.Sleep(sleep)
		sleep *= 2
	}
	return arg.ClientResourcesResponse{}, err
}
