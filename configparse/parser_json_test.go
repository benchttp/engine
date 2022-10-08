package configparse_test

import (
	"testing"

	"github.com/benchttp/engine/configparse"
)

func TestJSONParser(t *testing.T) {
	t.Run("return expected errors", func(t *testing.T) {
		testcases := []struct {
			label string
			in    []byte
			exp   string
		}{
			{
				label: "syntax error",
				in:    []byte("{\n  \"runner\": {},\n}\n"),
				exp:   "syntax error near 19: invalid character '}' looking for beginning of object key string",
			},
			{
				label: "unknown field",
				in:    []byte("{\n  \"notafield\": 123\n}\n"),
				exp:   `invalid field ("notafield"): does not exist`,
			},
			{
				label: "wrong type",
				in:    []byte("{\n  \"runner\": {\n    \"requests\": [123]\n  }\n}\n"),
				exp:   "wrong type for field runner.requests: want int, got array",
			},
			{
				label: "valid config",
				in:    []byte("{\n  \"runner\": {\n    \"requests\": 123\n  }\n}\n"),
				exp:   "",
			},
		}

		for _, tc := range testcases {
			t.Run(tc.label, func(t *testing.T) {
				var (
					parser configparse.JSONParser
					rawcfg configparse.Representation
				)

				gotErr := parser.Parse(tc.in, &rawcfg)

				if tc.exp == "" {
					if gotErr != nil {
						t.Fatalf("unexpected error: %v", gotErr)
					}
					return
				}

				if gotErr.Error() != tc.exp {
					t.Errorf(
						"unexpected error messages:\nexp %s\ngot %v",
						tc.exp, gotErr,
					)
				}
			})
		}
	})
}
