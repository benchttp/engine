package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/benchttp/runner/request"
)

const (
	DefaultConcurrency = 1
	DefaultRequests    = 0 // Use duration as exit condition if omitted.
	DefaultDuration    = 60 * time.Second
	DefaultTimeout     = 10 * time.Second
)

var (
	url         string
	concurrency int           // Number of connections to run concurrently.
	requests    int           // Number of requests to run, use duration as exit condition if omitted.
	duration    time.Duration // Duration of test, in seconds.
	timeout     time.Duration // Timeout for each http request, in seconds.
)

func parseArgs() {
	url = os.Args[len(os.Args)-1]

	flag.IntVar(&concurrency, "c", DefaultConcurrency, "Number of connections to run concurrently")
	flag.IntVar(&requests, "r", DefaultRequests, "Number of requests to run, use duration as exit condition if omitted")
	flag.DurationVar(&duration, "d", DefaultDuration, "Duration of test, in seconds")
	flag.DurationVar(&timeout, "t", DefaultTimeout, "Timeout for each http request, in seconds")
	flag.Parse()

	fmt.Printf("Testing url: %s\n", url)
	fmt.Printf("concurrency: %d\n", concurrency)
	if requests > 0 {
		fmt.Printf("requests: %d\n", requests)
	}
	fmt.Printf("duration: %s\n", duration)

	println()
}

func main() {
	parseArgs()

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	rec := request.Do(ctx, requests, concurrency, url, timeout)

	fmt.Println("total:", len(rec))
}
