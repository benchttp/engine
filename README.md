<h1 align="center">benchttp/engine</h1>

<p align="center">
  <a href="https://github.com/benchttp/sdk/actions/workflows/ci.yml?query=branch%3Amain">
    <img alt="Github Worklow Status" src="https://img.shields.io/github/workflow/status/benchttp/engine/Lint%20&%20Test%20&%20Build"></a>
  <a href="https://codecov.io/gh/benchttp/engine">
    <img alt="Code coverage" src="https://img.shields.io/codecov/c/gh/benchttp/engine?label=coverage"></a>
  <a href="https://goreportcard.com/report/github.com/benchttp/sdk">
    <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/benchttp/sdk" /></a>
  <br />
  <a href="https://pkg.go.dev/github.com/benchttp/sdk#section-documentation">
    <img alt="Go package Reference" src="https://img.shields.io/badge/pkg-reference-informational?logo=go" /></a>
  <a href="https://github.com/benchttp/sdk/releases">
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
go get github.com/benchttp/sdk
```

## Usage

### Basic usage

```go
package main

import (
    "context"
    "fmt"

    "github.com/benchttp/sdk/benchttp"
)

func main() {
    report, _ := benchttp.
        DefaultRunner(). // Default runner with safe configuration
        WithNewRequest("GET", "http://localhost:3000", nil). // Attach request
        Run(context.Background()) // Run benchmark, retrieve report

    fmt.Println(report.Metrics.ResponseTimes.Mean)
}
```

### Usage with JSON config via `configio`

```go
package main

import (
    "context"
    "fmt"

    "github.com/benchttp/sdk/benchttp"
    "github.com/benchttp/sdk/configio"
)

func main() {
    // JSON configuration obtained via e.g. a file or HTTP call
    jsonConfig := []byte(`
{
  "request": {
    "url": "http://localhost:3000"
  }
}`)

    // Instantiate a base Runner (here the default with a safe configuration)
    runner := benchttp.DefaultRunner()

    // Parse the json configuration into the Runner
    _ = configio.UnmarshalJSONRunner(jsonConfig, &runner)

    // Run benchmark, retrieve report
    report, _ := runner.Run(context.Background())

    fmt.Println(report.Metrics.ResponseTimes.Mean)
}
```

📄 Please refer to [our Wiki](https://github.com/benchttp/sdk/wiki/IO-Structures) for exhaustive `Runner` and `Report` structures (and more!)

## Development

### Prerequisites

1. Go 1.17 or higher is required
1. Golangci-lint for linting files

### Main commands

| Command         | Description                                       |
| --------------- | ------------------------------------------------- |
| `./script/lint` | Runs lint on the codebase                         |
| `./script/test` | Runs tests suites from all packages               |
| `./script/doc`  | Serves Go doc for this module at `localhost:9995` |
