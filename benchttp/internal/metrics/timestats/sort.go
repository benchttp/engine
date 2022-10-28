package timestats

import (
	"time"
)

// byFastest implements sort.Interface for []time.Duration.
type byFastest []time.Duration

func (a byFastest) Len() int {
	return len(a)
}

func (a byFastest) Less(i, j int) bool {
	return a[i] < a[j]
}

func (a byFastest) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
