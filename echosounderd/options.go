package echosounderd

import "time"

// Options represents the server's options/configuration.
type Options struct {
	ListenAddress           string
	StatServerListenAddress string
	StatServerReadTimeout   time.Duration
	StatServerWriteTimeout  time.Duration
}
