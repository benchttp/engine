package configio_test

import (
	"testing"

	"github.com/benchttp/sdk/configio"
)

var (
	goodFileYML  = configPath("valid/benchttp.yml")
	goodFileJSON = configPath("valid/benchttp.json")
	badFile      = configPath("does-not-exist.json")
)

func TestFindFile(t *testing.T) {
	t.Run("return first existing file form input", func(t *testing.T) {
		files := []string{badFile, goodFileYML, goodFileJSON}

		if got := configio.FindFile(files...); got != goodFileYML {
			t.Errorf("did not retrieve good file: exp %s, got %s", goodFileYML, got)
		}
	})

	t.Run("return first existing file from defaults", func(t *testing.T) {
		configio.DefaultPaths = []string{badFile, goodFileYML, goodFileJSON}

		if got := configio.FindFile(); got != goodFileYML {
			t.Errorf("did not retrieve good file: exp %s, got %s", goodFileYML, got)
		}
	})

	t.Run("return empty string when no match", func(t *testing.T) {
		files := []string{badFile}

		if got := configio.FindFile(files...); got != "" {
			t.Errorf("retrieved unexpected file: %s", got)
		}
	})
}
