# Output templating

## Go template syntax

See Go's [text/template package documentation](https://pkg.go.dev/text/template)

## Report structure reference for usage in templates

```go
{
    Benchmark {
        Length  int
        Success int
        Fail    int
        Duration time.Duration
        Records []{
            Time   time.Duration
            Code   int          
            Bytes  int          
            Error  string       
            Events []{
                Name string
                Time time.Duration
            }
        }
    }

    Metadata {
        Config {
            Request {
                Method string
                URL    *url.URL
                Header http.Header
                Body   Body
            }
            Runner {
                Requests       int
                Concurrency    int
                Interval       time.Duration
                RequestTimeout time.Duration
                GlobalTimeout  time.Duration
            }
            Output {
                Out      []string
                Silent   bool
                Template string
            }
        }

        FinishedAt time.Time
    }
}
```

### Additionnal template functions

- `stats`:
    - `{{ stats.Min }}`: Minimum recorded request time
    - `{{ stats.Max }}`: Maximum recorded request time
    - `{{ stats.Mean }}`: Mean request time

- `fail`:
    - `{{ fail }}`: Fails the test and exit 1 (better used in a condition!)
    - `{{ fail "Too long!" }}`: Same with error message

## Some examples

- Custom summary
    ```yml
    template: |
      {{ .Benchmark.Length }}/{{ .Metadata.Config.Runner.Requests }} requests
      {{ .Benchmark.Fail }} errors
      ✔︎ Done in {{ .Benchmark.Duration.Milliseconds }}ms.
    ```

    ```txt
    100/100 requests
    0 errors
    ✔︎ Done in 2034ms.
    ```

- Display only the average request time
    ```yml
    template: '{{ stats.Mean }}'
    ```

    ```txt
    237ms
    ```

- Fail the test if any request exceeds 200ms
    ```yml
    template: |
      {{- if ge stats.Max.Milliseconds 200 -}}
          {{ fail "TOO SLOW" }}
      {{- else -}}
          OK
      {{- end -}}
    ```

    if max >= 200ms:
    ```txt
    test failed: TOO SLOW
    exit status 1
    ```

    else:
    ```txt
    OK
    ```
