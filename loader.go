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

// SuiteFile describes location of the suite file.
type SuiteFile struct {
	// Path to the file.
	Path string
	// Base directory from file is loaded.
	BaseDir string
}

// RelDir returns difference between Path and BaseDir.
func (sf SuiteFile) RelDir() string {
	dir, _ := filepath.Rel(sf.BaseDir, filepath.Dir(sf.Path))
	return dir
}

// ToSuite method deserializes suite representation to the object model.
func (sf SuiteFile) ToSuite() *TestSuite {
	if sf.Path == "" {
		return nil
	}

	path := sf.Path
	info, err := os.Lstat(path)

	if err != nil {
		return nil
	}

	if info.IsDir() {
		debug.Print("Ignore dir: " + sf.Path)
		return nil
	}

	if !strings.HasSuffix(info.Name(), ".json") {
		debug.Print("Ignore non json: " + sf.Path)
		return nil
	}

	ok := isSuite(path)
	if !ok {
		return nil
	}

	err = validateSuite(path)
	if err != nil {
		fmt.Printf("Invalid suite file: %s\n%s\n", path, err.Error())
		return nil
	}

	content, e := ioutil.ReadFile(path)

	if e != nil {
		fmt.Println("Cannot read file:", path, "Error: ", e.Error())
		return nil
	}

	var testCases []TestCase
	err = json.Unmarshal(content, &testCases)
	if err != nil {
		fmt.Println("Cannot parse file:", path, "Error: ", err.Error())
		return nil
	}

	su := TestSuite{
		Name:  strings.TrimSuffix(info.Name(), filepath.Ext(info.Name())),
		Dir:   sf.RelDir(),
		Cases: testCases,
	}

	return &su
}

// SuiteFileIterator is an interface to iterate over a set of suite files
// in a given directory
type SuiteFileIterator interface {
	HasNext() bool
	Next() *SuiteFile
}

// DirSuiteFileIterator iterates over all suite files inside of specified root folder.
type DirSuiteFileIterator struct {
	RootDir string

	files []SuiteFile
	pos   int
}

func (ds *DirSuiteFileIterator) init() {
	filepath.Walk(ds.RootDir, ds.addSuiteFile)
	debug.Print("Collected: ", len(ds.files))
}

func (ds *DirSuiteFileIterator) addSuiteFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return nil
	}

	if info.IsDir() {
		return nil
	}

	ds.files = append(ds.files, SuiteFile{
		Path:    path,
		BaseDir: ds.RootDir,
	})
	return nil
}

// HasNext returns true in case there is at least one more suite.
func (ds *DirSuiteFileIterator) HasNext() bool {
	return len(ds.files) > ds.pos
}

// Next returns next deserialized suite.
func (ds *DirSuiteFileIterator) Next() *SuiteFile {
	if len(ds.files) <= ds.pos {
		return nil
	}

	file := ds.files[ds.pos]
	ds.pos = ds.pos + 1
	return &file
}

func load(source SuiteFileIterator, channel chan<- TestSuite) {

	for source.HasNext() {
		sf := source.Next()
		if sf == nil {
			continue
		}
		channel <- *sf.ToSuite()
	}

	close(channel)
}

// NewDirLoader returns channel of suites that are read from specified folder.
func NewDirLoader(rootDir string) <-chan TestSuite {
	channel := make(chan TestSuite)

	source := &DirSuiteFileIterator{RootDir: rootDir}
	source.init()

	go load(source, channel)

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
			"ignore": {
        		"type": "string"
			},
			"calls": {
				"type": "array",
				"items": {
					"type": "object",
					"properties": {
						"args": {
							"type": "object"
						},
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
								},
								"absent": {
								  "type" : "array"
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
