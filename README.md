# runner

## Usage

```txt
Usage of runner:
  -c int
        Number of connections to run concurrently (default 1)
  -d duration
        Duration of test, in seconds (default 1m0s)
  -r int
        Number of requests to run, use duration as exit condition if omitted
  -t duration
        Timeout for each http request, in seconds (default 10s)
```

## Example and ouput

Run the test for 10 seconds with 100 concurrent goroutines.

```sh
go run cmd/runner/main.go -c 100 -d 10s  http://echo.jsontest.com/title/ipsum/content/blah

Testing url: http://echo.jsontest.com/title/ipsum/content/blah
concurrency: 100
duration: 10s

2368
```

Run the test for 10 seconds or until 1000 requests have been made with 100 concurrent goroutines.

```sh
go run cmd/runner/main.go -c 100 -d 10s -r 1000 http://echo.jsontest.com/title/ipsum/content/blah

Testing url: http://echo.jsontest.com/title/ipsum/content/blah
concurrency: 100
requests: 1000
duration: 10s

1000
```
