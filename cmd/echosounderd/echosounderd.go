package main

import (
	"flag"
	"log"
	"time"

	"echosounder/echosounderd"
	"echosounder/internal/service"
)

type program struct {
	echosounderd *echosounderd.EchoSounderd
	opts         *echosounderd.Options
}

func (p *program) Init() error {
	return nil
}

func (p *program) Start() error {
	service := echosounderd.New(p.opts)
	service.Main()
	p.echosounderd = service
	return nil
}

func (p *program) Stop() error {
	if p.echosounderd != nil {
		p.echosounderd.Exit()
	}
	return nil
}

var tcpServerListen string
var statServerListen string

func init() {
	const (
		defaultTCPServerListen  = ":10023"
		defaultStatServerListen = ":10080"
	)

	flag.StringVar(&tcpServerListen, "listen", defaultTCPServerListen, "Listen address")
	flag.StringVar(&statServerListen, "stat-listen", defaultStatServerListen, "Listen address for stat server")
}

func main() {
	flag.Parse()

	opts := &echosounderd.Options{
		ListenAddress:           tcpServerListen,
		StatServerListenAddress: statServerListen,
		StatServerReadTimeout:   10 * time.Second,
		StatServerWriteTimeout:  10 * time.Second,
	}

	prg := &program{
		opts: opts,
	}
	if err := service.Run(prg); err != nil {
		log.Fatalf("Failed to run echosounderd: %s", err)
	}
}
