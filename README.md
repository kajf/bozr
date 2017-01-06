![Bozr](https://raw.githubusercontent.com/kajf/bozr/master/assets/bozr.png)

Minimalistic tool to perform REST API tests based on JSON description

[![Build Status](https://travis-ci.org/kajf/bozr.svg?branch=master)](https://travis-ci.org/kajf/bozr?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/kajf/bozr)](https://goreportcard.com/report/github.com/kajf/bozr)
## Usage

```bash
bozr [OPTIONS] (DIR|FILE)

Options:
  -H, --host		Server to test
  -h, --help		Print usage
  -i, --info		Enable info mode. Print request and response details.
  -d, --debug		Enable debug mode
      --junit		Enable junit xml reporter
  -v, --version		Print version information and quit

Examples:
  bozr ./examples/suite-file.json
  bozr -H http://example.com ./examples
```

Usage [demo](https://asciinema.org/a/85699)

## Installation
Download the [latest binary release](https://github.com/kajf/bozr/releases) and unpack it.

## Test Suite Format
Test suite (suite_name.json)

    ├ Test A [test case]
    |   ├ Name
    │   ├ Call one
    |   |   ├ args [value(s) for placeholders to use in request params, headers or body]
    │   │   ├ on [single http request]
    │   │   ├ expect [http response asserts: code, headers, body, schema, etc.]
    │   │   └ remember [optionally remember variable(s) for the next call to use in request params, headers or body]
    │   └ Call two
    |       ├ args
    │       ├ on
    │       ├ expect
    │       └ remember
    └ Test B
        ├ Name
        └ Call one
            ├ args
            ├ on
            ├ expect
            └ remember

### Section 'On'
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
method | HTTP method
url | HTTP request URL
headers | HTTP request headers
params | HTTP query params
bodyFile | File to send as a request payload (path relative to test suite json)
body | String or JSON object to send as a request payload

### Section 'Expect'
Represents assertions for http response of the call

```json
"expect": {
    "statusCode": 200,
    "contentType": "application/json",
    "body": {
        "errors.size()": 0
    }
}
```

Assertion | Description | Example
------------ | ------------- | --------------
statusCode | Expected http response header 'Status Code' | 200
contentType | Expected http response 'Content-Type' | application/json
bodySchemaFile | Path to json schema to validate response body against (path relative to test suite file) | login-schema.json
bodySchemaURI | URI to json schema to validate response body against (absolute or relative to the host) | http://example.com/api/scheme/login-schema.json
body | Body matchers: equals, search, size |
absent | Paths that are NOT expected to in response | ['user.cardNumber', 'user.password']
headers | Expected http headers, specified as a key-value pairs. |

### 'Expect' body matchers

```json
"expect": {
    "body": {
        "users.1.surname" : "Doe",
        "~users.name":"Joe",
        "errors.size()": 0
    }
}
```

Type | Assertion | Example
------ | ------------- | --------------
equals | Root 'users' array zero element has value of 'id' equal to '123'  | "users.0.id" : "123"
search | Root 'users' array contains element(s) with 'name' equal to 'Jack' or 'Dan' and 'Ron'  | "~users.name" : "Jack" or "~users.name" : ["Dan","Ron"]
size | Root 'company' element has 'users' array with '22' elements within 'buildings' array | "company.buildings.users.size()" : 22

XML:
- To match attribute use `-` symbol before attribute name. E.g. `users.0.-id`
- Namespaces are ignored
- Only string matcher values are supported (since xml has no real data types, so everything is a string)

### 'Expect absent' body matchers
Represents paths not expected to be in response body.
Mostly used for security checks (e.g. returned user object should not contain password or credit card number fields)
Path fromat is the same as in 'Expect' body section
```json
"expect": {
    "absent": ['user.cardNumber', 'user.password']
}
```

### Section 'Args'
Specifies plaseholder values for future reference (within test scope)

```json
"args": {
  "currencyCode" : "USD",
  "magicNumber" : "12f"
}
```
Given 'args' are defined like above, placeholders {currencyCode} and {magicNumber} may be used in params, body or bodyFile.

example_bodyfile.json

```json
{
  "bankAccount" : {
    "currency": "{currencyCode}",
    "amount" : 1000,
    "secret" : "{magicNumber}"
  }
}
```

Resulting data will contain "USD" and "12f" values instead of placeholders.

```json
{
  "bankAccount" : {
    "currency": "USD",
    "amount" : 1000,
    "secret" : "12f"
  }
}
```

You can use placeholders inside of the `url` and `headers` fields.

```json
{
    "on": {
        "method": "GET",
        "url": "{hateoas_reference}",
        "headers": {
            "X-Secret-Key": "{secret_key}"
        }
    }
}
```

### Section 'Remember'
Similar to 'Args' section, specifies plaseholder values for future reference (within test scope)

The difference is that values for placeholders are taken from response (syntax is similar to 'Expect' 'equal' matchers)

```json
"remember": {
  "currencyCode" : "currencies.0.code",
  "createdId" : "result.newId"
}
```

This section allowes more complex test scenarios like

'request login token, remember, then use remembered {token} to request some data and verify'

'create resource, remember resource id from response, then use remembered {id} to delete resource'

## Editor integration
To make work with test files convenient, we suggest to configure you text editors to use [this](./assets/test.schema.json) json schema. In this case editor will suggest what fields are available and highlight misspells.

Editor | JSON Autocomplete
-------|------------------
JetBrains Tools| [native support](https://www.jetbrains.com/help/webstorm/2016.1/json-schema.html?page=1)
Visual Studio Code| [native support](https://code.visualstudio.com/docs/languages/json#_json-schemas-settings)
Vim| [plugin](https://github.com/Quramy/vison)

## Dependency management
To build project you need a dependency management tool - https://glide.sh/
After you installed it, you can run the following command to download all dependencies:

```bash
cd bozr
glide install
```

Dependencies:
- github.com/xeipuuv/gojsonschema
- github.com/fatih/color
- github.com/mattn/go-colorable
- github.com/mattn/go-isatty
- github.com/clbanning/mxj
- github.com/fatih/structs
