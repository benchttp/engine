package configflags

import (
	"flag"

	"github.com/benchttp/runner/config"
)

// Which returns a slice of all config fields set via the CLI
// for the given *flag.FlagSet.
func Which(flagset *flag.FlagSet) []string {
	var fields []string
	flagset.Visit(func(f *flag.Flag) {
		if name := f.Name; config.IsField(name) {
			fields = append(fields, name)
		}
	})
	return fields
}
