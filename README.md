# Vega health check

It runs the server to check vega node. The program exposes health-check on the given port (8080 by default).

## Example 

### Core

```shell
vega-health-check vega --core-url "http://localhost:3003" --http-port 8080
```

### Data node

```shell
vega-health-check data-node --api-url "http://localhost:3008" --core-url "http://localhost:3003" --http-port 8080
```

### Blockexplorer

```shell
vega-health-check blockexplorer --blockexplorer-api-url "http://localhost:1515" --core-url "http://localhost:3003" --http-port 8080
```

### Example output 

```shell
➜  curl -s localhost:8080 | jq

{
  "status": "UNHEALTHY",
  "reasons": [
    "data node http endpoint is not online",
    "core http endpoint is not online"
  ]
}
```
