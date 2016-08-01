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
- github.com/xeipuuv/gojsonschema
- github.com/fatih/color
- github.com/mattn/go-colorable
- github.com/mattn/go-isatty
- github.com/clbanning/mxj
- github.com/fatih/structs
- github.com/lestrrat/go-libxml2
- github.com/pkg/errors

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
