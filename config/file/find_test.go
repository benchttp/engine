package file_test

import (
	"testing"

	"github.com/benchttp/runner/config/file"
)

var (
	goodFile = "../../test/testdata/config/benchttp.yml"
	badFile  = "./hello.yml"
)

func TestFind(t *testing.T) {
	t.Run("return first existing file", func(t *testing.T) {
		otherGoodFile := "../../test/testdata/config/benchttp.json"
		files := []string{badFile, goodFile, otherGoodFile}

		if got := file.Find(files); got != goodFile {
			t.Errorf("did not retrieve good file: exp %s, got %s", goodFile, got)
		}
	})

	t.Run("return empty string when no match", func(t *testing.T) {
		files := []string{badFile}

		if got := file.Find(files); got != "" {
			t.Errorf("retrieved unexpected file: %s", got)
		}
	})
}
