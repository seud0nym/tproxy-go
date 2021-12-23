Linux Transparent Proxy
=======================

A Golang implementation of a Linux transparent proxy, for intercepting and redirecting packets using iptables/ip6tables TPROXY target (https://www.kernel.org/doc/Documentation/networking/tproxy.txt).

This package is based on original work that can be found in: https://github.com/LiamHaworth/go-tproxy. It has been adapted to use a single target rather than pass-thru to the original destination, and added IPv6 support.