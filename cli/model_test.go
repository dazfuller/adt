package cli

import (
	"encoding/json"
	"os"
	"testing"
)

func Test_modelEntry_getModelDependencies(t *testing.T) {
	fileContent, _ := os.ReadFile("../testdata/models/building.dtdl")
	var jsonContent jsonObject
	_ = json.Unmarshal(fileContent, &jsonContent)

	entry, _ := newModelEntry(jsonContent)
	dependencies := entry.getModelDependencies()

	if len(dependencies) != 1 {
		t.Errorf("Expected 1 dependency, but got %d", len(dependencies))
	}

	expectedDependency := "dtmi:digitaltwins:testing:core:space;1"

	if dependencies[0] != expectedDependency {
		t.Errorf("Dependency of '%s' not found, instead got '%s'", expectedDependency, dependencies[0])
	}
}

func Test_modelEntry_setModelDependencies(t *testing.T) {
	d := ModelDirectory{}
	_ = d.Set("../testdata/models")
	models, _ := d.getModels()

	setModelDependencies(models)

	var buildingEntry *modelEntry
	for _, entry := range models {
		if entry.modelId == "dtmi:digitaltwins:testing:core:building;1" {
			buildingEntry = entry
		}
	}

	if buildingEntry == nil {
		t.Errorf("Unable to find building model")
	} else if len(buildingEntry.dependencies) == 0 {
		t.Errorf("Building dependencies have not been set")
	}

	expectedDependencyId := "dtmi:digitaltwins:testing:core:space;1"

	if buildingEntry.dependencies[0].modelId != expectedDependencyId {
		t.Errorf("Expected to find model dependency '%s', but found '%s", expectedDependencyId, buildingEntry.dependencies[0].modelId)
	}
}
