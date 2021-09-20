# Jmock

[![Tests Status](https://github.com/fullpipe/jmock/workflows/Tests/badge.svg)](https://github.com/fullpipe/jmock)
[![Docker Image](https://img.shields.io/docker/image-size/fullpipe/jmock/latest)](https://cloud.docker.com/repository/docker/fullpipe/jmock)

Simple and easy to use json/post API mock server

## Install

Install binary

```sh
go install github.com/fullpipe/jmock@latest
```

## Usage

First create mocks collection file some where in your project. Use [standard
wildcards](http://tldp.org/LDP/GNU-Linux-Tools-Summary/html/x11655.htm) for
request matching. For example `./mocks/users.json`:

```json
[
  {
    "request": {
      "method": "OPTIONS",
      "priority": 100
    },
    "response": {
      "code": 204,
      "headers": {
        "Access-Control-Allow-Origin": "*",
        "Access-Control-Allow-Headers": "*"
      }
    }
  },
  {
    "request": {
      "method": "POST",
      "headers": {
        "Authorization": "Bearer *"
      },
      "url": "/api/users",
      "json": {
        "name": "*"
      }
    },
    "response": {
      "code": 200
    }
  },
  {
    "request": {
      "method": "GET",
      "url": "/api/users/*"
    },
    "response": {
      "code": 200,
      "json": {
        "name": "John Doe"
      }
    }
  },
  {
    "request": {
      "method": "GET",
      "url": "/api/posts"
    },
    "proxy": "http://realapi.loc"
  }
]
```

Start jmock server

```sh
jmock "./mocks/*.json" --port 9090 --watch
```

Thats it your fake api is ready. Check the request

```sh
curl localhost:9090/api/users/1
```

Output
```json
{
  "name": "John Doe"
}
```

### Usage with docker

Run mock server

```sh
docker run -p 9090:9090 -v ${PWD}/mocks:/mocks fullpipe/jmock
```

Or if you need to watch files

```sh
docker run -p 9090:9090 -v ${PWD}/mocks:/mocks fullpipe/jmock /mocks/*.json --port 9090 --watch
```

Or with docker-compose

```yaml
services:
    api:
        image: fullpipe/jmock
        command: "/mocks/*.json --port 9090 --watch"
        ports:
            - "9090:9090"
        volumes:
            - ./mocks:/mocks
```

## Mocks

Mock consists of 3 blocks `request`, `response`, `proxy`

### Request

You could match request by:

```jsonc
    "request": {
      "method": "POST", // http method
      "url": "/api/users/*", // query path
      "headers": {
        "Authorization": "Bearer *"
      },
      "query" { // get params
          "country": "R*"
      },
      "post": { // post variables
          "first_name": "Jo*",
          "last_name": "?oe"
      },
      "json": { // JSON request body
        "name": "*",
        "gender": "?"
      },
      "priority": 42 // high number for more "sticky" requests
    }
```

### Response

For matched request server returns response:

```jsonc
    "response": {
      "code": 200, // status code
      "body": "plain text or html", // response body
      "json": { // response body with json
        "name": "John Doe"
      },
      "headers": { // add response headers if required
        "Access-Control-Allow-Origin": "*"
      }
    }
 ```

### Proxy (optional)

If you get one mock working. You could use `proxy` to
bypass matched request to real API.

```json
    "proxy": "http://realapihost.loc"
```

## Examples

### JSON RPC 2.0
```json
[
  {
    "request": {
      "mehtod": "POST",
      "url": "/rpc",
      "json": {
        "jsonrpc": "2.0",
        "method": "registerUser",
        "id": 1,
        "params": {
          "email": "*"
        }
      }
    },
    "response": {
      "code": 200,
      "json": {
        "jsonrpc": "2.0",
        "result": {
          "user_id": "15"
        },
        "id": 1
      }
    }
  }
]
```
