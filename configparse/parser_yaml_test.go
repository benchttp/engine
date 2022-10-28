package configparse_test

import (
	"errors"
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/benchttp/sdk/configparse"
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
					parser  configparse.YAMLParser
					rawcfg  configparse.Representation
					yamlErr *yaml.TypeError
				)

				gotErr := parser.Parse(tc.in, &rawcfg)

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
