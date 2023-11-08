package process

import (
	"errors"
	"net/netip"
)

var (
	ErrInvalidNetwork     = errors.New("invalid network")
	ErrPlatformNotSupport = errors.New("not support on this platform")
	ErrNotFound           = errors.New("process not found")
)

const (
	TCP = "tcp"
	UDP = "udp"
)

func FindProcessPath(network string, from netip.AddrPort, to netip.AddrPort) (string, error) {
	return findProcessPath(network, from, to)
}
