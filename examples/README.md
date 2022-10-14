## Base URL Context

```sh
boze -H https://postman-echo.com:443 base-url-context.json
```

```json
"context": {
    "hostname": "{ctx:base_url_hostname}",
    "host": "{ctx:base_url_host}",
    "port": "{ctx:base_url_port}",
    "schema": "{ctx:base_url_schema}"
}
```

```json
"context": {
    "hostname": "postman-echo.com",
    "host": "postman-echo.com:443",
    "port": "443",
    "schema": "https"
}
```