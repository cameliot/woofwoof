package main

import (
	"encoding/json"
	"net/http"

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

/*
Create HTTP server for serving watcher reports
*/
func serveHttp(httpConfig HttpConfig, watchers []*ServiceWatcher) {

	// Create HTTP Handler Functions
	// -- Report Handler
	http.HandleFunc(
		"/v1/services",
		func(res http.ResponseWriter, req *http.Request) {

			// Make report
			report := map[string]ServiceReport{}

			for _, watcher := range watchers {
				report[watcher.config.Handle] = watcher.Report()
			}

			// Write response
			json.NewEncoder(res).Encode(report)
		})

	// -- Welcome Handler
	http.HandleFunc(
		"/",
		func(res http.ResponseWriter, req *http.Request) {
			fmt.Fprintf(res, `
            <html>
                <h1>WoofWoof v.%s</h1>
                <p>The goodboy companion for your alpaca and llama herd.</p>

                <h2>Find your service reports here:</h2>
                <ul>
                    <li><a href='/v1/services'>/v1/services</a></li>
                </ul>
            </html>
        `, version)

		})

	// Start HTTP server
	log.Println("Serving reports via HTTP on:", httpConfig.Listen)
	http.ListenAndServe(httpConfig.Listen, nil)
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

	actions, dispatch := alpaca.DialMqtt(
		config.Broker.Uri,
		config.AlpacaRoutes(),
	)

	// Periodically trigger pings
	go func() {
		for {
			dispatch(Ping("*"))
			time.Sleep(1 * time.Minute)
		}
	}()

	// Create service watchers
	watchers := []*ServiceWatcher{}
	for _, serviceConfig := range config.Services {
		watchers = append(watchers, NewServiceWatcher(serviceConfig, dispatch))
	}

	// Start HTTP
	go serveHttp(config.Http, watchers)

	for action := range actions {
		for _, watcher := range watchers {
			watcher.Handle(action)
		}
	}
}
