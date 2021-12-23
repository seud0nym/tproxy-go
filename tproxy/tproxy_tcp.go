//go:build linux

package tproxy

import (
	"fmt"
	"net"
)

// ListenTCP will construct a new TCP listener socket with the Linux IP_TRANSPARENT
// option et on the underlying socket
func ListenTCP(network string, laddr *net.TCPAddr) (net.Listener, error) {
	listener, err := net.ListenTCP(network, laddr)
	if err != nil {
		return nil, err
	}

	fileDescriptorSource, err := listener.File()
	if err != nil {
		return nil, &net.OpError{Op: "listen", Net: network, Source: nil, Addr: laddr, Err: fmt.Errorf("get file descriptor: %s", err)}
	}
	defer fileDescriptorSource.Close()

	opErr := setListenSocketOptions(int(fileDescriptorSource.Fd()), network, laddr)
	if opErr != nil {
		return nil, err
	}

	return &Listener{listener}, nil
}

// Listener describes a TCP Listener with the Linux IP_TRANSPARENT option defined
// on the listening socket
type Listener struct {
	base net.Listener
}

// Accept waits for and returns the next connection to the listener.
// This command wraps the AcceptTProxy method of the Listener
func (listener *Listener) Accept() (net.Conn, error) {
	return listener.AcceptTProxy()
}

// AcceptTProxy will accept a TCP connection and wrap it to a TProxy connection
// to provide TProxy functionality
func (listener *Listener) AcceptTProxy() (*net.TCPConn, error) {
	tcpConn, err := listener.base.(*net.TCPListener).AcceptTCP()
	if err != nil {
		return nil, err
	}

	return tcpConn, nil
}

// Addr returns the network address the listener is accepting connections from
func (listener *Listener) Addr() net.Addr {
	return listener.base.Addr()
}

// Close will close the listener from accepting any more connections.
// Any blocked connections will unblock and close
func (listener *Listener) Close() error {
	return listener.base.Close()
}
