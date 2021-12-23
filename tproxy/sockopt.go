//go:build linux

package tproxy

import (
	"fmt"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

// Sets the listener socket options
func setListenSocketOptions(fileDescriptor int, network string, laddr net.Addr) *net.OpError {
	var err error

	if network == "tcp" || network == "tcp4" || network == "udp" || network == "udp4" {
		if err = syscall.SetsockoptInt(fileDescriptor, syscall.SOL_IP, syscall.IP_TRANSPARENT, 1); err != nil {
			return &net.OpError{Op: "listen", Net: network, Source: nil, Addr: laddr, Err: fmt.Errorf("set socket option: IP_TRANSPARENT: %s", err)}
		}
		if err = syscall.SetsockoptInt(fileDescriptor, syscall.SOL_IP, syscall.IP_RECVORIGDSTADDR, 1); err != nil {
			return &net.OpError{Op: "listen", Net: network, Source: nil, Addr: laddr, Err: fmt.Errorf("set socket option: IP_RECVORIGDSTADDR: %s", err)}
		}
	}

	if network == "tcp" || network == "tcp6" || network == "udp" || network == "udp6" {
		if err = syscall.SetsockoptInt(fileDescriptor, unix.SOL_IPV6, unix.IPV6_TRANSPARENT, 1); err != nil {
			return &net.OpError{Op: "listen", Net: network, Source: nil, Addr: laddr, Err: fmt.Errorf("set socket option: IPV6_TRANSPARENT: %s", err)}
		}
		if err = syscall.SetsockoptInt(fileDescriptor, unix.SOL_IPV6, unix.IPV6_RECVORIGDSTADDR, 1); err != nil {
			return &net.OpError{Op: "listen", Net: network, Source: nil, Addr: laddr, Err: fmt.Errorf("set socket option: IPV6_RECVORIGDSTADDR: %s", err)}
		}
	}

	return nil
}
