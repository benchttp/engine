package signals

import (
	"os"
	"os/signal"
	"syscall"
)

// ListenOSInterrupt listens for OS interrupt signals and calls callback
// on receive. It should be called in a separate goroutine from the main
// program as it blocks the execution until a signal is received.
func ListenOSInterrupt(callback func()) {
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-sigC
	callback()
}
