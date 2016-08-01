# t-rest
Minimalistic tool to perform REST API tests based on JSON description

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

## Command-line arguments
- d - path to directory with test-cases-json files
- h - remote host address to run tests against
- v - verbose console output
```bash
t-rest -h http://localhost:8080 -d ./suites -v
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
## Call Section 'Expect'
Section represents assertions for http response of the call

Assertions | Description | Example
------------ | ------------- | --------------
statusCode | expected http response header 'Status Code' | 200
contentType | expected http response 'Content-Type' | application/json
bodySchema | path to json schema to validate respnse body against (path relative to test suite file) | login-schema.json
body | body matchers: equals, like, size() |

### 'Expect' body mathcers
Type | Description | Example
------------ | ------------- | --------------
equals | Exact path within response body. Should contain array indexes and full tree path to element | "employees.0.id" : "123"
like | |
size | |
