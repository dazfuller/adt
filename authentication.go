package adt

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"net/url"
)

const (
	resourceId   = "https://digitaltwins.azure.net"
	authorityUrl = "https://login.microsoftonline.com"
	apiVersion   = "2020-10-31"
)

type authenticationMethod struct {
	useAzureCli  bool
	tenantId     string
	clientId     string
	clientSecret string
}

type twinConfiguration struct {
	Url          url.URL
	useAzureCli  bool
	clientId     string
	clientSecret string
	tenantId     string
	scopes       []string
	authorityUrl url.URL
}

func newTwinConfiguration(endpoint string, authenticationMethod authenticationMethod) (*twinConfiguration, error) {
	authority, _ := url.Parse(authorityUrl)
	var scopes []string

	if authenticationMethod.useAzureCli {
		scopes = []string{resourceId}
	} else {
		scopes = []string{fmt.Sprintf("%s/.default", resourceId)}
	}

	config := twinConfiguration{
		scopes:       scopes,
		authorityUrl: *authority,
		useAzureCli:  authenticationMethod.useAzureCli,
		tenantId:     authenticationMethod.tenantId,
		clientId:     authenticationMethod.clientId,
		clientSecret: authenticationMethod.clientSecret,
	}

	err := config.setAdtEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (configuration *twinConfiguration) setAdtEndpoint(endpoint string) error {
	adtEndpointUrl, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("unable to set digital twin endpoint: %s", err)
	}

	configuration.Url = *adtEndpointUrl
	return nil
}

func (configuration *twinConfiguration) getBearerToken() (*string, error) {
	var credentials azcore.TokenCredential
	var err error
	if configuration.useAzureCli {
		credentials, err = azidentity.NewAzureCLICredential(nil)
	} else {
		credentials, err = azidentity.NewClientSecretCredential(configuration.tenantId, configuration.clientSecret, configuration.clientSecret, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to create credentials: %s", err)
	}

	ctx := context.Background()
	tokenRequestOptions := policy.TokenRequestOptions{Scopes: configuration.scopes}

	token, err := credentials.GetToken(ctx, tokenRequestOptions)
	if err != nil {
		return nil, fmt.Errorf("unable to acquire token for Azure Digital Twin scope: %s", err)
	}

	return &token.Token, nil
}
