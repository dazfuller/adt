package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Defines a byte order mark
var byteOrderMark = []byte{0xEF, 0xBB, 0xBF}

// ModelDirectory defines a location on disk where models may be uploaded from
type ModelDirectory struct {
	Path string
}

// String returns the string representation of the object
func (directory *ModelDirectory) String() string {
	return directory.Path
}

// Set creates a new instance of the ModelDirectory type
func (directory *ModelDirectory) Set(path string) error {
	if len(path) == 0 {
		return fmt.Errorf("a path must be specified")
	}

	folderInfo, err := os.Stat(path)

	if err != nil && os.IsNotExist(err) {
		return fmt.Errorf("the specified path does not exist")
	} else if err != nil {
		return fmt.Errorf("an error occured validating the path: %s", err)
	} else if !folderInfo.IsDir() {
		return fmt.Errorf("the specified path is not a directory")
	}

	*directory = ModelDirectory{Path: path}
	return nil
}

// Gets all models found recursively under the defined path
func (directory *ModelDirectory) getModels() ([]*modelEntry, error) {
	models := make([]*modelEntry, 0)

	err := filepath.Walk(directory.Path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		extension := strings.ToLower(filepath.Ext(info.Name()))

		// Only inspect files which are JSON or DTDL files
		if !info.IsDir() && (extension == ".json" || extension == ".dtdl") {
			fileContent, err := os.ReadFile(path)
			if err != nil {
				log.Printf("Unable to open file '%s', %s", path, err)
				return nil
			}

			// Strip the byte order mark if it exists
			if bytes.Compare(fileContent[0:3], byteOrderMark) == 0 {
				fileContent = bytes.TrimPrefix(fileContent, byteOrderMark)
			}

			// Read the contents of the file and create a new modelEntry for it
			var jsonContent jsonObject
			err = json.Unmarshal(fileContent, &jsonContent)
			if err != nil {
				log.Printf("Ignoring gile '%s' as it does not contain valid json (%s): %s", path, fileContent[0:2], err)
				return nil
			}

			_, ok := jsonContent["@id"]
			if !ok {
				_, ok = jsonContent["id"]
			}
			if !ok {
				log.Printf("Ignoring file '%s' as it does not contain a valid DTDL document as it is missing the @id/id property", path)
				return nil
			}

			entry, _ := newModelEntry(jsonContent)

			models = append(models, entry)
		}

		return nil
	})

	return models, err
}
