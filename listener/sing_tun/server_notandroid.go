//go:build !android

package sing_tun

import (
	tun "github.com/MerlinKodo/sing-tun"
)

func (l *Listener) buildAndroidRules(tunOptions *tun.Options) error {
	return nil
}
func (l *Listener) openAndroidHotspot(tunOptions tun.Options) {}
