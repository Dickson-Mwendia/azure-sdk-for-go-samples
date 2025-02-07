// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package resources

import (
	"context"
	"fmt"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/services/internal/iam"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
)

func getDeploymentsClient() resources.DeploymentsClient {
	deployClient := resources.NewDeploymentsClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	deployClient.Authorizer = a
	deployClient.AddToUserAgent(config.UserAgent())
	return deployClient
}

// CreateDeployment creates a template deployment using the
// referenced JSON files for the template and its parameters
func CreateDeployment(ctx context.Context, deploymentName string, template, params *map[string]interface{}) (de resources.DeploymentExtended, err error) {
	deployClient := getDeploymentsClient()
	future, err := deployClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		deploymentName,
		resources.Deployment{
			Properties: &resources.DeploymentProperties{
				Template:   template,
				Parameters: params,
				Mode:       resources.Incremental,
			},
		},
	)
	if err != nil {
		return de, fmt.Errorf("cannot create deployment: %v", err)
	}

	err = future.WaitForCompletionRef(ctx, deployClient.Client)
	if err != nil {
		return de, fmt.Errorf("cannot get the create deployment future respone: %v", err)
	}

	return future.Result(deployClient)
}

// ValidateDeployment validates the template deployments and their
// parameters are correct and will produce a successful deployment.GetResource
func ValidateDeployment(ctx context.Context, deploymentName string, template, params *map[string]interface{}) (valid resources.DeploymentValidateResult, err error) {
	deployClient := getDeploymentsClient()
	return deployClient.Validate(ctx,
		config.GroupName(),
		deploymentName,
		resources.Deployment{
			Properties: &resources.DeploymentProperties{
				Template:   template,
				Parameters: params,
				Mode:       resources.Incremental,
			},
		})
}
