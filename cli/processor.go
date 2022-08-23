package cli

import (
	"fmt"
)

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
