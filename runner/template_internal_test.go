package runner

import (
	"errors"
	"testing"
)

func TestReport_applyTemplate(t *testing.T) {
	testcases := []struct {
		label   string
		pattern string
		expStr  string
		expErr  error
	}{
		{
			label:   "return errTemplateEmpty if pattern is empty",
			pattern: "",
			expStr:  "",
			expErr:  errTemplateEmpty,
		},
		{
			label:   "return errTemplateSyntaxt if pattern has syntax error",
			pattern: "{{ else }}",
			expStr:  "",
			expErr:  errTemplateSyntax,
		},
		{
			label:   "return errTemplateSyntaxt if pattern doesn't match report values",
			pattern: "{{ .Foo }}", // Report.Foo doesn't exist
			expStr:  "",
			expErr:  errTemplateSyntax,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.label, func(t *testing.T) {
			r := &Report{}
			gotStr, gotErr := r.applyTemplate(tc.pattern)
			if !errors.Is(gotErr, tc.expErr) {
				t.Errorf("unexpected error: %v", gotErr)
			}
			if gotStr != tc.expStr {
				t.Errorf("unexpected string: %q", gotStr)
			}
		})
	}
}

func TestReport_templateFuncs(t *testing.T) {
	t.Run("fail", func(t *testing.T) {
		rep := Report{}

		untypedFailFunc := retrieveTemplateFuncOrFatal(t, &rep, "fail")

		failFunc, ok := untypedFailFunc.(func(...interface{}) string)
		if !ok {
			t.Fatalf("wrong type:\nexp func(...interface{}) string\ngot %T", untypedFailFunc)
		}

		if got := failFunc("a", "b", "c"); got != "" {
			t.Errorf("unexpected output: exp always %q, got %q", "", got)
		}

		gotErr := rep.errTemplateFailTriggered
		if !errors.Is(gotErr, ErrTemplateFailTriggered) {
			t.Fatalf("unexpected error:\nexp ErrTemplateFailTriggered\ngot %v", gotErr)
		}
		if gotMsg, expMsg := gotErr.Error(), "test failed: abc"; gotMsg != expMsg {
			t.Errorf("unexpected error message:\nexp %q\ngot %q", expMsg, gotMsg)
		}
	})
}

// helpers

func retrieveTemplateFuncOrFatal(t *testing.T, r *Report, name string) interface{} {
	t.Helper()
	v, exists := r.templateFuncs()[name]
	if !exists {
		t.Fatalf("template func %q does not exist", name)
	}
	return v
}
