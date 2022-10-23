<h1 align="center">benchttp/engine</h1>

<p align="center">
  <a href="https://github.com/benchttp/engine/actions/workflows/ci.yml?query=branch%3Amain">
    <img alt="Github Worklow Status" src="https://img.shields.io/github/workflow/status/benchttp/engine/Lint%20&%20Test%20&%20Build"></a>
  <a href="https://codecov.io/gh/benchttp/engine">
    <img alt="Code coverage" src="https://img.shields.io/codecov/c/gh/benchttp/engine?label=coverage"></a>
  <a href="https://goreportcard.com/report/github.com/benchttp/engine">
    <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/benchttp/engine" /></a>
  <br />
  <a href="https://pkg.go.dev/github.com/benchttp/engine#section-documentation">
    <img alt="Go package Reference" src="https://img.shields.io/badge/pkg-reference-informational?logo=go" /></a>
  <a href="https://github.com/benchttp/engine/releases">
    <img alt="Latest version" src="https://img.shields.io/github/v/tag/benchttp/engine?label=release"></a>
</p>

## About

Benchttp engine is a Go library providing a way to perform benchmarks and tests
on HTTP endpoints.

## Installation

### Prerequisites

Go1.17 environment or higher is required.

Install.

```txt
go get github.com/benchttp/engine
```

## Usage

### Basic usage

```go
package main

import (
    "context"
    "fmt"

    "github.com/benchttp/engine/runner"
)

func main(t *testing.T) {
    // Set the request to send
    req, _ := http.NewRequest("GET", "http://localhost:3000", nil)

    // Configure the runner
    rnr := runner.DefaultRunner()
    rnr.Request = req

    // Run benchmark, get report
    report, _ := rnr.Run(context.Background())

    fmt.Println(report.Metrics.ResponseTimes.Mean)
}
```

### Usage with JSON config via `configparse`

```go
package main

import (
    "context"
    "fmt"

    "github.com/benchttp/engine/configparse"
)

func main() {
    // JSON configuration obtained via e.g. a file or HTTP call
    jsonConfig := []byte(`
{
  "request": {
    "url": "http://localhost:9999"
  }
}`)

    runner, _ := configparse.JSON(jsonConfig)
    report, _ := runner.Run(context.Background())

    fmt.Println(report.Metrics.ResponseTimes.Mean)
}
```

📄 Please refer to [our Wiki](https://github.com/benchttp/engine/wiki/IO-Structures) for exhaustive `Config` and `Report` structures (and more!)

## Development

### Prerequisites

1. Go 1.17 or higher is required
1. Golangci-lint for linting files

### Main commands

| Command      | Description                         |
| ------------ | ----------------------------------- |
| `make lint`  | Runs lint on the codebase           |
| `make tests` | Runs tests suites from all packages |
| `make check` | Runs both lint and tests            |
