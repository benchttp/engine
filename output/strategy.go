package output

import "github.com/benchttp/runner/config"

// Strategy reprents an output strategy for the input value.
type Strategy uint8

const (
	// StrategyStdout writes the input value to os.Stdout.
	Stdout Strategy = 1 << iota
	// JSONFile writes the input value as a file
	// in the working directory.
	JSONFile
	// Benchttp sends the input value to Benchttp API server.
	Benchttp
)

// is returns true if s matches the target strategy.
// s can match several strategies.
func (s Strategy) is(target Strategy) bool {
	return s&target != 0
}

// exportStrategy returns the strategy to use for exporting the report.
func exportStrategy(cfgStrategies []config.OutputStrategy) Strategy {
	var s Strategy
	for _, o := range cfgStrategies {
		if o == config.OutputStdout {
			s |= Stdout
		}
		if o == config.OutputJSON {
			s |= JSONFile
		}
		if o == config.OutputBenchttp {
			s |= Benchttp
		}
	}
	return s
}
