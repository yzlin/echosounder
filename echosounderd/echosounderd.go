package echosounderd

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	"echosounder/internal/util"
	"echosounder/internal/version"
)

// EchoSounderd represents the echosounder TCP server daemon.
type EchoSounderd struct {
	sync.RWMutex
	opts               *Options
	tcpServerListener  net.Listener
	statServer         *http.Server
	statServerListener net.Listener
	waitGroup          util.WaitGroupWrapper
}

// New initializes an EchoSounderd object with corresponding options.
func New(opts *Options) *EchoSounderd {
	e := &EchoSounderd{
		opts: opts,
	}

	log.SetOutput(os.Stdout)

	log.Printf(version.String("echosounderd"))

	return e
}

// Main is the orion service's main function.
func (e *EchoSounderd) Main() {
	tcpListener, err := net.Listen("tcp", e.opts.ListenAddress)
	if err != nil {
		log.Fatalf("TCP server failed to listen on (%s) - %s", e.opts.ListenAddress, err)
	}
	log.Printf("TCP server on %s (pid %d)", tcpListener.Addr().(*net.TCPAddr), syscall.Getpid())

	statServerListener, err := net.Listen("tcp", e.opts.StatServerListenAddress)
	if err != nil {
		log.Fatalf("Stat server failed to listen on %s - %s", e.opts.StatServerListenAddress, err)
	}
	log.Printf("Stat server on %s (pid %d)", statServerListener.Addr().(*net.TCPAddr), syscall.Getpid())

	e.Lock()
	e.statServerListener = statServerListener
	e.statServer = &http.Server{
		ReadTimeout:  e.opts.StatServerReadTimeout,
		WriteTimeout: e.opts.StatServerWriteTimeout,
	}
	e.Unlock()

	e.tcpServerListener = tcpListener

	// run tcp server
	e.waitGroup.Wrap(func() {
		var tempDelay time.Duration
		for {
			conn, err := e.tcpServerListener.Accept()
			if err != nil {
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					if tempDelay == 0 {
						tempDelay = 5 * time.Millisecond
					} else {
						tempDelay *= 2
					}
					if max := 1 * time.Second; tempDelay > max {
						tempDelay = max
					}
					log.Printf("tcp: Accept error: %v; retrying in %v", err, tempDelay)
					time.Sleep(tempDelay)
					continue
				}
			}
			tempDelay = 0

			go e.handleTCPRequest(conn)
		}
	})

	// run stat server
	if e.statServer != nil {
		e.waitGroup.Wrap(func() {
			if err := e.statServer.Serve(e.statServerListener); err != nil {
				log.Fatal(err)
			}
		})
	}
}

func (e *EchoSounderd) Exit() {
	if e.statServer != nil {
		// Shutdown gracefully, but wait no longer than 10 seconds
		// TODO: configurable
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := e.statServer.Shutdown(ctx); err != nil {
			log.Printf("Failed to shut down stat server gracefully: %s", err)
		}
	}

	e.waitGroup.Wait()

	log.Print("Server stopped")
}

func (e *EchoSounderd) handleTCPRequest(conn net.Conn) {
	defer func() {
		log.Printf("[%s] disconnected", conn.RemoteAddr())
		conn.Close()
	}()

	// TODO: configurable read deadline
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	reader := bufio.NewReader(conn)

cmdLoop:
	for {
		resp := ""

		bytes, _, err := reader.ReadLine()
		if err != nil {
			return
		}

		cmd := string(bytes)
		log.Printf("[%s] received: %s", conn.RemoteAddr(), cmd)

		cmd = strings.TrimSpace(cmd)
		switch cmd {
		case "quit":
			break cmdLoop

		default:
			resp = fmt.Sprintf("Unknown command: %s", cmd)
		}

		if resp != "" {
			conn.Write([]byte(resp + "\n"))
		}
	}
}
