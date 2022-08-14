package metrics

import (
	"fmt"
)

func StatusCodeDistributionComputeError(statusCode int) error {
	return fmt.Errorf("%d is not a valid HTTP status code", statusCode)
}
