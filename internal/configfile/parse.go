package configfile

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/benchttp/engine/runner"
	"github.com/benchttp/engine/runner/configio"
)

// configFileRepr represents a raw unmarshaled config file.
// It implements configio.Interface.
type configFileRepr struct {
	Extends      *string          `yaml:"extends" json:"extends"`
	configio.DTO `yaml:",inline"` // do not read as field
}

// Parse parses a benchttp runner config file into a runner.ConfigGlobal
// and returns it or the first non-nil error occurring in the process,
// which can be any of the values declared in the package.
func Parse(filename string) (cfg runner.Config, err error) {
	rawConfigs, err := parseFileRecursive(filename, []configio.Interface{}, set{})
	if err != nil {
		return
	}
	return configio.ParseManyWithDefault(rawConfigs)
}

// set is a collection of unique string values.
type set map[string]bool

// add adds v to the receiver. If v is already set, it returns a non-nil
// error instead.
func (s set) add(v string) error {
	if _, exists := s[v]; exists {
		return errors.New("value already set")
	}
	s[v] = true
	return nil
}

// parseFileRecursive parses a config file and its parent found from key
// "extends" recursively until the root config file is reached.
// It returns the list of all parsed configs or the first non-nil error
// occurring in the process.
func parseFileRecursive(
	filename string,
	rawConfigs []configio.Interface,
	seen set,
) ([]configio.Interface, error) {
	// avoid infinite recursion caused by circular reference
	if err := seen.add(filename); err != nil {
		return rawConfigs, ErrCircularExtends
	}

	// parse current file, append parsed config
	raw, err := parseFile(filename)
	if err != nil {
		return rawConfigs, err
	}
	rawConfigs = append(rawConfigs, raw)

	// root config reached: stop now and return the parsed configs
	if raw.Extends == nil {
		return rawConfigs, nil
	}

	// config has parent: resolve its path and parse it recursively
	parentPath := filepath.Join(filepath.Dir(filename), *raw.Extends)
	return parseFileRecursive(parentPath, rawConfigs, seen)
}

// parseFile parses a single config file and returns the result as an
// unmarshaledConfig and an appropriate error predeclared in the package.
func parseFile(filename string) (raw configFileRepr, err error) {
	b, err := os.ReadFile(filename)
	switch {
	case err == nil:
	case errors.Is(err, os.ErrNotExist):
		return raw, errWithDetails(ErrFileNotFound, filename)
	default:
		return raw, errWithDetails(ErrFileRead, filename, err)
	}

	ext := extension(filepath.Ext(filename))
	parser, err := newParser(ext)
	if err != nil {
		return raw, errWithDetails(ErrFileExt, ext, err)
	}

	if err = parser.parse(b, &raw); err != nil {
		return raw, errWithDetails(ErrParse, filename, err)
	}

	return raw, nil
}
