package cli

import (
	"strings"
	"testing"
)

func errorText(s string) *string {
	return &s
}

func assertExpectedError(t *testing.T, actual error, expected *string) {
	if expected == nil && actual != nil {
		t.Errorf("Expected nil error, but got %v", actual)
	} else if expected != nil {
		if actual == nil {
			t.Errorf("Expected error containing '%s', but got nil", *expected)
		} else if !strings.Contains(actual.Error(), *expected) {
			t.Errorf("Expected error containing '%s', but got '%s'", *expected, actual)
		}
	}
}

func TestModelDirectory_Set(t *testing.T) {
	tests := []struct {
		name          string
		sourcePath    string
		expectedError *string
	}{
		{name: "NonexistentPath", sourcePath: "../testdata/invalid", expectedError: errorText("the specified path does not exist")},
		{name: "InvalidDirectory", sourcePath: "../testdata/models/building.dtdl", expectedError: errorText("the specified path is not a directory")},
		{name: "ValidPath", sourcePath: "../testdata/models", expectedError: nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := ModelDirectory{}
			err := d.Set(test.sourcePath)

			assertExpectedError(t, err, test.expectedError)
		})
	}
}

func TestModelDirectory_String(t *testing.T) {
	input := "../testdata/models"
	d := ModelDirectory{}
	_ = d.Set(input)

	if d.String() != "../testdata/models" {
		t.Errorf("Expected %s, but got %s", input, d.String())
	}
}

func TestModelDirectory_getModels(t *testing.T) {
	input := "../testdata/models"
	d := ModelDirectory{}
	_ = d.Set(input)

	models, err := d.getModels()
	if err != nil {
		t.Errorf("Expected a collection of models, but received error: %s", err)
	} else if len(models) != 3 {
		t.Errorf("Expected 3 models but got %d", len(models))
	}

	expectedIds := []string{"dtmi:digitaltwins:testing:core:building;1", "dtmi:digitaltwins:testing:core:level;1", "dtmi:digitaltwins:testing:core:room;1"}
	for _, id := range expectedIds {
		found := false
		for _, m := range models {
			if m.modelId == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("id %s was not found in the collected models", id)
		}
	}
}
