package configio

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/benchttp/sdk/benchttp"

	"github.com/benchttp/sdk/internal/errorutil"
)

// DefaultPaths is the default list of paths looked up by FindFile when
// called without parameters.
var DefaultPaths = []string{
	"./.benchttp.yml",
	"./.benchttp.yaml",
	"./.benchttp.json",
}

// FindFile returns the first name that matches a file path.
// If input paths is empty, it uses DefaultPaths.
// If no match is found, it returns an empty string.
func FindFile(paths ...string) string {
	if len(paths) == 0 {
		paths = DefaultPaths
	}
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil { // err IS nil: file exists
			return path
		}
	}
	return ""
}

// UnmarshalFile parses given filename as a benchttp runner configuration
// into a runner.Runner and stores the retrieved values into *dst.
// It returns the first error occurring in the process, which can be
// any of the values declared in the package.
func UnmarshalFile(filename string, dst *benchttp.Runner) error {
	f, err := file{path: filename}.decodeAll()
	if err != nil {
		return err
	}
	return f.reprs().mergeInto(dst)
}

// file represents a config file
type file struct {
	prev *file
	path string
	repr representation
}

// decodeAll reads f.path as a file and decodes it into f.repr.
// If the decoded file references another file, the operation
// is repeated recursively until root file is reached.
func (f file) decodeAll() (file, error) {
	if err := f.decode(); err != nil {
		return file{}, err
	}

	if isRoot := f.repr.Extends == nil; isRoot {
		return f, nil
	}

	nextPath := filepath.Join(filepath.Dir(f.path), *f.repr.Extends)
	if f.seen(nextPath) {
		return file{}, errorutil.WithDetails(ErrFileCircular, nextPath)
	}

	return f.extend(nextPath).decodeAll()
}

// decode reads f.path as a file and decodes it into f.repr.
func (f *file) decode() (err error) {
	b, err := os.ReadFile(f.path)
	switch {
	case err == nil:
	case errors.Is(err, os.ErrNotExist):
		return errorutil.WithDetails(ErrFileNotFound, f.path)
	default:
		return errorutil.WithDetails(ErrFileRead, f.path, err)
	}

	ext, err := f.format()
	if err != nil {
		return err
	}

	if err := DecoderOf(ext, b).Decode(&f.repr); err != nil {
		return errorutil.WithDetails(ErrFileParse, f.path, err)
	}

	return nil
}

func (f file) format() (Format, error) {
	switch ext := filepath.Ext(f.path); ext {
	case ".yml", ".yaml":
		return FormatYAML, nil
	case ".json":
		return FormatJSON, nil
	default:
		return "", errorutil.WithDetails(ErrFileExt, ext, f.path)
	}
}

func (f file) extend(nextPath string) file {
	return file{prev: &f, path: nextPath}
}

// seen returns true if the given path has already been decoded
// by the receiver or any of its ancestors.
func (f file) seen(p string) bool {
	if f.path == "" || p == "" {
		panicInternal("file.seen", "empty f.path or p")
	}
	return f.path == p || (f.prev != nil && f.prev.seen(p))
}

// reprs returns a slice of Representation, starting with the receiver
// and ending with the last child.
func (f file) reprs() representations {
	reprs := []representation{f.repr}
	if f.prev != nil {
		reprs = append(reprs, f.prev.reprs()...)
	}
	return reprs
}
