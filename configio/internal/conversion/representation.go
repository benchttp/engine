package conversion

import (
	"fmt"

	"github.com/benchttp/sdk/benchttp"
)

// Repr is a raw data model for formatted runner config (json, yaml).
// It serves as a receiver for unmarshaling processes and for that reason
// its types are kept simple (certain types are incompatible with certain
// unmarshalers).
// It exposes a method Unmarshal to convert its values into a runner.Config.
type Repr struct {
	Extends *string        `yaml:"extends" json:"extends"`
	Request requestRepr    `yaml:"request" json:"request"`
	Runner  runnerRepr     `yaml:"runner" json:"runner"`
	Tests   []testcaseRepr `yaml:"tests" json:"tests"`
}

type requestRepr struct {
	Method      *string             `yaml:"method" json:"method"`
	URL         *string             `yaml:"url" json:"url"`
	QueryParams map[string]string   `yaml:"queryParams" json:"queryParams"`
	Header      map[string][]string `yaml:"header" json:"header"`
	Body        *struct {
		Type    string `yaml:"type" json:"type"`
		Content string `yaml:"content" json:"content"`
	} `yaml:"body" json:"body"`
}

type runnerRepr struct {
	Requests       *int    `yaml:"requests" json:"requests"`
	Concurrency    *int    `yaml:"concurrency" json:"concurrency"`
	Interval       *string `yaml:"interval" json:"interval"`
	RequestTimeout *string `yaml:"requestTimeout" json:"requestTimeout"`
	GlobalTimeout  *string `yaml:"globalTimeout" json:"globalTimeout"`
}

type testcaseRepr struct {
	Name      *string     `yaml:"name" json:"name"`
	Field     *string     `yaml:"field" json:"field"`
	Predicate *string     `yaml:"predicate" json:"predicate"`
	Target    interface{} `yaml:"target" json:"target"`
}

func (repr Repr) Validate() error {
	return repr.Decode(&benchttp.Runner{})
}

// Decode parses the Representation receiver as a benchttp.Runner
// and stores any non-nil field value into the corresponding field
// of dst.
func (repr Repr) Decode(dst *benchttp.Runner) error {
	for _, decoder := range converters {
		if err := decoder.decode(repr, dst); err != nil {
			return err
		}
	}
	return nil
}

type Reprs []Repr

// MergeInto successively parses the given representations into dst.
//
// The input Representation slice must never be nil or empty, otherwise it panics.
func (reprs Reprs) MergeInto(dst *benchttp.Runner) error {
	if len(reprs) == 0 { // supposedly catched upstream, should not occur
		panicInternal("parseAndMergeReprs", "nil or empty []Representation")
	}

	for _, repr := range reprs {
		if err := repr.Decode(dst); err != nil {
			return err
			// TODO: uncomment once wrapped from configio/file.go
			// return errorutil.WithDetails(ErrFileParse, err)
		}
	}
	return nil
}

func panicInternal(funcname, detail string) {
	const reportURL = "https://github.com/benchttp/sdk/issues/new"
	source := fmt.Sprintf("configio.%s", funcname)
	panic(fmt.Sprintf(
		"%s: unexpected internal error: %s, please file an issue at %s",
		source, detail, reportURL,
	))
}
