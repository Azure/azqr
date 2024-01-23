// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package graph

import (
	"context"

	"github.com/Azure/azqr/internal/ref"
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
			Top:          ref.Of(int32(1000)),
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
		results, err := q.client.Resources(ctx, request, nil)
		if err == nil {
			result.Count = *results.TotalRecords
			result.Data = append(result.Data, results.Data.([]interface{})...)
			skipToken = results.SkipToken
		} else {
			log.Fatal().Err(err).Msg("Failed to run Resource Graph query")
			return nil
		}
	}
	return &result
}
