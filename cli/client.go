package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

const (
	maxModelsApiLimit = 250          // Maximum number of models allowed per API request when adding models
	maxModelsPerBatch = 40           // Maximum number of models allowed per API request when adding models in batches
	apiVersion        = "2020-10-31" // Digital Twin Rest API version to use
)

// pagedDigitalTwinsModelDataCollection defines a paged response from the Azure Digital Twin GET model API. It contains
// a list of digital twin models, and a continuation token to retrieve more results
type pagedDigitalTwinsModelDataCollection struct {
	NextLink string       `json:"nextLink"` // The continuation token if provided
	Value    []jsonObject `json:"value"`    // Collection of models
}

// client managed connecting to the Azure Digital Twin resource
type client struct {
	configuration *twinConfiguration
	httpClient    *http.Client
}

// Creates a new instance of the client type
func newClient(configuration *twinConfiguration) *client {
	return &client{
		configuration: configuration,
		httpClient:    &http.Client{},
	}
}

// Gets the URL required to access the Azure Digital Twin model API with the api-version information. It takes an
// optional modelId for when a single model is being accessed, and an optional parameters parameter for additional
// parameters needed to be passed to the API
func (client *client) getModelUrl(modelId *string, parameters *map[string]string) string {
	endpoint := client.configuration.endpoint
	if modelId == nil || len(*modelId) == 0 {
		endpoint.Path = "/models"
	} else {
		endpoint.Path = fmt.Sprintf("/models/%s", *modelId)
	}

	params := url.Values{}
	params.Add("api-version", apiVersion)

	if parameters != nil {
		for k, v := range *parameters {
			params.Add(k, v)
		}
	}

	endpoint.RawQuery = params.Encode()
	return endpoint.String()
}

// Gets all the models from the Azure Digital Twin instance
func (client *client) listModels() ([]*modelEntry, error) {
	results := make([]*modelEntry, 0)

	token, err := client.configuration.getBearerToken()
	if err != nil {
		return nil, err
	}

	endpoint := client.getModelUrl(nil, &map[string]string{"includeModelDefinition": "true"})

	for {
		req, _ := http.NewRequest("GET", endpoint, nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.Token))
		req.Header.Set("Accept", "application/json")

		log.Printf("Retrieving models from: %s", endpoint)

		resp, err := client.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve data from %s\n%s", endpoint, err)
		} else if resp.StatusCode != 200 {
			return nil, handleResponseError(resp)
		}

		var pagedResult pagedDigitalTwinsModelDataCollection
		respContent, _ := io.ReadAll(resp.Body)
		_ = json.Unmarshal(respContent, &pagedResult)

		for i := range pagedResult.Value {
			entry, _ := newModelEntry(pagedResult.Value[i])
			results = append(results, entry)
		}

		endpoint = pagedResult.NextLink

		if len(endpoint) == 0 {
			break
		}
	}

	return results, nil
}

// Removes all models which have been added to the Azure Digital Twin instance
func (client *client) clearModels(models []*modelEntry) error {
	token, err := client.configuration.getBearerToken()
	if err != nil {
		return err
	}

	for i, entry := range models {
		endpoint := client.getModelUrl(&entry.modelId, nil)

		req, _ := http.NewRequest("DELETE", endpoint, nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.Token))

		log.Printf("Deleting entry %d/%d: %s", i+1, len(models), entry.modelId)
		resp, err := client.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("unable to delete model %s\n%s", entry.modelId, err)
		} else if resp.StatusCode != 204 {
			return handleResponseError(resp)
		}
	}

	return nil
}

// In the event of an API error response, this handles it at returns an error detailing the error
func handleResponseError(resp *http.Response) error {
	respContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("non-success status code returned: %d", resp.StatusCode)
	} else {
		var respError map[string]interface{}
		_ = json.Unmarshal(respContent, &respError)
		return fmt.Errorf("non-success status code returned: %d\n%v", resp.StatusCode, respError["error"])
	}
}
