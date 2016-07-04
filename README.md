# t-rest
Minimalistic tool to autogenerate REST API tests based on JSON description

## Dependency management
To build project you need a dependency management tool - https://glide.sh/
After you installed it, you can run the following command to download all dependencies:

```bash
cd t-rest
glide install
```

Dependencies:
- https://github.com/xeipuuv/gojsonschema

## Command-line params
- d - path to directory with test-cases-json files 
- h - remote host address to run tests against
```bash
t-rest -h http://localhost:8080 -d ./suites
```
