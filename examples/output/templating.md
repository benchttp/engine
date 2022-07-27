# Output templating

## Go template syntax

See Go's [text/template package documentation](https://pkg.go.dev/text/template)

## Report structure reference for usage in templates

See [IO Structures](https://github.com/benchttp/engine/wiki/IO-Structures#go-1) in our wiki.

### Additionnal template functions

- `fail`:

  - `{{ fail }}`: Fails the test and exit 1 (better used in a condition!)
  - `{{ fail "Too long!" }}`: Same with error message

## Some examples

- Custom summary

  ```yml
  template: |
    {{ .Metrics.TotalCount }}/{{ .Metadata.Config.Runner.Requests }} requests
    {{ .Metrics.FailureCount }} errors
    ✔︎ Done in {{ .Metadata.TotalDuration.Milliseconds }}ms.
  ```

  ```txt
  100/100 requests
  0 errors
  ✔︎ Done in 2034ms.
  ```

- Display only the average request time

  ```yml
  template: "{{ .Metrics.Avg }}"
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
