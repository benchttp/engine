package config

import "strings"

type OutputStrategy string

const (
	OutputBenchttp OutputStrategy = "benchttp"
	OutputJSON     OutputStrategy = "json"
	OutputStdout   OutputStrategy = "stdout"
)

func IsOutput(v string) bool {
	switch OutputStrategy(strings.ToLower(v)) {
	case OutputBenchttp, OutputJSON, OutputStdout:
		return true
	}
	return false
}
