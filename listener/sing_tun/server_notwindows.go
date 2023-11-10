//go:build !windows

package sing_tun

import (
	tun "github.com/MerlinKodo/sing-tun"
)

func tunNew(options tun.Options) (tun.Tun, error) {
	return tun.New(options)
}
