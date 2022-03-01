package main

import "fmt"

// benchttpVersion is the current version of benchttp
// as output by `benchttp version`. It is assumed to be set
// by `go build -ldflags "-X main.benchttpVersion=<version>"`,
// allowing us to set the value dynamically at build time
// using latest git tag.
//
// Its default value "development" is only used when the app
// is ran locally without a build (e.g. `go run ./cmd/benchttp`).
var benchttpVersion = "development"

// cmdVersion handles subcommand "benchttp version".
type cmdVersion struct{}

// ensure cmdVersion implements command
var _ command = (*cmdVersion)(nil)

func (cmdVersion) execute([]string) error {
	fmt.Println("benchttp", benchttpVersion)
	return nil
}
