package server_test

import (
	"context"
	"errors"
	"net/http"
	"os/exec"
	"sync"
	"testing"
	"time"
)

const (
	serverBuildScript  = "./script/build-server"
	serverBuildTarget  = "./bin/server/server"
	serverBuildTimeout = 5 * time.Second
	serverTestsTimeout = 30 * time.Second

	serverPort        = "8888"
	serverURL         = "http://localhost:" + serverPort
	serverRunEndpoint = serverURL + "/run"
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

	cmd := exec.CommandContext(ctx, serverBuildScript)
	return cmd.Run()
}

func startServer() (shutdown func()) {
	ctx, cancel := context.WithTimeout(context.Background(), serverTestsTimeout)
	ok := false
	shutdown = func() {
		ok = true
		cancel()
		time.Sleep(1 * time.Second)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	cmd := exec.CommandContext(ctx, serverBuildTarget, "-port", serverPort)

	go func() {
		if err := cmd.Run(); err != nil && !ok {
			panic("cmd.Run: " + err.Error())
		}
	}()

	if err := pingServer(ctx, wg.Done); err != nil {
		panic("pingServer: " + err.Error())
	}

	wg.Wait()
	return shutdown
}

func pingServer(ctx context.Context, onConnect func()) error {
	const pollRate = 1 * time.Second

	ping := func() error {
		resp, err := http.Get(serverURL)
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
			return errors.New("could not reach server")
		default:
			if err := ping(); err != nil {
				time.Sleep(pollRate)
			} else {
				onConnect()
				return nil
			}
		}
	}
}
