package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/benchttp/engine/httpclient"
)

var (
	// stdout is used to send messages to the parent process.
	// Use stdout to send info and debug level messages.
	stdout = log.New(os.Stdout, "", 0)
	// stderr is used to send messages to the parent process.
	// Use stderr to send error level messages or to notify
	// any action which are followed by os.Exit.
	stderr = log.New(os.Stderr, "", log.LstdFlags)
)

func main() {
	if err := run(); err != nil {
		stderr.Fatal(err)
	}
}

func run() error {
	useAutoPort := flag.Bool("auto-port", true, "automatically find and use an available port")
	flag.Parse()

	port, err := resolvePort(*useAutoPort)
	if err != nil {
		return err
	}

	// Notify server is ready.
	stdout.Println(readySignalString(port))

	return httpclient.ListenAndServe(port)
}

func resolvePort(autoPort bool) (string, error) {
	if autoPort {
		return httpclient.FindFreePort()
	}
	return devPort()
}

func devPort() (string, error) {
	err := godotenv.Load("./.env.development")
	if err != nil {
		return "", fmt.Errorf("could not load .env file: %w", err)
	}
	return os.Getenv("SERVER_PORT"), nil
}

const readySignal = "READY"

func readySignalString(port string) string {
	return fmt.Sprintf("%s http://localhost:%s", readySignal, port)
}
