package main

import (
	"github.com/xeipuuv/gojsonschema"
	"strings"
	"testing"
)

func Test_validateSuiteShape(t *testing.T) {

	tests := []struct {
		name    string
		args    gojsonschema.JSONLoader
		wantErr string
	}{
		{
			name: "object in args not allowed",
			args: gojsonschema.NewStringLoader(`[{
				"calls": [{
					"args": {
						"str":"abc",
						"obj": {}
					},
                  	"on": {"method": "GET","url":"smth"},
                  	"expect": {"statusCode":200}
				}]
			}]`),
			wantErr: "Invalid type",
		},
		{
			name: "basic type in args allowed",
			args: gojsonschema.NewStringLoader(`[{
				"calls": [{
					"args": {
						"str":"abc",
						"num": 1,
						"f": 0.12,
						"b": true,
						"n": null
					},
                  	"on": {"method": "GET","url":"smth"},
                  	"expect": {"statusCode":200}
				}]
			}]`),
			wantErr: "",
		},
		{
			name: "object in on.headers not allowed",
			args: gojsonschema.NewStringLoader(`[{
				"calls": [{
                  	"on": {
						"method": "GET",
						"url":"smth",
						"headers": {
							"str":"abc",
							"obj": {}
						}
					},
                  	"expect": {"statusCode":200}
				}]
			}]`),
			wantErr: "Invalid type",
		},
		{
			name: "number in on.headers not allowed",
			args: gojsonschema.NewStringLoader(`[{
				"calls": [{
                  	"on": {
						"method": "GET",
						"url":"smth",
						"headers": {
							"str":"abc",
							"num": 1
						}
					},
                  	"expect": {"statusCode":200}
				}]
			}]`),
			wantErr: "Invalid type",
		},
		{
			name: "string in on.headers allowed",
			args: gojsonschema.NewStringLoader(`[{
				"calls": [{
                  	"on": {
						"method": "GET",
						"url":"smth",
						"headers": {
							"str":"abc"
						}
					},
                  	"expect": {"statusCode":200}
				}]
			}]`),
			wantErr: "",
		},
		{
			name: "object in expect.headers not allowed",
			args: gojsonschema.NewStringLoader(`[{
				"calls": [{
                  	"on": {
						"method": "GET",
						"url":"smth"
					},
                  	"expect": {
						"statusCode":200,
						"headers": {
							"str":"abc",
							"obj": {}
						}
					}
				}]
			}]`),
			wantErr: "Invalid type",
		},
		{
			name: "number in expect.headers not allowed",
			args: gojsonschema.NewStringLoader(`[{
				"calls": [{
                  	"on": {
						"method": "GET",
						"url":"smth"
					},
                  	"expect": {
						"statusCode":200,
						"headers": {
							"str":"abc",
							"num": 1
						}
					}
				}]
			}]`),
			wantErr: "Invalid type",
		},
		{
			name: "string in expect.headers allowed",
			args: gojsonschema.NewStringLoader(`[{
				"calls": [{
                  	"on": {
						"method": "GET",
						"url":"smth"
					},
                  	"expect": {
						"statusCode":200,
						"headers": {
							"str":"abc"
						}
					}
				}]
			}]`),
			wantErr: "",
		},
		{
			name: "object in remember.headers not allowed",
			args: gojsonschema.NewStringLoader(`[{
				"calls": [{
                  	"on": {
						"method": "GET",
						"url":"smth"
					},
                  	"expect": {
						"statusCode":200
					},
					"remember": {
						"headers": {
							"str":"abc",
							"obj": {}
						}
					}
				}]
			}]`),
			wantErr: "Invalid type",
		},
		{
			name: "number in remember.headers not allowed",
			args: gojsonschema.NewStringLoader(`[{
				"calls": [{
                  	"on": {
						"method": "GET",
						"url":"smth"
					},
                  	"expect": {
						"statusCode":200
					},
					"remember": {
						"headers": {
							"str":"abc",
							"num": 1
						}
					}
				}]
			}]`),
			wantErr: "Invalid type",
		},
		{
			name: "string in remember.headers allowed",
			args: gojsonschema.NewStringLoader(`[{
				"calls": [{
                  	"on": {
						"method": "GET",
						"url":"smth"
					},
                  	"expect": {
						"statusCode":200
					},
					"remember": {
						"headers": {
							"str":"abc"
						}
					}
				}]
			}]`),
			wantErr: "",
		},
		{
			name: "objects in absent not allowed",
			args: gojsonschema.NewStringLoader(`[{
				"calls": [{
                  	"on": {"method": "GET","url":"smth"},
                  	"expect": {
						"statusCode":200, "absent": [{}]
					}
				}]
			}]`),
			wantErr: "Invalid type",
		},
		{
			name: "numbers in absent not allowed",
			args: gojsonschema.NewStringLoader(`[{
				"calls": [{
					"on": {"method": "GET","url":"smth"},
                  	"expect": {
						"statusCode":200,
						"absent": [1,2]
					}
				}]
			}]`),
			wantErr: "Invalid type",
		},
		{
			name: "null in absent not allowed",
			args: gojsonschema.NewStringLoader(`[{
				"calls": [{
                  	"on": {"method": "GET","url":"smth"},
                  	"expect": {
						"statusCode":200,
						"absent": [null]
					}
				}]
			}]`),
			wantErr: "Invalid type",
		},
		{
			name: "string in absent allowed",
			args: gojsonschema.NewStringLoader(`[{
				"calls": [{
                  	"on": {"method": "GET","url":"smth"},
                  	"expect": {"statusCode":200},
					"absent": [
						"root.items.name"
					]
				}]
			}]`),
			wantErr: "",
		},
		{
			name: "object in on.params not allowed",
			args: gojsonschema.NewStringLoader(`[{
				"calls": [{
                  	"on": {
						"method": "GET",
						"url":"smth",
						"params": {
							"str":"abc",
							"obj": {}
						}
					},
                  	"expect": {"statusCode":200}
				}]
			}]`),
			wantErr: "Invalid type",
		},
		{
			name: "string in on.params is allowed",
			args: gojsonschema.NewStringLoader(`[{
				"calls": [{
                  	"on": {
						"method": "GET",
						"url":"smth",
						"params": {
							"str":"abc"
						}
					},
                  	"expect": {"statusCode":200}
				}]
			}]`),
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := validateSuiteDetailed(tt.args)

			if err == nil && tt.wantErr == "" {
				return
			}

			if err != nil && strings.Contains(err.Error(), tt.wantErr) {
				return
			}

			t.Errorf("validateSuiteDetailed() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}
