//go:build !linux

package dialer

import (
	"net"
	"sync"

	"github.com/MerlinKodo/clash-rev/log"
)

var printMarkWarn = sync.OnceFunc(func() {
	log.Warnln("Routing mark on socket is not supported on current platform")
})

func bindMarkToDialer(mark int, dialer *net.Dialer, _ string, _ net.IP) {
	printMarkWarn()
}

func bindMarkToListenConfig(mark int, lc *net.ListenConfig, _, address string) {
	printMarkWarn()
}
