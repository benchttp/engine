package file

// rawConfig is a raw data model for runner config files.
// It serves as a receiver for unmarshaling processes and for that reason
// its types are kept simple (certain types are incompatible with certain
// unmarshalers).
type rawConfig struct {
	Request struct {
		Method      string            `yaml:"method" json:"method"`
		URL         string            `yaml:"url" json:"url"`
		QueryParams map[string]string `yaml:"queryParams" json:"queryParams"`
		Timeout     string            `yaml:"timeout" json:"timeout"`
	} `yaml:"request" json:"request"`

	RunnerOptions struct {
		Requests      int    `yaml:"requests" json:"requests"`
		Concurrency   int    `yaml:"concurrency" json:"concurrency"`
		GlobalTimeout string `yaml:"globalTimeout" json:"globalTimeout"`
	} `yaml:"runnerOptions" json:"runnerOptions"`
}
