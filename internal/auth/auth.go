package auth

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	// TokenDir is the directory for the token file.
	TokenDir = ".config/benchttp" // nolint:gosec // no creds
	// TokenName is the name for the token file.
	TokenName = "token.txt"

	// ErrTokenPath reports an error resolving a token path.
	ErrTokenPath = errors.New("path error")
	// ErrTokenRead reports an error reading a token file.
	ErrTokenRead = errors.New("invalid token")
	// ErrTokenSave reports an error saving a token file.
	ErrTokenSave = errors.New("cannot save token")
)

// ReadToken reads a token from a file and sets userToken to the retrieved
// value or returns a non-nil error that is either ErrTokenFind or ErrTokenRead.
func ReadToken() (string, error) {
	// Resolve token path
	tokenPath, err := tokenPath()
	if err != nil {
		return "", err
	}

	// Open token file and get its value
	b, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", fmt.Errorf("%w: %s: %s", ErrTokenPath, tokenPath, err)
	}

	return strings.TrimSpace(string(b)), nil
}

// SaveToken create a file to the default path and writes the token
// into it, or returns an error that is either ErrTokenFind or ErrTokenSave.
func SaveToken(token string) error {
	// Resolve token path
	tokenPath, err := tokenPath()
	if err != nil {
		return err
	}

	// Remove previous token file if it exists
	if _, err := os.Stat(tokenPath); !errors.Is(err, os.ErrNotExist) {
		if err := os.Remove(tokenPath); err != nil {
			return fmt.Errorf("%w: %s: %s", ErrTokenPath, tokenPath, err)
		}
	}

	// Create new token file
	f, err := os.Create(tokenPath)
	if err != nil {
		return fmt.Errorf("%w: %s: %s", ErrTokenSave, tokenPath, err)
	}

	// Write token to file
	if _, err := f.WriteString(token + "\n"); err != nil {
		return fmt.Errorf("%w: %s: %s", ErrTokenSave, tokenPath, err)
	}

	return nil
}

// DeleteToken removes the content of the token file.
func DeleteToken() error {
	return SaveToken("")
}

// tokenPath resolves the default path for the token file. It retrieves
// the user's home directory and joins TokenDir and TokenName to it.
// If it fails it returns an ErrTokenFind error.
func tokenPath() (string, error) {
	// Retrieve user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrTokenPath, err)
	}

	// Resolve TokenDir path, making directories if needed
	dir := filepath.Join(home, TokenDir)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", fmt.Errorf("%w: %s", ErrTokenPath, err)
	}

	// Add TokenName to final path
	return filepath.Join(dir, TokenName), nil
}
