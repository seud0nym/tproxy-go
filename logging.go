//go:build linux

package main

import (
	"io"
	"log"
	"log/syslog"
	"os"
)

func logging(console bool) {
	logger, err := syslog.New(syslog.LOG_DAEMON|syslog.LOG_NOTICE, "tproxy-go")
	if err != nil {
		log.SetOutput(os.Stderr)
		log.Printf("ERROR: Writing to stderr because failed to open syslog (%s)", err)
	} else {
		log.SetFlags(0)
		if console {
			mw := io.MultiWriter(os.Stdout, logger)
			log.SetOutput(mw)
		} else {
			log.SetOutput(logger)
		}
	}
}
