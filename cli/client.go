package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	maxModelsApiLimit = 250          // Maximum number of models allowed per API request when adding models
	maxModelsPerBatch = 40           // Maximum number of models allowed per API request when adding models in batches
	apiVersion        = "2020-10-31" // Digital Twin Rest API version to use
)

type pagedDigitalTwinsModelDataCollection struct {
	NextLink string       `json:"nextLink"`
	Value    []jsonObject `json:"value"`
}

type client struct {
	configuration *twinConfiguration
	httpClient    *http.Client
}

func newClient(configuration *twinConfiguration) *client {
	return &client{
		configuration: configuration,
		httpClient:    &http.Client{},
	}
}

func (client *client) getModelUrl(modelId *string, parameters *map[string]string) string {
	endpoint := client.configuration.endpoint
	if modelId == nil || len(*modelId) == 0 {
		endpoint.Path = "/models"
	} else {
		endpoint.Path = fmt.Sprintf("/models/%s", modelId)
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

		resp, err := client.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("snable to retrieve data from %s\n%s", endpoint, err)
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
