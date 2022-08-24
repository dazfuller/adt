package cli

import (
	"fmt"
)

// ListModels retrieves all models which have been created against the Azure Digital Twin endpoint using the
// authentication method provided
func ListModels(endpoint string, method *AuthenticationMethod) error {
	config, _ := newTwinConfiguration(endpoint, method)
	client := newClient(config)

	models, err := client.listModels()
	if err != nil {
		return fmt.Errorf("an error occured listing models in the twin: %s", err)
	}

	for _, model := range models {
		fmt.Println(model.modelId)
	}

	return nil
}

// ClearModels will remove all models which have been created against the Azure Digital Twin endpoint using the
// authentication method provided
func ClearModels(endpoint string, method *AuthenticationMethod) error {
	config, _ := newTwinConfiguration(endpoint, method)
	client := newClient(config)

	models, err := client.listModels()
	if err != nil {
		return fmt.Errorf("an error occured retrieving models from the twin: %s", err)
	}

	setModelDependencies(models)
	sorted := sortModels(models)

	reversed := make([]*modelEntry, len(sorted))

	for i, j := len(sorted)-1, 0; i >= 0; i, j = i-1, j+1 {
		reversed[j] = sorted[i]
	}

	err = client.clearModels(reversed)
	if err != nil {
		return fmt.Errorf("unable to clear models from the digital twin: %s", err)
	}

	return nil
}
