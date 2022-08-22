package cli

import "fmt"

type processingStatus int

const (
	none       processingStatus = iota // Indicates that a model has not yet been processed ar is in processing
	processing                         // The current entry is being evaluated
	processed                          // The entry has been fully processed
)

// Defines a structure for a JSON object
type jsonObject map[string]interface{}

// Retrieves the model id from the current object
func (object *jsonObject) getModelId() (*string, error) {
	var idToken interface{}
	var ok bool

	if idToken, ok = (*object)["@id"]; !ok {
		if idToken, ok = (*object)["id"]; !ok {
			return nil, fmt.Errorf("unable to find 'id' or '@id' in the model")
		}
	}
	id := idToken.(string)
	return &id, nil
}

// Represents a model entry, including its modelId, dependencies, and status
type modelEntry struct {
	model        jsonObject       // The object of the model
	modelId      string           // ID of the model
	dependencies []*modelEntry    // References to other modelEntry instances which the current instance is dependent on
	status       processingStatus // The processingStatus for the entry
}

// Creates a new modelEntry instance based on a jsonObject
func newModelEntry(object jsonObject) (*modelEntry, error) {
	entry := new(modelEntry)
	entry.model = object

	modelId, err := object.getModelId()
	if err != nil {
		return nil, fmt.Errorf("unable to create model entry: %s", err)
	}

	entry.modelId = *modelId
	entry.dependencies = make([]*modelEntry, 0)
	entry.status = none

	return entry, nil
}

// Gets the list of model IDs which the current modelEntry is dependent on
func (entry *modelEntry) getModelDependencies() []string {
	dependencies := make([]string, 0)
	contents, ok := entry.model["contents"]
	if ok {
		componentsArr, ok := contents.([]interface{})
		if ok {
			for _, item := range componentsArr {
				itemMap := item.(map[string]interface{})
				itemType, ok := itemMap["@type"].(string)
				if ok && itemType == "Component" {
					schema, ok := itemMap["schema"].(string)
					if ok {
						dependencies = append(dependencies, schema)
					}
				}
			}
		}
	}

	extends, ok := entry.model["extends"]
	if ok {
		switch extends := extends.(type) {
		case []interface{}:
			items := make([]string, len(extends))
			for i := range extends {
				items[i] = extends[i].(string)
			}
			dependencies = append(dependencies, items...)
		case interface{}:
			dependencies = append(dependencies, extends.(string))
		}
	}

	check := make(map[string]bool)
	var distinct []string

	for _, str := range dependencies {
		if _, ok := check[str]; !ok {
			check[str] = true
			distinct = append(distinct, str)
		}
	}

	return distinct
}
