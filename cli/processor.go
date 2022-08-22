package cli

import (
	"fmt"
	"log"
)

func ListModels() {
	config, _ := newTwinConfiguration("https://testdevtwindaz.api.neu.digitaltwins.azure.net", authenticationMethod{useAzureCli: true})
	client := newClient(config)

	models, err := client.listModels()
	if err != nil {
		log.Fatalf("An error occured listing models in the twin: %s", err)
	}

	fmt.Printf("Found %d model\n", len(models))
	for _, model := range models {
		fmt.Println(model.modelId)
	}
}
