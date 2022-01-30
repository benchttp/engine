package file

import "os"

// Find returns the first name tham matches a file path.
// If no match is found, it returns an empty string.
func Find(names []string) string {
	for _, path := range names {
		if _, err := os.Stat(path); err == nil { // err IS nil: file exists
			return path
		}
	}
	return ""
}
