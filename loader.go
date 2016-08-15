package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

type SuiteSource interface {
	Next() *TestSuite
}

type DirSuiteSource struct {
	RootDir string

	files []FileSuiteSource
	pos   int
}

func (ds *DirSuiteSource) init() {
	filepath.Walk(ds.RootDir, ds.addFileSource)
}

func (ds *DirSuiteSource) addFileSource(path string, info os.FileInfo, err error) error {
	if err != nil {
		return nil
	}

	if info.IsDir() {
		return nil
	}

	ds.files = append(ds.files, FileSuiteSource{Path: path})
	return nil
}

func (ds *DirSuiteSource) Next() *TestSuite {
	if len(ds.files) <= ds.pos {
		return nil
	}

	file := ds.files[ds.pos]
	ds.pos = ds.pos + 1
	su := file.Next()

	if su == nil {
		return nil
	}

	dir, _ := filepath.Rel(ds.RootDir, filepath.Dir(file.Path))
	su.Dir = dir

	return su
}

type FileSuiteSource struct {
	Path string
}

func (fs *FileSuiteSource) Next() *TestSuite {
	if fs.Path == "" {
		return nil
	}

	path := fs.Path
	fs.Path = ""
	info, _ := os.Lstat(path)

	if info.IsDir() {
		fmt.Println("DIR")
		return nil
	}

	if !strings.HasSuffix(info.Name(), ".json") {
		return nil
	}

	ok := isSuite(path)
	if !ok {
		return nil
	}

	err := validateSuite(path)
	if err != nil {
		fmt.Printf("Invalid suite file: %s\n%s\n", path, err.Error())
		return nil
	}

	content, e := ioutil.ReadFile(path)

	if e != nil {
		fmt.Println("Cannot read file: " + e.Error())
		return nil
	}

	var testCases []TestCase
	err = json.Unmarshal(content, &testCases)
	if err != nil {
		fmt.Println("Cannot parse file: " + err.Error())
		return nil
	}

	dir, _ := filepath.Rel(filepath.Dir(path), filepath.Dir(path))
	su := TestSuite{
		Name:  strings.TrimSuffix(info.Name(), filepath.Ext(info.Name())),
		Dir:   dir,
		Cases: testCases,
	}

	return &su
}

func load(source SuiteSource, channel chan<- TestSuite) {
	content := source.Next()

	for content != nil {
		channel <- *content
		content = source.Next()
	}

	close(channel)
}

func NewDirLoader(rootDir string) <-chan TestSuite {
	channel := make(chan TestSuite)

	source := &DirSuiteSource{RootDir: rootDir}
	source.init()

	go load(source, channel)

	return channel
}

func NewFileLoader(path string) <-chan TestSuite {
	channel := make(chan TestSuite)

	go load(&FileSuiteSource{Path: path}, channel)

	return channel
}

func isSuite(path string) bool {
	schemaLoader := gojsonschema.NewStringLoader(suiteShapeSchema)

	path, _ = filepath.Abs(path)
	documentLoader := gojsonschema.NewReferenceLoader("file:///" + path)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return false
	}

	return result.Valid()
}

func validateSuite(path string) error {
	schemaLoader := gojsonschema.NewStringLoader(suiteDetailedSchema)

	path, _ = filepath.Abs(path)
	documentLoader := gojsonschema.NewReferenceLoader("file:///" + path)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		var msg string
		for _, desc := range result.Errors() {
			msg = fmt.Sprintf(msg+"%s\n", desc)
		}
		return errors.New(msg)
	}

	return nil
}

// used to detect suite
const suiteShapeSchema = `
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "type": "array",
  "items": {
    "type": "object",
    "properties": {
      "name": {
        "type": "string"
      },
      "calls": {
        "type": "array"
      }
    },
    "required": [
      "name",
      "calls"
    ]
  }
}
`

// used to validate suite
const suiteDetailedSchema = `
{
	"$schema": "http://json-schema.org/draft-04/schema#",
	"type": "array",
	"items": {
		"type": "object",
		"properties": {
			"name": {
				"type": "string"
			},
			"calls": {
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"on": {
							"type": "object",
							"properties": {
								"method": {
									"type": "string"
								},
								"url": {
									"type": "string"
								},
								"headers": {
									"type": "object"
								},
								"params": {
									"type": "object"
								}
							},
							"required": [
								"method",
								"url"
							]
						},
						"expect": {
							"type": "object",
							"properties": {
								"statusCode": {
									"type": "integer"
								},
								"contentType": {
									"type": "string"
								},
								"headers": {
									"type": "object"
								},
								"body": {
									"type": "object"
								},
								"bodySchemaFile": {
									"type": "string"
								},
								"bodySchemaURI": {
									"type": "string"
								}
							},
							"additionalProperties": false
						}
					},
					"required": [
						"on",
						"expect"
					]
				}
			}
		},
		"required": [
			"calls"
		]
	}
}
`
