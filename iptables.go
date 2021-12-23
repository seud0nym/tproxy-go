//go:build linux

package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"
)

// Holds the parameters from a parsed iptables/ip6tables rule
type TProxy struct {
	target, proto, family string
	listenIP, destIP      net.IP
	listenPort, destPort  int
}

// Parses output from iptables and ip6tables to find tproxy rules
// and returns arrays of the created listeners
func parseRules() ([]net.Listener, []*net.UDPConn) {
	tcpListeners := make([]net.Listener, 0)
	udpListeners := make([]*net.UDPConn, 0)

	cmd := "/usr/sbin/ip%stables"
	for _, v := range []string{"", "6"} {
		result, err := exec.Command(fmt.Sprintf(cmd, v), "-t", "mangle", "-S", "PREROUTING").Output()
		if err != nil {
			log.Fatal(err)
		}

		scanner := bufio.NewScanner(strings.NewReader(string(result)))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "!tproxy-go/") {
				tproxy := TProxy{}
				if v == "" {
					tproxy.family = "4"
				} else {
					tproxy.family = v
				}

				last := ""
				for _, token := range strings.Split(line, " ") {
					switch last {
					case "-j":
						tproxy.target = token
					case "-p":
						tproxy.proto = token
					case "--dport":
						tproxy.destPort, err = strconv.Atoi(token)
						if err != nil {
							log.Fatalf("FATAL: Failed to parse --dport from rule [%s]: %s", line, err)
						}
					case "--comment":
						tokens := strings.Split(strings.ReplaceAll(token, "\"", ""), "@")
						if len(tokens) == 2 {
							tproxy.destIP = net.ParseIP(tokens[1])
							if tproxy.destIP == nil {
								log.Fatalf("FATAL: Failed to parse destination IP from rule [%s]", line)
							}
						}
					case "--on-port":
						tproxy.listenPort, err = strconv.Atoi(token)
						if err != nil {
							log.Fatalf("FATAL: Failed to parse --on-port from rule [%s]: %s", line, err)
						}
					case "--on-ip":
						tproxy.listenIP = net.ParseIP(token)
						if tproxy.listenIP == nil {
							log.Fatalf("FATAL: Failed to parse --on-ip from rule [%s]", line)
						}
					}
					last = token
				}

				if strings.EqualFold(tproxy.target, "TPROXY") {
					switch tproxy.proto {
					case "tcp":
						tcpListeners = append(tcpListeners, makeTCPListener(tproxy))
					case "udp":
						udpListeners = append(udpListeners, makeUDPListener(tproxy))
					}
				}
			}
		}
	}

	return tcpListeners, udpListeners
}
