package configio_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/benchttp/sdk/benchttp"
	"github.com/benchttp/sdk/configio"
	"github.com/benchttp/sdk/configio/internal/testdata"
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
				label:  "empty path",
				file:   testdata.ConfigFile{Path: ""},
				expErr: configio.ErrFileNotFound,
			},
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
			{
				label:  "empty reference",
				file:   testdata.InvalidExtendsEmpty(),
				expErr: configio.ErrFileNotFound,
			},
		}

		for _, tc := range testcases {
			t.Run(tc.label, func(t *testing.T) {
				runner := benchttp.Runner{}
				err := configio.UnmarshalFile(tc.file.Path, &runner)

				assertError(t, tc.expErr, err)
				assertEqualRunners(t, tc.file.Runner, runner)
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
				assertEqualRunners(t, tc.file.Runner, runner)
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
		assertEqualRunners(t, cfg.Runner, runner)
	})

	t.Run("keep dst values not set in config", func(t *testing.T) {
		const keptConcurrency = 5 // not set in config file

		cfg := testdata.ValidPartial()
		exp := cfg.Runner
		exp.Concurrency = keptConcurrency
		dst := benchttp.Runner{Concurrency: keptConcurrency}

		err := configio.UnmarshalFile(cfg.Path, &dst)

		mustAssertNilError(t, err)
		assertEqualRunners(t, exp, dst)
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
				assertEqualRunners(t, tc.cfg.Runner, dst)
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

func assertError(t *testing.T, exp, got error) {
	t.Helper()
	if !errors.Is(got, exp) {
		t.Errorf("unexpected error:\nexp %v\ngot %v", exp, got)
	}
}

func assertEqualRunners(t *testing.T, exp, got benchttp.Runner) {
	t.Helper()

	opts := cmp.Options{
		cmp.Comparer(compareHTTPRequests),
		cmpopts.IgnoreUnexported(benchttp.Runner{}),
	}

	if !cmp.Equal(exp, got, opts...) {
		t.Errorf("unexpected runner:\n%s", cmp.Diff(exp, got, opts...))
	}
}

func compareHTTPRequests(a, b *http.Request) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}

	return cmp.Equal(a, b,
		cmp.Comparer(compareHeaders),
		cmp.Comparer(compareBodies),
		cmpopts.IgnoreUnexported(http.Request{}),
		cmpopts.IgnoreFields(http.Request{},
			"Proto", "ProtoMajor", "ProtoMinor", "GetBody", "ContentLength",
			"TransferEncoding", "Close", "Host", "Form", "PostForm",
			"MultipartForm", "Trailer", "RemoteAddr", "RequestURI", "TLS",
			"Cancel", "Response",
		),
	)
}

func compareHeaders(a, b http.Header) bool {
	return cmp.Equal(a, b, cmpopts.EquateEmpty())
}

func compareBodies(a, b io.ReadCloser) bool {
	isEmpty := func(r io.ReadCloser) bool {
		return r == nil || r == http.NoBody
	}

	if aEmpty, bEmpty := isEmpty(a), isEmpty(b); aEmpty || bEmpty {
		return aEmpty && bEmpty
	}

	defer a.Close()
	defer b.Close()
	ba, err := io.ReadAll(a)
	if err != nil {
		panic(err)
	}
	bb, err := io.ReadAll(b)
	if err != nil {
		panic(err)
	}
	return bytes.Equal(ba, bb)
}
