package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cameliot/alpaca"
)

var version = "unknown"

/*
 Show some usage help
*/
func usage() {
	fmt.Fprintf(os.Stderr, "usage: woofwoo /path/to/config.conf\n")
	flag.PrintDefaults()
	os.Exit(-1)
}

func main() {
	// Parse cli flags
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		usage()
	}

	log.Println("Reading config from:", args[0])

	// Load config from file
	config := LoadConfig(args[0])

	fmt.Println(config)

	actions, dispatch := alpaca.DialMqtt(
		config.Broker.Uri,
		config.AlpacaRoutes(),
	)

	// Periodically trigger pings
	go func() {
		dispatch(Ping("*"))
		time.Sleep(1 * time.Minute)
	}()

	// Create service watchers
	watchers := []*ServiceWatcher{}
	for _, serviceConfig := range config.Services {
		watchers = append(watchers, NewServiceWatcher(serviceConfig))
	}

	for action := range actions {
		fmt.Println(action)
		for _, watcher := range watchers {
			watcher.Handle(action)
		}
	}
}
