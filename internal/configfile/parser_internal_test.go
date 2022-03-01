package configfile

import (
	"errors"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestYAMLParser(t *testing.T) {
	t.Run("return expected errors", func(t *testing.T) {
		testcases := []struct {
			label  string
			in     []byte
			expErr error
		}{
			{
				label: "unknown field",
				in:    []byte("notafield: 123\n"),
				expErr: &yaml.TypeError{
					Errors: []string{
						`line 1: invalid field ("notafield"): does not exist`,
					},
				},
			},
			{
				label: "wrong type unknown value",
				in:    []byte("runner:\n  requests: [123]\n"),
				expErr: &yaml.TypeError{
					Errors: []string{
						`line 2: wrong type: want int`,
					},
				},
			},
			{
				label: "wrong type known value",
				in:    []byte("runner:\n  requests: \"123\"\n"),
				expErr: &yaml.TypeError{
					Errors: []string{
						`line 2: wrong type ("123"): want int`,
					},
				},
			},
			{
				label: "cumulate errors",
				in:    []byte("runner:\n  requests: [123]\n  concurrency: \"123\"\nnotafield: 123\n"),
				expErr: &yaml.TypeError{
					Errors: []string{
						`line 2: wrong type: want int`,
						`line 3: wrong type ("123"): want int`,
						`line 4: invalid field ("notafield"): does not exist`,
					},
				},
			},
			{
				label:  "no errors custom fields",
				in:     []byte("x-data: &count\n  requests: 100\rrunner:\n  <<: *count\n"),
				expErr: nil,
			},
		}

		for _, tc := range testcases {
			t.Run(tc.label, func(t *testing.T) {
				var (
					parser  yamlParser
					rawcfg  unmarshaledConfig
					yamlErr *yaml.TypeError
				)

				gotErr := parser.parse(tc.in, &rawcfg)

				if tc.expErr == nil {
					if gotErr != nil {
						t.Fatalf("unexpected error: %v", gotErr)
					}
					return
				}

				if !errors.As(gotErr, &yamlErr) && tc.expErr != nil {
					t.Fatalf("unexpected error: %v", gotErr)
				}

				if !reflect.DeepEqual(yamlErr, tc.expErr) {
					t.Errorf("unexpected error messages:\nexp %v\ngot %v", tc.expErr, yamlErr)
				}
			})
		}
	})
}

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
					parser jsonParser
					rawcfg unmarshaledConfig
				)

				gotErr := parser.parse(tc.in, &rawcfg)

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
