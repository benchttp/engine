package configparse_test

import (
	"encoding/json"
	"errors"
	"net/url"
	"reflect"
	"testing"

	"github.com/benchttp/engine/config"
	"github.com/benchttp/engine/internal/configparse"
)

func TestJSON(t *testing.T) {
	baseInput := object{
		"request": object{
			"url": "https://example.com",
		},
	}

	testcases := []struct {
		name      string
		input     []byte
		expConfig config.Global
		expError  error
	}{
		{
			name: "returns error if input json has bad keys",
			input: baseInput.assign(object{
				"badkey": "marcel-patulacci",
			}).json(),
			expConfig: config.Global{},
			expError:  errors.New(`invalid field ("badkey"): does not exist`),
		},
		{
			name: "returns error if input json has bad values",
			input: baseInput.assign(object{
				"runner": object{
					"concurrency": "bad value", // want int
				},
			}).json(),
			expConfig: config.Global{},
			expError:  errors.New(`wrong type for field runner.concurrency: want int, got string`),
		},
		{
			name: "unmarshals JSON config and merges it with default",
			input: baseInput.assign(object{
				"runner": object{"concurrency": 3},
			}).json(),
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

type object map[string]interface{}

func (o object) json() []byte {
	b, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}
	return b
}

func (o object) assign(other object) object {
	newObject := object{}
	for k, v := range o {
		newObject[k] = v
	}
	for k, v := range other {
		newObject[k] = v
	}
	return newObject
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

func TestMap(t *testing.T) {
	base := object{"a": "a"}
	base.assign(object{"a": "CHANGED"})
	if base["a"] != "a" {
		t.Errorf("base map was mutated")
	}
}
