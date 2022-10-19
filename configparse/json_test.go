package configparse_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/benchttp/engine/configparse"
	"github.com/benchttp/engine/runner"
)

func TestJSON(t *testing.T) {
	baseInput := object{
		"request": object{
			"url": "https://example.com",
		},
	}

	testcases := []struct {
		name          string
		input         []byte
		isValidRunner func(runner.Runner) bool
		expError      error
	}{
		{
			name: "returns error if input json has bad keys",
			input: baseInput.assign(object{
				"badkey": "marcel-patulacci",
			}).json(),
			isValidRunner: func(cfg runner.Runner) bool { return true },
			expError:      errors.New(`invalid field ("badkey"): does not exist`),
		},
		{
			name: "returns error if input json has bad values",
			input: baseInput.assign(object{
				"runner": object{
					"concurrency": "bad value", // want int
				},
			}).json(),
			isValidRunner: func(runner.Runner) bool { return true },
			expError:      errors.New(`wrong type for field runner.concurrency: want int, got string`),
		},
		{
			name: "unmarshals JSON config and merges it with default",
			input: baseInput.assign(object{
				"runner": object{"concurrency": 3},
			}).json(),
			isValidRunner: func(r runner.Runner) bool {
				defaultRunner := runner.DefaultRunner()

				isInputValueParsed := r.Concurrency == 3
				isMergedWithDefault := r.Request.Method == defaultRunner.Request.Method &&
					r.GlobalTimeout == defaultRunner.GlobalTimeout

				return isInputValueParsed && isMergedWithDefault
			},
			expError: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gotRunner, gotError := configparse.JSON(tc.input)
			if !tc.isValidRunner(gotRunner) {
				t.Errorf("unexpected config:\n%+v", gotRunner)
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

func sameErrors(a, b error) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Error() == b.Error()
}
