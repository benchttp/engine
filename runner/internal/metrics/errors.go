package metrics

import (
	"fmt"
)

func StatusCodeDistributionComputeError(statusCode int) error {
	return fmt.Errorf("%d is not a valid HTTP status code", statusCode)
}

func RequestEventsDistributionComputeErr(eventName string) error {
	return fmt.Errorf("%s is not a valid event name", eventName)
}
