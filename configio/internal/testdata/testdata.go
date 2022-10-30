package testdata

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"time"

	"github.com/benchttp/sdk/benchttp"
)

// ConfigFile represents a testdata configuration file.
type ConfigFile struct {
	// Path is the relative file path from configio.
	Path string
	// Runner is the expected benchttp.Runner.
	Runner benchttp.Runner
}

// ValidFullJSON returns a valid full configuration file.
func ValidFullJSON() ConfigFile {
	return validConfig("full.json", kindFull)
}

func ValidFullYAML() ConfigFile {
	return validConfig("full.yaml", kindFull)
}

func ValidFullYML() ConfigFile {
	return validConfig("full.yml", kindFull)
}

func ValidPartial() ConfigFile {
	return validConfig("partial.yml", kindPartial)
}

func ValidExtends() ConfigFile {
	return validConfig("extends/child.yml", kindExtended)
}

func ValidExtendsNested() ConfigFile {
	return validConfig("extends/nest-0/nest-1/child.yml", kindExtended)
}

func InvalidPath() ConfigFile {
	return invalidConfig("does-not-exist.json")
}

func InvalidFieldsJSON() ConfigFile {
	return invalidConfig("fields.json")
}

func InvalidFieldsYML() ConfigFile {
	return invalidConfig("fields.yml")
}

func InvalidExtension() ConfigFile {
	return invalidConfig("extension.yams")
}

func InvalidExtendsCircular() ConfigFile {
	return invalidConfig("extends/circular-0.yml")
}

func InvalidExtendsSelf() ConfigFile {
	return invalidConfig("extends/circular-self.yml")
}

type kind uint8

const (
	kindFull kind = iota
	kindPartial
	kindExtended
)

var basePath = filepath.Join("internal", "testdata")

func validConfig(name string, k kind) ConfigFile {
	return ConfigFile{
		Path:   filepath.Join(basePath, "valid", name),
		Runner: runnerOf(k),
	}
}

func invalidConfig(name string) ConfigFile {
	return ConfigFile{
		Path:   filepath.Join(basePath, "invalid", name),
		Runner: benchttp.Runner{},
	}
}

func runnerOf(k kind) benchttp.Runner {
	switch k {
	case kindFull:
		return fullRunner()
	case kindPartial:
		return partialRunner()
	case kindExtended:
		return extendedRunner()
	default:
		panic("invalid kind")
	}
}

// fullRunner returns the expected runner from full configurations
func fullRunner() benchttp.Runner {
	request := httptest.NewRequest(
		"POST",
		"http://localhost:3000/benchttp?param0=value0&param1=value1",
		bytes.NewReader([]byte(`{"key0":"val0","key1":"val1"}`)),
	)
	request.Header = http.Header{
		"key0": []string{"val0", "val1"},
		"key1": []string{"val0"},
	}
	return benchttp.Runner{
		Request: request,

		Requests:       100,
		Concurrency:    1,
		Interval:       50 * time.Millisecond,
		RequestTimeout: 2 * time.Second,
		GlobalTimeout:  60 * time.Second,

		Tests: []benchttp.TestCase{
			{
				Name:      "maximum response time",
				Field:     "ResponseTimes.Max",
				Predicate: "LTE",
				Target:    120 * time.Millisecond,
			},
			{
				Name:      "100% availability",
				Field:     "RequestFailureCount",
				Predicate: "EQ",
				Target:    0,
			},
		},
	}
}

// partialRunner returns the expected runner from partial configurations
func partialRunner() benchttp.Runner {
	return benchttp.Runner{
		Request:       httptest.NewRequest("GET", "http://localhost:3000/partial", nil),
		GlobalTimeout: 42 * time.Second,
	}
}

// extendedRunner returns the expected runner from extending configurations.
func extendedRunner() benchttp.Runner {
	return benchttp.Runner{
		// child override
		Request: httptest.NewRequest("PUT", "http://localhost:3000/child", nil),
		// parent kept value
		GlobalTimeout: 42 * time.Second,
	}
}
