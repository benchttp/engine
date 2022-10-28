package configparse

import "os"

var DefaultPaths = []string{
	"./.benchttp.yml",
	"./.benchttp.yaml",
	"./.benchttp.json",
}

// Find returns the first name that matches a file path.
// If input paths is empty, it uses DefaultPaths.
// If no match is found, it returns an empty string.
func Find(paths ...string) string {
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
