package output

import (
	"testing"

	"github.com/benchttp/runner/config"
)

type strategyTest struct {
	label  string
	base   Strategy
	target Strategy
	exp    bool
}

func (test strategyTest) Run(t *testing.T) {
	t.Helper()

	t.Run(test.label, func(t *testing.T) {
		if got := test.base.is(test.target); got != test.exp {
			t.Errorf("exp %v, got %v", test.exp, got)
		}
	})
}

func TestStrategy_is(t *testing.T) {
	t.Run("return false for non matching Strategy", func(t *testing.T) {
		for _, test := range []strategyTest{
			{
				label:  "invalid base Strategy",
				base:   0,
				target: Stdout,
				exp:    false,
			},
			{
				label:  "invalid target Strategy",
				base:   Stdout,
				target: 0,
				exp:    false,
			},
			{
				label:  "single Strategies",
				base:   JSONFile,
				target: Benchttp,
				exp:    false,
			},
			{
				label:  "combined base Strategies",
				base:   Stdout | JSONFile,
				target: Benchttp,
				exp:    false,
			},
			{
				label:  "combined target Strategies",
				base:   Stdout,
				target: JSONFile | Benchttp,
				exp:    false,
			},
		} {
			test.Run(t)
		}
	})

	t.Run("return true for matching Strategy", func(t *testing.T) {
		for _, test := range []strategyTest{
			{
				label:  "identity Stdout",
				base:   Stdout,
				target: Stdout,
				exp:    true,
			},
			{
				label:  "identity JSONFile",
				base:   JSONFile,
				target: JSONFile,
				exp:    true,
			},
			{
				label:  "identity Benchttp",
				base:   Benchttp,
				target: Benchttp,
				exp:    true,
			},
			{
				label:  "combined base Strategies",
				base:   Stdout | Benchttp,
				target: Benchttp,
				exp:    true,
			},
			{
				label:  "combined target Strategies",
				base:   Benchttp,
				target: Stdout | Benchttp,
				exp:    true,
			},
		} {
			test.Run(t)
		}
	})
}

func TestExportStrategy(t *testing.T) {
	testcases := []struct {
		label string
		in    []config.OutputStrategy
		exp   Strategy
	}{
		{
			label: "return no Strategy for empty input",
			in:    []config.OutputStrategy{},
			exp:   0,
		},
		{
			label: "return single Strategy Stdout",
			in:    []config.OutputStrategy{config.OutputStdout},
			exp:   Stdout,
		},
		{
			label: "return single Strategy JSONFile",
			in:    []config.OutputStrategy{config.OutputJSON},
			exp:   JSONFile,
		},
		{
			label: "return single Strategy Benchttp",
			in:    []config.OutputStrategy{config.OutputBenchttp},
			exp:   Benchttp,
		},
		{
			label: "return combined Strategies",
			in: []config.OutputStrategy{
				config.OutputStdout,
				config.OutputJSON,
				config.OutputBenchttp,
			},
			exp: Stdout | JSONFile | Benchttp,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			if got := exportStrategy(tc.in); got != tc.exp {
				t.Errorf("exp %d, got %d", tc.exp, got)
			}
		})
	}
}
