package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/cameliot/alpaca"
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

}
