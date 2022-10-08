package httpclient

import (
	"net"
	"strconv"
)

// FindFreePort returns an available port as a string or the first
// non-nil error occurring in the process.
func FindFreePort() (string, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return "", err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return "", err
	}
	defer listener.Close()

	return strconv.Itoa(listener.Addr().(*net.TCPAddr).Port), nil
}
