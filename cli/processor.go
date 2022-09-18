package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
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

// DownloadModels reads all models from the Digital Twin instance into the output location using the fileExtension
// specified.
//
// The download structure will be based on the model name structure broken apart by the colon and the
// semicolon, and so a model id of "dtmi:rec33:architectural:building;1" will become the following path
// "dtmi/rec33/architectural/building_1.dtdl" (assuming a file extension of 'dtdl')
func DownloadModels(endpoint string, method *AuthenticationMethod, output ModelDirectory, fileExtension string) error {
	// Validate the file extension
	fileExtensionLower := strings.TrimPrefix(strings.ToLower(fileExtension), ".")
	if fileExtensionLower != "json" && fileExtensionLower != "dtdl" {
		return fmt.Errorf("file extension '%s' is not valid, only 'json' or 'dtdl' should be provided", fileExtensionLower)
	}

	config, _ := newTwinConfiguration(endpoint, method)
	client := newClient(config)

	models, err := client.listModels()
	if err != nil {
		return fmt.Errorf("an error occured listing models in the twin: %s", err)
	}

	// If there's no models to download then exit here
	if len(models) == 0 {
		return nil
	}

	// Clear anything in the output path
	err = os.RemoveAll(output.Path)
	if err != nil {
		return fmt.Errorf("unable to clear output directory %s. %s", output, err)
	}

	err = os.Mkdir(output.Path, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to create output directory %s. %s", output, err)
	}

	// Process each model
	for _, model := range models {
		nameParts := strings.Split(model.modelId, ":")
		dirParts := nameParts[:len(nameParts)-1]

		// Lower case the path
		for i := range dirParts {
			dirParts[i] = strings.ToLower(dirParts[i])
		}

		// Generate the file name, output directory, and full output path
		filename := fmt.Sprintf("%s.%s", strings.ReplaceAll(nameParts[len(nameParts)-1], ";", "_"), fileExtensionLower)
		outputDir := filepath.Join(output.Path, filepath.Join(dirParts...))
		outputFilePath := filepath.Join(outputDir, filename)

		err = os.MkdirAll(outputDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("unable to create directory %s: %s", outputDir, err)
		}

		modelContent, err := model.model.ToJson()
		if err != nil {
			return fmt.Errorf("unable to parse content of model %s. %s", model.modelId, err)
		}

		log.Printf("Writing model %s to %s", model.modelId, outputFilePath)
		err = os.WriteFile(outputFilePath, modelContent, os.ModePerm)
		if err != nil {
			return fmt.Errorf("unable to write content of model %s to %s. %s", model.modelId, outputFilePath, err)
		}
	}

	return nil
}
