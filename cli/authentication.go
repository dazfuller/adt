package cli

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"log"
	"net/url"
)

const (
	resourceId   = "https://digitaltwins.azure.net"    // The azure resource identifier for Azure Digital Twins
	authorityUrl = "https://login.microsoftonline.com" // Authority URL for user authentication
)

// AuthenticationMethod defined how the application will authenticate with an Azure Digital Twin instance
type AuthenticationMethod struct {
	UseAzureCli  bool   // Indicates if the Azure CLI credential should be used
	TenantId     string // When using client credentials, specifies the Azure tenant to authenticate against
	ClientId     string // The id of the client used for client credential authentication
	ClientSecret string // The secret of the client used for client credential authentication
}

// Describes all the configuration required for interacting with an Azure Digital Twin instance
type twinConfiguration struct {
	endpoint     url.URL  // The URL of the Azure Digital Twin instance
	useAzureCli  bool     // Indicates if the Azure CLI credential should be used
	tenantId     string   // When using client credentials, specifies the Azure tenant to authenticate against
	clientId     string   // The id of the client used for client credential authentication
	clientSecret string   // The secret of the client used for client credential authentication
	scopes       []string // The scopes to create an authentication token for
	authorityUrl url.URL  // Authority URL required for authenticating the user
}

// Creates a new twinConfiguration instance using the endpoint and AuthenticationMethod information provided
func newTwinConfiguration(endpoint string, authenticationMethod *AuthenticationMethod) (*twinConfiguration, error) {
	authority, _ := url.Parse(authorityUrl)
	var scopes []string

	if authenticationMethod.UseAzureCli {
		scopes = []string{resourceId}
	} else {
		scopes = []string{fmt.Sprintf("%s/.default", resourceId)}
	}

	config := twinConfiguration{
		scopes:       scopes,
		authorityUrl: *authority,
		useAzureCli:  authenticationMethod.UseAzureCli,
		tenantId:     authenticationMethod.TenantId,
		clientId:     authenticationMethod.ClientId,
		clientSecret: authenticationMethod.ClientSecret,
	}

	err := config.setAdtEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// Validates and sets the Azure Digital Twin endpoint for the twinConfiguration instance. Attempting to set
// an invalid url will result in an error.
func (configuration *twinConfiguration) setAdtEndpoint(endpoint string) error {
	adtEndpointUrl, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("unable to set digital twin endpoint: %s", err)
	}

	configuration.endpoint = *adtEndpointUrl
	return nil
}

// Gets a bearer token for the twinConfiguration instance
func (configuration *twinConfiguration) getBearerToken() (*azcore.AccessToken, error) {
	var credentials azcore.TokenCredential
	var err error
	if configuration.useAzureCli {
		credentials, err = azidentity.NewAzureCLICredential(nil)
	} else {
		credentials, err = azidentity.NewClientSecretCredential(configuration.tenantId, configuration.clientId, configuration.clientSecret, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to create credentials: %s", err)
	}

	ctx := context.Background()
	tokenRequestOptions := policy.TokenRequestOptions{Scopes: configuration.scopes}

	log.Print("Getting bearer token")

	token, err := credentials.GetToken(ctx, tokenRequestOptions)
	if err != nil {
		return nil, fmt.Errorf("unable to acquire token for Azure Digital Twin scope: %s", err)
	}

	return &token, nil
}
