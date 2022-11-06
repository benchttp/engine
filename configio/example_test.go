package configio_test

import (
	"fmt"
	"time"

	"github.com/benchttp/sdk/benchttp"
	"github.com/benchttp/sdk/configio"
)

var jsonConfig = []byte(
	`{"request": {"method": "GET", "url": "https://example.com"}}`,
)

var yamlConfig = []byte(
	`{request: {method: PUT}, runner: {requests: 42}}`,
)

func ExampleBuilder() {
	runner := benchttp.Runner{Requests: -1, Concurrency: 3}

	b := configio.Builder{}
	_ = b.WriteJSON(jsonConfig)
	_ = b.WriteYAML(yamlConfig)
	b.SetInterval(100 * time.Millisecond)

	b.Mutate(&runner)

	// Output:
	// PUT
	// https://example.com
	// 42
	// 3
	// 100ms
	fmt.Println(runner.Request.Method)
	fmt.Println(runner.Request.URL)
	fmt.Println(runner.Requests)
	fmt.Println(runner.Concurrency)
	fmt.Println(runner.Interval)
}
