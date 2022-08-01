package ds_test

import (
	"reflect"
	"testing"

	"github.com/benchttp/engine/internal/ds"
)

func TestStringSet_Add(t *testing.T) {
	testcases := []struct {
		label  string
		input  []string
		expSet ds.StringSet
		expOks []bool
	}{
		{
			label:  "adds new values and returns true",
			input:  []string{"a", "b", "c"},
			expOks: []bool{true, true, true},
			expSet: ds.StringSet{
				"a": struct{}{},
				"b": struct{}{},
				"c": struct{}{},
			},
		},
		{
			label:  "noop and returns false on existing values",
			input:  []string{"a", "a", "b"},
			expOks: []bool{true, false, true},
			expSet: ds.StringSet{
				"a": struct{}{},
				"b": struct{}{},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			set := ds.StringSet{}

			for i, v := range tc.input {
				if gotOk, expOk := set.Add(v), tc.expOks[i]; gotOk != expOk {
					t.Errorf("exp ok == %v, got %v", gotOk, expOk)
				}
			}

			if !reflect.DeepEqual(set, tc.expSet) {
				t.Errorf("exp %v\ngot %v", tc.expSet, set)
			}
		})
	}
}
