package server_test

import (
	"context"
	"net/http"
	"os/exec"
	"testing"
	"time"
)

const (
	serverBuildScript  = "./script/build"
	serverBuildTarget  = "./bin/server/server"
	serverBuildTimeout = 5 * time.Second
	serverTestsTimeout = 30 * time.Second

	serverPort       = "8888"
	serverAddr       = "localhost:" + serverPort
	serverDummyToken = "6db67fafc4f5bf965a5a" //nolint:gosec // dummy token for development
)

func TestMain(m *testing.M) {
	shutdown := setupServer()
	defer shutdown()

	m.Run()
}

func setupServer() (clean func()) {
	if err := buildServer(); err != nil {
		panic(err)
	}
	return startServer()
}

func buildServer() error {
	ctx, cancel := context.WithTimeout(context.Background(), serverBuildTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, serverBuildScript, "server")
	return cmd.Run()
}

func startServer() (shutdown func()) {
	ctx, cancel := context.WithTimeout(context.Background(), serverTestsTimeout)
	cmd := exec.CommandContext(ctx, serverBuildTarget, "-port", serverPort)

	stopped := true
	go func() {
		if err := cmd.Run(); err != nil && stopped {
			panic("cmd.Run: " + err.Error())
		}
	}()

	waitServerReady(ctx)

	return func() {
		stopped = false
		cancel()
		time.Sleep(1 * time.Second)
	}
}

func waitServerReady(ctx context.Context) {
	const pollRate = 1 * time.Second

	ping := func() error {
		resp, err := http.Get("http://" + serverAddr)
		if err != nil {
			return err
		}
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			panic("timeout: could not reach server")
		default:
			if err := ping(); err != nil {
				time.Sleep(pollRate)
			} else {
				// wg.Done()
				return
			}
		}
	}
}
