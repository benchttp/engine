package configparse_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/benchttp/sdk/benchttp"
	"github.com/benchttp/sdk/configparse"
)

func TestJSON(t *testing.T) {
	const testURL = "https://example.com"
	baseInput := object{
		"request": object{
			"url": testURL,
		},
	}

	testcases := []struct {
		name          string
		input         []byte
		isValidRunner func(base, got benchttp.Runner) bool
		expError      error
	}{
		{
			name: "returns error if input json has bad keys",
			input: baseInput.assign(object{
				"badkey": "marcel-patulacci",
			}).json(),
			isValidRunner: func(_, _ benchttp.Runner) bool { return true },
			expError:      errors.New(`invalid field ("badkey"): does not exist`),
		},
		{
			name: "returns error if input json has bad values",
			input: baseInput.assign(object{
				"runner": object{
					"concurrency": "bad value", // want int
				},
			}).json(),
			isValidRunner: func(_, _ benchttp.Runner) bool { return true },
			expError:      errors.New(`wrong type for field runner.concurrency: want int, got string`),
		},
		{
			name:  "unmarshals JSON config and merges it with base runner",
			input: baseInput.json(),
			isValidRunner: func(base, got benchttp.Runner) bool {
				isParsed := got.Request.URL.String() == testURL
				isMerged := got.GlobalTimeout == base.GlobalTimeout
				return isParsed && isMerged
			},
			expError: nil,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gotRunner := benchttp.DefaultRunner()
			gotError := configparse.JSON(tc.input, &gotRunner)

			if !tc.isValidRunner(benchttp.DefaultRunner(), gotRunner) {
				t.Errorf("unexpected runner:\n%+v", gotRunner)
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
	return (a == nil && b == nil) || !(a == nil || b == nil) || a.Error() == b.Error()
}
