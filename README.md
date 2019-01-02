# Jmock

[![Build Status](https://travis-ci.com/fullpipe/jmock.svg?branch=master)](https://travis-ci.com/fullpipe/jmock)
[![Docker Image](https://img.shields.io/microbadger/image-size/fullpipe%2Fjmock.svg)](https://cloud.docker.com/repository/docker/fullpipe/jmock)

Simple and easy to use json/post API mock server

## Install

Install binary

```
go get -u github.com/fullpipe/jmock
```

or use with docker

```
docker run -v ${PWD}/mocks:/mocks fullpipe/jmock
```

## Usage

First create mocks collection file some where in your project. Use [standard
wildcards](http://tldp.org/LDP/GNU-Linux-Tools-Summary/html/x11655.htm) for
request matching. For example `./mocks/users.json`:

```json
[
  {
    "request": {
      "method": "POST",
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

```bash
jmock ./mocks/*.json --port 9090
```

Thats it your fake api is ready. Check the request

```bash
curl localhost:9091/api/users/1
```

Output
```json
{
  "name": "John Doe"
}
```

## Mocks

Mock consists of 3 blocks `request`, `response`, `proxy`

### Request

You could match request by:

```
    "request": {
      "method": "POST", // http method
      "url": "/api/users/*", // query path
      "post": { // post variables
          "first_name": "Jo*",
          "last_name": "?oe"
      },
      "json": { // JSON request body
        "name": "*",
        "gender": "?"
      }
    }
```

### Response

Response for matched request with required response:
```
    "response": {
      "code": 200, // status code
      "body": "plain text or html", // response body
      "json": { // response body with json
        "name": "John Doe"
      }
    }
 ```

### Proxy (optional)

If you get one mock working. You could simply delete it. Or use `proxy` to
redirect matched request to real API.

```
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

## Todo

- request match level
- headers matching
- add more test cases
