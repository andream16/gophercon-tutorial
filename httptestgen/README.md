# httptestgen

Example code gen tool that reads that specs like:

```json
[
  {
    "func": "CreateUserHandler",
    "test-cases": [
      {
        "case_descr": "it should succeed when a valid user is passed",
        "request": {
          "method": "POST",
          "path": "/users",
          "body": {
            "name": "Andrea",
            "email": "andrea@gitpod.io"
          },
          "headers": {
            "Content-Type": "application/json",
            "Authorization": "Bearer token123"
          }
        },
        "response": {
          "status_code": "201",
          "body": {
            "user": {
              "id": 1,
              "name": "Andrea",
              "email": "andrea@gitpod.io"
            },
            "message": "User created successfully"
          },
          "headers": {
            "Content-Type": "application/json"
          }
        }
      }
    ]
  }
]
```

And generates tests for specified http handlers (`CreateUserHandler`).

# Example usage

## CLI

```shell
go run ./cmd \
  -input=examples/handler/handler.go \
  -output=examples/handler/handler_test.go \
  -testcases=examples/handler/testdata/testcases.json \
  -request-type=CreateUserRequest
```

## Go Generate

Add this to your target file.
```go
//go:generate go run ../path/to/cmd -input=handler.go -output=handler_test.go -testcases=testdata/testcases.json -request-type=CreateUserRequest
```

Then run:

```sh
go generate
```
