// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mysql/armmysql"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "eastus"
	resourceGroupName = "sample-resource-group"
	serverName        = "sample2server"
	databaseName      = "sample-database"
)

func main() {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	server, err := createServer(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("mysql server:", *server.ID)

	database, err := createDatabase(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("mysql database:", *database.ID)

	database, err = getDatabase(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get mysql database:", *database.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createServer(ctx context.Context, cred azcore.TokenCredential) (*armmysql.Server, error) {
	serversClient := armmysql.NewServersClient(subscriptionID, cred, nil)

	pollerResp, err := serversClient.BeginCreate(
		ctx,
		resourceGroupName,
		serverName,
		armmysql.ServerForCreate{
			Location: to.StringPtr(location),
			Properties: &armmysql.ServerPropertiesForCreate{
				CreateMode: armmysql.CreateModeDefault.ToPtr(),
			},
			SKU: &armmysql.SKU{
				Name:     to.StringPtr("GP_Gen5_2"),
				Tier:     armmysql.SKUTierGeneralPurpose.ToPtr(),
				Capacity: to.Int32Ptr(2),
				Family:   to.StringPtr("Gen5"),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.Server, nil
}

func createDatabase(ctx context.Context, cred azcore.TokenCredential) (*armmysql.Database, error) {
	databaseClient := armmysql.NewDatabasesClient(subscriptionID, cred, nil)
	pollerResp, err := databaseClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		serverName,
		databaseName,
		armmysql.Database{
			Properties: &armmysql.DatabaseProperties{
				Charset:   to.StringPtr("utf8"),
				Collation: to.StringPtr("utf8_general_ci"),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 30*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.Database, nil
}

func getDatabase(ctx context.Context, cred azcore.TokenCredential) (*armmysql.Database, error) {
	databaseClient := armmysql.NewDatabasesClient(subscriptionID, cred, nil)
	resp, err := databaseClient.Get(ctx, resourceGroupName, serverName, databaseName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Database, nil
}

func createResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*armresources.ResourceGroup, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)

	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.StringPtr(location),
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context, cred azcore.TokenCredential) (*http.Response, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.RawResponse, nil
}
