package configfile_test

import (
	"testing"

	"github.com/benchttp/runner/internal/configfile"
)

var (
	goodFileYML  = configPath("valid/benchttp.yml")
	goodFileJSON = configPath("valid/benchttp.json")
	badFile      = configPath("does-not-exist.json")
)

func TestFind(t *testing.T) {
	t.Run("return first existing file", func(t *testing.T) {
		files := []string{badFile, goodFileYML, goodFileJSON}

		if got := configfile.Find(files); got != goodFileYML {
			t.Errorf("did not retrieve good file: exp %s, got %s", goodFileYML, got)
		}
	})

	t.Run("return empty string when no match", func(t *testing.T) {
		files := []string{badFile}

		if got := configfile.Find(files); got != "" {
			t.Errorf("retrieved unexpected file: %s", got)
		}
	})
}
