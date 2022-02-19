package output

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/benchttp/runner/ansi"
	"github.com/benchttp/runner/config"
	"github.com/benchttp/runner/output/export"
	"github.com/benchttp/runner/requester"
)

// Output represent a benchmark result as exported by the runner.
type Output struct {
	Report   requester.Report
	Metadata struct {
		Config     config.Global
		FinishedAt time.Time
	}
}

// New returns an Output initialized with rep and cfg.
func New(rep requester.Report, cfg config.Global) *Output {
	return &Output{
		Report: rep,
		Metadata: struct {
			Config     config.Global
			FinishedAt time.Time
		}{
			Config:     cfg,
			FinishedAt: time.Now(),
		},
	}
}

// Export exports an Output using the Strategies set in the attached
// config.Global. If any error occurs for a given Strategy, it does not
// block the other exports and returns an ExportError listing the errors.
func (o Output) Export() error {
	var ok bool
	var errs []error

	s := exportStrategy(o.Metadata.Config.Output.Out)
	if s.is(Stdout) {
		fmt.Println(ansi.Bold("→ Summary"))
		export.Stdout(o)
		ok = true
	}
	if s.is(JSONFile) {
		filename := genFilename()
		if err := export.JSONFile(filename, o); err != nil {
			errs = append(errs, err)
		} else {
			fmt.Println(ansi.Bold("→ JSON generated"))
			fmt.Println(filename)
		}
		ok = true
	}
	if s.is(Benchttp) {
		if err := export.HTTP(o); err != nil {
			errs = append(errs, err)
		}
		fmt.Println(ansi.Bold("→ Data sent to Benchttp"))
		ok = true
	}

	if !ok {
		return ErrInvalidStrategy
	}
	if len(errs) != 0 {
		return &ExportError{Errors: errs}
	}
	return nil
}

// export.Interface implementation

var _ export.Interface = (*Output)(nil)

// String returns a default summary of an Output as a string.
func (o Output) String() string {
	line := func(name string, value interface{}) string {
		const pattern = "%-18s %v\n"
		return fmt.Sprintf(pattern, name, value)
	}

	msString := func(d time.Duration) string {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}

	formatRequests := func(n, max int) string {
		maxString := strconv.Itoa(max)
		if maxString == "-1" {
			maxString = "∞"
		}
		return fmt.Sprintf("%d/%s", n, maxString)
	}

	var (
		b strings.Builder

		cfg            = o.Metadata.Config
		rep            = o.Report
		min, max, mean = rep.Stats()
	)

	b.WriteString(line("Endpoint", cfg.Request.URL))
	b.WriteString(line("Requests", formatRequests(rep.Length, cfg.Runner.Requests)))
	b.WriteString(line("Errors", rep.Fail))
	b.WriteString(line("Min response time", msString(min)))
	b.WriteString(line("Max response time", msString(max)))
	b.WriteString(line("Mean response time", msString(mean)))
	b.WriteString(line("Test duration", msString(rep.Duration)))
	return b.String()
}

// HTTPRequest returns the *http.Request to be sent to Benchttp server.
// The output is encoded as gob in the request body.
func (o Output) HTTPRequest() (*http.Request, error) {
	// Encode request body as gob
	b, err := encodeGob(o)
	if err != nil {
		return nil, err
	}

	// Create request
	r, err := http.NewRequest("POST", benchttpEndpoint, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	return r, nil
}

// helpers

// encodeGob encodes the given Output as gob-encoded bytes.
func encodeGob(o Output) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(o); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// genFilename generates a JSON file name suffixed with a timestamp
// located in the working directory.
func genFilename() string {
	return fmt.Sprintf("./benchttp.report.%s.json", timestamp())
}

// timestamp returns the current time in format yy-mm-ddThh:mm:ssZhh:mm.
func timestamp() string {
	now := time.Now().UTC()
	y, m, d := now.Date()
	hh, mm, ss := now.Clock()
	return strings.ReplaceAll(
		fmt.Sprintf("%4d%2d%2d%2d%2d%2d", y, m, d, hh, mm, ss),
		" ", "0",
	)
}
