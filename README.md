# t-rest
Minimalistic tool to perform REST API tests based on JSON description

[![Build Status](https://travis-ci.org/kajf/t-rest.svg?branch=master)](https://travis-ci.org/kajf/t-rest?branch=master)

## Command-line arguments
- d - path to directory with test-cases-json files
- h - remote host address to run tests against
- v - verbose console output
```bash
t-rest -h http://localhost:8080 -d ./examples -v
```
## Test Suite Format
Test suite (suite_name.json)

    ├ Test A [single test]
    │   ├ Call one
    │   │   ├ on [single http request]
    │   │   ├ expect [http response asserts: code, headers, body, schema, etc.]
    │   │   └ remember [optionally remember variable(s) for the next call to use in request params, headers or body]
    │   └ Call two
    │       ├ on
    │       ├ expect
    │       └ remember
    └ Test B
        └ Call one
            ├ on
            ├ expect
            └ remember
## Section 'On'
Represents http request parameters
```json
    "on": {
        "method": "POST",
        "url": "/api/company/users",
        "headers": {
            "Accept": "application/json",
            "Content-Type": "application/json"
        },
        "params": {
            "role": "admin"
        },
        "bodyFile" : "admins.json"
    }
```
Field | Description
------------ | -------------
method | http method
url | http request URL
headers | http request headers
params | http query params
bodyFile | file to send as a request payload (path relative to test suite json)
body | string to send as a request payload

## Section 'Expect'
Represents assertions for http response of the call
JSON example
```json
    "expect": {
        "statusCode": 200,
        "contentType": "application/json",
        "body": {
            "errors.size()": "0"
        }
    }
```
Assertion | Description | Example
------------ | ------------- | --------------
statusCode | expected http response header 'Status Code' | 200
contentType | expected http response 'Content-Type' | application/json
bodySchema | path to json schema to validate response body against (path relative to test suite file) | login-schema.json
body | body matchers: equals, search, size |

### 'Expect' body matchers
```json
    "expect": {
        "body": {
            "users.1.surname" : "Doe",
            "~users.name":"Joe",
            "errors.size()": "0"
        }
    }
```
Type | Assertion | Example
------------ | ------------- | --------------
equals | Root 'users' array zero element has value of 'id' equal to '123'  | "users.0.id" : "123"
search | Root 'users' array contains element with 'name' equal to 'Jack'  | "users.name" : "Jack"
size | Root 'company' element has 'users' array with '22' elements within 'buildings' array | "company.buildings.users.size()" : "22"

XML:
- To match attribute use `-` symbol before attribute name. E.g. `users.0.-id`
- Namespaces are ignored

## Dependency management
To build project you need a dependency management tool - https://glide.sh/
After you installed it, you can run the following command to download all dependencies:

```bash
cd t-rest
glide install
```

Dependencies:
- github.com/xeipuuv/gojsonschema
- github.com/fatih/color
- github.com/mattn/go-colorable
- github.com/mattn/go-isatty
- github.com/clbanning/mxj
- github.com/fatih/structs
