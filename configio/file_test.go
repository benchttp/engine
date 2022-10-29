package configio_test

import (
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/benchttp/sdk/benchttp"
	"github.com/benchttp/sdk/benchttptest"
	"github.com/benchttp/sdk/configio"
	"github.com/benchttp/sdk/configio/testdata"
)

func TestFindFile(t *testing.T) {
	var (
		fileYAML = testdata.ValidFullYAML().Path
		fileJSON = testdata.ValidFullJSON().Path
		nofile   = testdata.InvalidPath().Path
	)

	testcases := []struct {
		name         string
		inputPaths   []string
		defaultPaths []string
		exp          string
	}{
		{
			name:         "return first existing input path",
			inputPaths:   []string{nofile, fileYAML, fileJSON},
			defaultPaths: []string{},
			exp:          fileYAML,
		},
		{
			name:         "return first existing default path if no input",
			inputPaths:   []string{},
			defaultPaths: []string{nofile, fileYAML, fileJSON},
			exp:          fileYAML,
		},
		{
			name:         "return empty string if no file found",
			inputPaths:   []string{nofile},
			defaultPaths: []string{},
			exp:          "",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			if len(tc.defaultPaths) > 0 {
				configio.DefaultPaths = tc.defaultPaths
			}
			got := configio.FindFile(tc.inputPaths...)
			if got != tc.exp {
				t.Errorf("exp %q, got %q", tc.exp, got)
			}
		})
	}
}

func TestUnmarshalFile(t *testing.T) {
	t.Run("return file errors early", func(t *testing.T) {
		testcases := []struct {
			label  string
			file   testdata.ConfigFile
			expErr error
		}{
			{
				label:  "not found",
				file:   testdata.InvalidPath(),
				expErr: configio.ErrFileNotFound,
			},
			{
				label:  "unsupported extension",
				file:   testdata.InvalidExtension(),
				expErr: configio.ErrFileExt,
			},
			{
				label:  "yaml invalid fields",
				file:   testdata.InvalidFieldsYML(),
				expErr: configio.ErrFileParse,
			},
			{
				label:  "json invalid fields",
				file:   testdata.InvalidFieldsJSON(),
				expErr: configio.ErrFileParse,
			},
			{
				label:  "self reference",
				file:   testdata.InvalidExtendsSelf(),
				expErr: configio.ErrFileCircular,
			},
			{
				label:  "circular reference",
				file:   testdata.InvalidExtendsCircular(),
				expErr: configio.ErrFileCircular,
			},
		}

		for _, tc := range testcases {
			t.Run(tc.label, func(t *testing.T) {
				runner := benchttp.Runner{}
				err := configio.UnmarshalFile(tc.file.Path, &runner)

				assertStaticError(t, tc.expErr, err)
				benchttptest.AssertEqualRunners(t, tc.file.Runner, runner)
			})
		}
	})

	t.Run("happy path all extensions", func(t *testing.T) {
		for _, tc := range []struct {
			name string
			file testdata.ConfigFile
		}{
			{name: "full json", file: testdata.ValidFullJSON()},
			{name: "full yaml", file: testdata.ValidFullYAML()},
			{name: "full yml", file: testdata.ValidFullYML()},
		} {
			t.Run(tc.name, func(t *testing.T) {
				runner := benchttp.Runner{}
				err := configio.UnmarshalFile(tc.file.Path, &runner)

				mustAssertNilError(t, err)
				benchttptest.AssertEqualRunners(t, tc.file.Runner, runner)
			})
		}
	})

	t.Run("override dst with set config values", func(t *testing.T) {
		cfg := testdata.ValidPartial()
		runner := benchttp.Runner{
			Request:       httptest.NewRequest("GET", "http://a.b", nil), // overridden
			GlobalTimeout: 1 * time.Second,                               // overridden
		}

		err := configio.UnmarshalFile(cfg.Path, &runner)

		mustAssertNilError(t, err)
		benchttptest.AssertEqualRunners(t, cfg.Runner, runner)
	})

	t.Run("keep dst values not set in config", func(t *testing.T) {
		const keptConcurrency = 5 // not set in config file

		cfg := testdata.ValidPartial()
		exp := cfg.Runner
		exp.Concurrency = keptConcurrency
		dst := benchttp.Runner{Concurrency: keptConcurrency}

		err := configio.UnmarshalFile(cfg.Path, &dst)

		mustAssertNilError(t, err)
		benchttptest.AssertEqualRunners(t, exp, dst)
	})

	t.Run("extend config files", func(t *testing.T) {
		for _, tc := range []struct {
			name string
			cfg  testdata.ConfigFile
		}{
			{name: "same directory", cfg: testdata.ValidExtends()},
			{name: "nested directory", cfg: testdata.ValidExtendsNested()},
		} {
			t.Run(tc.name, func(t *testing.T) {
				dst := benchttp.Runner{}
				err := configio.UnmarshalFile(tc.cfg.Path, &dst)

				mustAssertNilError(t, err)
				benchttptest.AssertEqualRunners(t, tc.cfg.Runner, dst)
			})
		}
	})
}

// helpers

func mustAssertNilError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("exp nil error, got %v", err)
	}
}

func assertStaticError(t *testing.T, exp, got error) {
	t.Helper()
	if !errors.Is(got, exp) {
		t.Errorf("unexpected error:\nexp %v\ngot %v", exp, got)
	}
}
