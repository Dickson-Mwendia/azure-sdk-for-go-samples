// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package cognitiveservices

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/config"
	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/entitysearch"
	"github.com/Azure/go-autorest/autorest"
)

func getEntitySearchClient(accountName string) entitysearch.EntitiesClient {
	apiKey := getFirstKey(accountName)
	entitySearchClient := entitysearch.NewEntitiesClient()
	csAuthorizer := autorest.NewCognitiveServicesAuthorizer(apiKey)
	entitySearchClient.Authorizer = csAuthorizer
	entitySearchClient.AddToUserAgent(config.UserAgent())
	return entitySearchClient
}

//SearchEntities retunrs a list of entities
func SearchEntities(accountName string) (*entitysearch.Entities, error) {
	entitySearchClient := getEntitySearchClient(accountName)
	query := "tom cruise"
	market := "en-us"
	searchResponse, err := entitySearchClient.Search(
		context.Background(),            // context
		query,                           // query keyword
		"",                              // Accept-Language header
		"",                              // pragma header
		"",                              // User-Agent header
		"",                              // X-MSEdge-ClientID header
		"",                              // X-MSEdge-ClientIP header
		"",                              // X-Search-Location header
		"",                              // country code
		market,                          // market
		[]entitysearch.AnswerType{},     // response filter
		[]entitysearch.ResponseFormat{}, // response format
		entitysearch.Strict,             // safe search
		"",                              // set lang
	)
	if err != nil {
		return nil, err
	}

	return searchResponse.Entities, nil
}
