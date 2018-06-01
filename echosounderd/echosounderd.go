package echosounderd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
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
	stat               stat
	dummyRateLimit     <-chan time.Time
}

// New initializes an EchoSounderd object with corresponding options.
func New(opts *Options) *EchoSounderd {
	e := &EchoSounderd{
		opts: opts,
	}

	e.dummyRateLimit = time.Tick(time.Second / 30) // 30 req/s

	log.SetOutput(os.Stdout)

	log.Printf(version.String("echosounderd"))

	return e
}

// Main is the echosounder service's main function.
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
		http.HandleFunc("/stat", func(w http.ResponseWriter, r *http.Request) {
			totalReq := atomic.LoadInt64(&e.stat.totalRequest)
			processedReq := atomic.LoadInt64(&e.stat.processedRequest)

			resp, _ := json.Marshal(&struct {
				CurrentConnection int64 `json:"current_connection"`
				TotalRequest      int64 `json:"total_request"`
				ProcessedRequest  int64 `json:"processed_request"`
				WaitingRequest    int64 `json:"waiting_request"`
			}{
				CurrentConnection: atomic.LoadInt64(&e.stat.currentConnection),
				TotalRequest:      totalReq,
				ProcessedRequest:  processedReq,
				WaitingRequest:    totalReq - processedReq,
			})

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Write(resp)
			w.Write([]byte("\n"))
		})
		e.waitGroup.Wrap(func() {
			if err := e.statServer.Serve(e.statServerListener); err != nil {
				log.Fatal(err)
			}
		})
	}
}

// Exit shuts the service down
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
	atomic.AddInt64(&e.stat.currentConnection, 1)
	log.Printf("[%s] connected", conn.RemoteAddr())
	defer func() {
		atomic.AddInt64(&e.stat.currentConnection, -1)
		log.Printf("[%s] disconnected", conn.RemoteAddr())
		conn.Close()
	}()

	reader := bufio.NewReader(conn)

cmdLoop:
	for {
		// TODO: configurable read deadline
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		// Prompt
		conn.Write([]byte(">>> "))

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
			conn.Write([]byte("bye~\n"))
			break cmdLoop

		case "stat":
			totalReq := atomic.LoadInt64(&e.stat.totalRequest)
			processedReq := atomic.LoadInt64(&e.stat.processedRequest)
			resp = fmt.Sprintf(`
	Current connection: %d
	Total request: %d
	Processed request: %d
	Waiting request: %d`,
				atomic.LoadInt64(&e.stat.currentConnection),
				totalReq,
				processedReq,
				totalReq-processedReq,
			)

		case "dummy":
			atomic.AddInt64(&e.stat.totalRequest, 1)
			<-e.dummyRateLimit
			p, err := requestDummyAPI()
			if err != nil {
				log.Printf("[%s] Failed to request dummy API: %s", conn.RemoteAddr(), err)
				resp = fmt.Sprintf("Failed to request dummy API: %s", err)
			} else {
				resp = fmt.Sprintf("\tTitle: %s\n\tBody: %s", p.Title, p.Body)
			}
			atomic.AddInt64(&e.stat.processedRequest, 1)

		default:
			resp = fmt.Sprintf("Unknown command: %s", cmd)
		}

		if resp != "" {
			conn.Write([]byte(resp + "\n"))
		}
	}
}
