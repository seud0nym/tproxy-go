//go:build linux

package main

import (
	"io"
	"log"
	"net"
	"sync"

	"github.com/seud0nym/tproxy-go/tproxy"
)

// Creates the listener
func makeTCPListener(proxy TProxy) net.Listener {
	tcpListen := &net.TCPAddr{IP: proxy.listenIP, Port: proxy.listenPort}
	tcpDest := &net.TCPAddr{IP: proxy.destIP, Port: proxy.destPort}
	tcpListener, err := tproxy.ListenTCP("tcp"+proxy.family, tcpListen)
	if err != nil {
		log.Fatalf("FATAL: Error while binding IPv%s TCP listener on %s: %s", proxy.family, tcpListen, err)
	}

	debug("Listening on %s for proxied TCP requests to port %d to be handled by %s", tcpListen, proxy.destPort, tcpDest)
	go waitForTCPConn("tcp"+proxy.family, tcpListener, tcpDest.String())
	return tcpListener
}

// Accept TCP connections and hand them off to handleTCPConn
func waitForTCPConn(network string, listener net.Listener, dest string) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("FATAL: Unrecoverable error while accepting connection on %s: %s", listener.Addr(), err)
			return
		}

		go handleTCPConn(network, conn, dest)
	}
}

// Opens a TCP connection to the proxy destination and setup two routines to stream data between the connections
func handleTCPConn(network string, conn net.Conn, dest string) {
	debug("Intercepted TCP connection from %s with destination of %s", conn.RemoteAddr().String(), conn.LocalAddr().String())
	defer conn.Close()

	remoteConn, err := net.Dial(network, dest)
	if err != nil {
		log.Printf("ERROR: Failed to connect to remote host [%s] to handle intercepted TCP connection %s->%s: %s", dest, conn.RemoteAddr().String(), conn.LocalAddr().String(), err)
		return
	}
	defer remoteConn.Close()

	var streamWait sync.WaitGroup
	streamWait.Add(2)

	streamConn := func(dst io.Writer, src io.Reader) {
		io.Copy(dst, src)
		streamWait.Done()
	}

	go streamConn(remoteConn, conn)
	go streamConn(conn, remoteConn)

	streamWait.Wait()
}
