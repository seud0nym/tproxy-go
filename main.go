//go:build linux

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"time"
)

var (
	appVersion = time.Now().Format("2006.01.02") + "-dev"
	verbose    = true
)

func main() {
	var console bool
	var version bool

	flag.BoolVar(&console, "l", false, "Log messages to stdout as well as syslog")
	flag.BoolVar(&verbose, "v", false, "Verbose logging")
	flag.BoolVar(&version, "V", false, "Show version and exit")
	flag.Parse()

	if version {
		os.Exit(0)
	}

	logging(console)

	log.Printf("tproxy-go v%s https://github.com/seud0nym/tproxy-go\n", appVersion)

	tcpListeners, udpListeners := parseRules()
	if len(tcpListeners) > 0 || len(udpListeners) > 0 {
		for _, listener := range tcpListeners {
			defer listener.Close()
		}
		for _, listener := range udpListeners {
			defer listener.Close()
		}

		interruptListener := make(chan os.Signal)
		signal.Notify(interruptListener, os.Interrupt)
		<-interruptListener
		debug("Shutting down")
	} else {
		debug("No TPROXY rules found in iptables/ip6tables mangle table PREROUTING chain with a comment prefixed by \"!tproxy-go@\"")
	}

	os.Exit(0)
}

// Writes debug messages
func debug(format string, v ...interface{}) {
	if verbose {
		log.Printf("DEBUG: "+format, v...)
	}
}
