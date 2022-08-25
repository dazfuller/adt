package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"testing"
)

func Test_modelEntry_getModelDependencies(t *testing.T) {
	fileContent, _ := os.ReadFile("../testdata/models/building.dtdl")
	var jsonContent jsonObject
	_ = json.Unmarshal(fileContent, &jsonContent)

	entry, _ := newModelEntry(jsonContent)
	dependencies := entry.getModelDependencies()

	if len(dependencies) != 1 {
		t.Fatalf("Expected 1 dependency, but got %d", len(dependencies))
	}

	expectedDependency := "dtmi:digitaltwins:testing:core:space;1"

	if dependencies[0] != expectedDependency {
		t.Fatalf("Dependency of '%s' not found, instead got '%s'", expectedDependency, dependencies[0])
	}
}

func Test_setModelDependencies(t *testing.T) {
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
		t.Fatalf("Unable to find building model")
	} else if len(buildingEntry.dependencies) == 0 {
		t.Fatalf("Building dependencies have not been set")
	}

	expectedDependencyId := "dtmi:digitaltwins:testing:core:space;1"

	if buildingEntry.dependencies[0].modelId != expectedDependencyId {
		t.Fatalf("Expected to find model dependency '%s', but found '%s", expectedDependencyId, buildingEntry.dependencies[0].modelId)
	}
}

func Test_sortModels(t *testing.T) {
	d := ModelDirectory{}
	_ = d.Set("../testdata/models")
	models, _ := d.getModels()

	setModelDependencies(models)

	sorted := sortModels(models)
	expectedSize := 5
	if len(sorted) != expectedSize {
		t.Fatalf("Expected a collection of %d elements, but got %d", expectedSize, len(sorted))
	}

	var meetingRoomIndex, roomIndex, spaceIndex int
	for i, entry := range sorted {
		if entry.modelId == "dtmi:digitaltwins:testing:core:meetingroom;1" {
			meetingRoomIndex = i
		} else if entry.modelId == "dtmi:digitaltwins:testing:core:room;1" {
			roomIndex = i
		} else if entry.modelId == "dtmi:digitaltwins:testing:core:space;1" {
			spaceIndex = i
		}
	}

	if spaceIndex != 0 {
		t.Fatalf("Space model should be the first item in the sorted collection, but it is an index %d", spaceIndex)
	}

	if !(spaceIndex < roomIndex) {
		t.Fatalf("Expected space model to be before the room model. space [%d], room [%d]", spaceIndex, roomIndex)
	}

	if !(roomIndex < meetingRoomIndex) {
		t.Fatalf("Expected room model to be before the meeting room model. room [%d], meeting room [%d]", roomIndex, meetingRoomIndex)
	}
}

func Test_sortModels_circular(t *testing.T) {
	if os.Getenv("TEST_CRASHER") == "1" {
		d := ModelDirectory{}
		_ = d.Set("../testdata/models")
		models, _ := d.getModels()

		setModelDependencies(models)

		// Create a circular dependency
		var roomEntry, meetingRoomEntry *modelEntry

		for _, entry := range models {
			if entry.modelId == "dtmi:digitaltwins:testing:core:room;1" {
				roomEntry = entry
			} else if entry.modelId == "dtmi:digitaltwins:testing:core:meetingroom;1" {
				meetingRoomEntry = entry
			}
		}

		roomEntry.dependencies = append(roomEntry.dependencies, meetingRoomEntry)
		_ = sortModels(models)
	}

	cmd := exec.Command(os.Args[0], "-test.run=Test_sortModels_circular")
	cmd.Env = append(os.Environ(), "TEST_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}

	t.Fatalf("process ran with error %v, but wanted exit status 1", err)
}
