package ansi

import (
	"fmt"
	"strings"
)

type style string

const (
	reset style = "\033[0m"

	bold style = "\033[1m"

	grey   style = "\033[1;30m"
	red    style = "\033[1;31m"
	green  style = "\033[1;32m"
	yellow style = "\033[1;33m"
	cyan   style = "\033[1;36m"

	erase style = "\033[1A"
)

func withStyle(in string, s style) string {
	return fmt.Sprintf("%s%s%s", s, in, reset)
}

// Bold returns the bold version of the input string.
func Bold(in string) string {
	return withStyle(in, bold)
}

// Green returns the green version of the input string.
func Green(in string) string {
	return withStyle(in, green)
}

// Yellow returns the yellow version of the input string.
func Yellow(in string) string {
	return withStyle(in, yellow)
}

// Cyan returns the cyan version of the input string.
func Cyan(in string) string {
	return withStyle(in, cyan)
}

// Red returns the red version of the input string.
func Red(in string) string {
	return withStyle(in, red)
}

// Grey returns the grey version of the input string.
func Grey(in string) string {
	return withStyle(in, grey)
}

// Erase returns a string that erases the previous line n times.
func Erase(n int) string {
	if n < 1 {
		return ""
	}
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.Write([]byte(erase))
	}
	return b.String()
}
