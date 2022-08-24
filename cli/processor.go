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

	if len(models) == 0 {
		fmt.Println("No models to remove")
		return nil
	}

	setModelDependencies(models)
	sorted := sortModels(models)

	reversed := make([]*modelEntry, len(sorted))

	for i, j := len(sorted)-1, 0; i >= 0; i, j = i-1, j+1 {
		reversed[j] = sorted[i]
	}

	fmt.Printf("Removing %d model(s) from the digital twin instance\n", len(reversed))

	err = client.clearModels(reversed)
	if err != nil {
		return fmt.Errorf("unable to clear models from the digital twin: %s", err)
	}

	fmt.Println("Successfully cleared all models from the digital twin instance")

	return nil
}

// UploadModels will read all model files (.json and .dtdl files) in a given path recursively, and then attempt to
// upload them to the Azure Digital Twin instance
func UploadModels(endpoint string, method *AuthenticationMethod, source ModelDirectory) error {
	config, _ := newTwinConfiguration(endpoint, method)
	client := newClient(config)

	models, err := source.getModels()
	if err != nil {
		return fmt.Errorf("unable to retrieve models from %s: %s", source.Path, err)
	}

	if len(models) == 0 {
		return fmt.Errorf("No models found to upload\n")
	}

	setModelDependencies(models)
	sorted := sortModels(models)

	fmt.Printf("Uploading %d models to the digital twin instance\n", len(sorted))

	err = client.uploadModels(sorted)
	if err != nil {
		return fmt.Errorf("unable to upload models: %s", err)
	}

	fmt.Printf("Successfully uploaded models from %s\n", source.Path)

	return nil
}
