package service

import (
	"os"
	"os/signal"
	"syscall"
)

// Service wraps simple 3-step Init(), Start(), Stop() with corresponding signal-responsible
// which basic services require.
type Service interface {
	// Init initializes the daemon. It's called before Start().
	Init() error

	// Start is called after init to start the daemon.
	Start() error

	// Stop is called when receiving syscall.SIGINT, syscall.SIGTERM
	Stop() error
}

// Run runs the service with customizable signals.
func Run(svc Service, sig ...os.Signal) error {
	if err := svc.Init(); err != nil {
		return err
	}

	if err := svc.Start(); err != nil {
		return err
	}

	if len(sig) == 0 {
		sig = []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, sig...)
	<-signalChan

	return svc.Stop()
}
