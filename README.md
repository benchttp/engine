# runner

<p align="center">
  <a href="https://github.com/benchttp/runner/actions/workflows/ci.yml?query=branch%3Amain">
    <img alt="Github Worklow Status" src="https://img.shields.io/github/workflow/status/benchttp/runner/Lint%20&%20Test%20&%20Build"></a>
  <a href="https://codecov.io/gh/benchttp/runner">
    <img alt="Code coverage" src="https://img.shields.io/codecov/c/gh/benchttp/runner?label=coverage"></a>
  <a href="https://goreportcard.com/report/github.com/benchttp/runner">
    <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/benchttp/runner" /></a>
  <br />
  <a href="https://pkg.go.dev/github.com/benchttp/runner#section-documentation">
    <img alt="Go package Reference" src="https://img.shields.io/badge/pkg-reference-informational?logo=go" /></a>
  <a href="https://github.com/benchttp/runner/releases">
    <img alt="Latest version" src="https://img.shields.io/github/v/tag/benchttp/runner?label=release"></a>
</p>

## Usage

```sh
# Build executable
make build

# Run benchttp
./bin/benchttp [options]
```

### Configuration file

By default, the runner will read the config files located in the working directory,
in that order of priority: `.benchttp.yml`, `.benchttp.yaml`, `.benchttp.json`.

For configuration examples, see [./test/testdata/config](./test/testdata/config).

### CLI

Any configuration option can be overridden via the CLI:

```txt
CLI options:
  -configFile string
        Path to config file (default ".benchttp.(yml|yaml|json)")
  -concurrency int
        Number of connections to run concurrently (default 1)
  -globalTimeout duration
        Global timeout for the test (default 30s)
  -requests int
        Maximum number of requests to run, 0 means no limit (default 0)
  -timeout duration
        Timeout for each HTTP request (default 10s)
```

## Example and ouput

Run the test for 10 seconds with 100 concurrent goroutines.

```sh
./bin/benchttp -concurrency 100 -globalTimeout 10s -url http://echo.jsontest.com/title/ipsum/content/blah

Testing url: http://echo.jsontest.com/title/ipsum/content/blah
concurrency: 100
duration: 10s

2368
```

Run the test for 10 seconds or until 1000 requests have been made with 100 concurrent goroutines.

```sh
./bin/benchttp -concurrency 100 -globalTimeout 10s -requests 1000 -url http://echo.jsontest.com/title/ipsum/content/blah

Testing url: http://echo.jsontest.com/title/ipsum/content/blah
concurrency: 100
requests: 1000
duration: 10s

1000
```
