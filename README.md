# Go Project

A zero dependency project starter template with HTTP server for Go.

## Features

- simple project structure
- zero dependency
- HTTP server
  - `net/http` compatible
  - middleware support
  - centralized error handling
  - named route parameters

## Requirements

- Go 1.22+ (for `net/http`)

## Structure

- `/app` - app specific code
  - `/app/config` - app configuration
  - `/app/server` - HTTP server
    - `/app/server/handler` - HTTP handlers
    - `/app/server/middleware` - HTTP middleware
- `/cmd` - entry points
  - `/cmd/app` - app entry point

## Makefile

Display Makefile help:

```
$ make
audit                Run QC checks (tests, go mod verify, go vet, govulncheck, gosec)
build                Build the project
docker-build         Build the docker image
help                 Display help
run                  Run the project
test                 Run tests (use `test v=1` to see verbose output)
test-bench           Run tests with benchmarks
test-cover           Run tests and display coverage
tidy                 Run go mod tidy
update               Update dependencies
```
