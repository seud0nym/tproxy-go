//go:build linux

package main

import (
	"log"
	"net"
	"time"

	"github.com/seud0nym/tproxy-go/tproxy"
)

// Creates the listener
func makeUDPListener(proxy TProxy) *net.UDPConn {
	udpListen := &net.UDPAddr{IP: proxy.listenIP, Port: proxy.listenPort}
	udpDest := &net.UDPAddr{IP: proxy.destIP, Port: proxy.destPort}
	udpListener, err := tproxy.ListenUDP("udp"+proxy.family, udpListen)
	if err != nil {
		log.Fatalf("FATAL: Error error while binding IPv%s UDP listener on %s: %s", proxy.family, udpListen, err)
	}

	debug("Listening on %s for proxied UDP requests to port %d to be handled by %s", udpListen, proxy.destPort, udpDest)
	go waitForUDPConn("udp"+proxy.family, udpListener, udpDest.String())
	return udpListener
}

// Accept UDP connections and hand them off to handleUDPConn
func waitForUDPConn(network string, listener *net.UDPConn, dest string) {
	for {
		buff := make([]byte, 1024)
		n, srcAddr, dstAddr, err := tproxy.ReadFromUDP(listener, buff)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				log.Printf("ERROR: Temporary error while reading data from %s: %s", listener.LocalAddr(), netErr)
			}

			log.Fatalf("FATAL: Unrecoverable error while reading data from %s: %s", listener.LocalAddr(), err)
			return
		}

		go handleUDPConn(network, buff[:n], srcAddr, dstAddr, dest)
	}
}

// Opens a connection to the proxy destination and write the received data to the remote host, then wait a few seconds for any response data
func handleUDPConn(network string, data []byte, srcAddr, dstAddr *net.UDPAddr, proxy string) {
	debug("Intercepted UDP connection from %s with destination of %s", srcAddr, dstAddr)

	localConn, err := tproxy.DialUDP(network, dstAddr, srcAddr)
	if err != nil {
		log.Printf("ERROR: Failed to connect to original UDP source [%s] to handle intercepted UDP connection to %s: %s", srcAddr, dstAddr, err)
		return
	}
	defer localConn.Close()

	remoteConn, err := net.Dial(network, proxy)
	if err != nil {
		log.Printf("ERROR: Failed to connect to remote host [%s] to handle intercepted UDP connection %s->%s: %s", proxy, srcAddr, dstAddr, err)
		return
	}
	defer remoteConn.Close()

	bytesWritten, err := remoteConn.Write(data)
	if err != nil {
		log.Printf("ERROR: Encountered error while writing to remote host [%s] to handle intercepted UDP connection %s->%s: %s", remoteConn.RemoteAddr(), srcAddr, dstAddr, err)
		return
	} else if bytesWritten < len(data) {
		log.Printf("ERROR: Not all bytes [%d < %d] in buffer written to remote host [%s] when handling intercepted UDP connection %s->%s", bytesWritten, len(data), remoteConn.RemoteAddr(), srcAddr, dstAddr)
		return
	}

	data = make([]byte, 1024)
	remoteConn.SetReadDeadline(time.Now().Add(2 * time.Second)) // Add deadline to ensure it doesn't block forever
	bytesRead, err := remoteConn.Read(data)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return
		}

		log.Printf("ERROR: Encountered error while reading from remote host [%s] when handling intercepted UDP connection %s->%s: %s", remoteConn.RemoteAddr(), srcAddr, dstAddr, err)
		return
	}

	bytesWritten, err = localConn.Write(data)
	if err != nil {
		log.Printf("ERROR: Encountered error while writing to local host [%s] when handling intercepted UDP connection %s->%s: %s", localConn.RemoteAddr(), srcAddr, dstAddr, err)
		return
	} else if bytesWritten < bytesRead {
		log.Printf("ERROR: Not all bytes [%d < %d] in buffer written to local host [%s] when handling intercepted UDP connection %s->%s", bytesWritten, len(data), remoteConn.RemoteAddr(), srcAddr, dstAddr)
		return
	}
}
