package configparse_test

import (
	"errors"
	"net/url"
	"reflect"
	"testing"

	"github.com/benchttp/runner/config"
	"github.com/benchttp/runner/internal/configparse"
)

const validInput = `
{
  "request": {
    "url": "https://example.com"
  },
  "runner": {
    "concurrency": 3
  }
}`

const inputWithBadKeys = `
{
  "request": {
    "url": "https://example.com"
  },
  "runner": {
    "badkey": "marcel patulacci"
  }
}`

const inputWithBadValues = `
{
  "request": {
    "url": "https://example.com"
  },
  "runner": {
    "concurrency": "bad value"
  }
}`

func TestJSON(t *testing.T) {
	testcases := []struct {
		name      string
		input     []byte
		expConfig config.Global
		expError  error
	}{
		{
			name:      "returns error if input json has bad keys",
			input:     []byte(inputWithBadKeys),
			expConfig: config.Global{},
			expError:  errors.New(`invalid field ("badkey"): does not exist`),
		},
		{
			name:      "returns error if input json has bad values",
			input:     []byte(inputWithBadValues),
			expConfig: config.Global{},
			expError:  errors.New(`wrong type for field runner.concurrency: want int, got string`),
		},
		{
			name:  "unmarshals JSON config and merges it with default",
			input: []byte(validInput),
			expConfig: config.Default().Override(
				config.Global{
					Request: config.Request{
						URL: mustParseURL("https://example.com"),
					},
					Runner: config.Runner{
						Concurrency: 3,
					},
				},
				"url",
				"concurrency",
			),
			expError: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gotConfig, gotError := configparse.JSON(tc.input)
			if !reflect.DeepEqual(gotConfig, tc.expConfig) {
				t.Errorf("unexpected config:\nexp %+v\ngot %+v", tc.expConfig, gotConfig)
			}

			if !sameErrors(gotError, tc.expError) {
				t.Errorf("unexpected error:\nexp %v,\ngot %v", tc.expError, gotError)
			}
		})
	}
}

func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}

func sameErrors(a, b error) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Error() == b.Error()
}
