//go:build !linux && !darwin && !windows

package dialer

import (
	"net"
	"strconv"
)

func bindIfaceToDialer(ifaceName string, dialer *net.Dialer, network string, destination net.IP) error {
	if !destination.IsGlobalUnicast() {
		return nil
	}

	local := uint64(0)
	if dialer.LocalAddr != nil {
		_, port, err := net.SplitHostPort(dialer.LocalAddr.String())
		if err == nil {
			local, _ = strconv.ParseUint(port, 10, 16)
		}
	}

	addr, err := lookupLocalAddr(ifaceName, network, destination, int(local))
	if err != nil {
		return err
	}

	dialer.LocalAddr = addr

	return nil
}

func bindIfaceToListenConfig(ifaceName string, _ *net.ListenConfig, network, address string) (string, error) {
	_, port, err := net.SplitHostPort(address)
	if err != nil {
		port = "0"
	}

	local, _ := strconv.ParseUint(port, 10, 16)

	addr, err := lookupLocalAddr(ifaceName, network, nil, int(local))
	if err != nil {
		return "", err
	}

	return addr.String(), nil
}
