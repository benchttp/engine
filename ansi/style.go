package ansi

import (
	"fmt"
	"strings"
)

type style string

const (
	reset style = "\033[0m"

	erase style = "\033[1A"

	grey   style = "\033[1;30m"
	red    style = "\033[1;31m"
	green  style = "\033[1;32m"
	yellow style = "\033[1;33m"
	cyan   style = "\033[1;36m"
)

func withStyle(in string, s style) string {
	return fmt.Sprintf("%s%s%s", s, in, reset)
}

func Green(in string) string {
	return withStyle(in, green)
}

func Yellow(in string) string {
	return withStyle(in, yellow)
}

func Cyan(in string) string {
	return withStyle(in, cyan)
}

func Red(in string) string {
	return withStyle(in, red)
}

func Grey(in string) string {
	return withStyle(in, grey)
}

func Erase(lines int) string {
	if lines < 1 {
		return ""
	}
	var b strings.Builder
	for i := 0; i < lines; i++ {
		b.Write([]byte(erase))
	}
	return b.String()
}
