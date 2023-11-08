package executor

import (
	"fmt"
	"os"
	"sync"

	"github.com/MerlinKodo/clash-rev/adapter"
	"github.com/MerlinKodo/clash-rev/adapter/outboundgroup"
	"github.com/MerlinKodo/clash-rev/component/auth"
	"github.com/MerlinKodo/clash-rev/component/dialer"
	"github.com/MerlinKodo/clash-rev/component/iface"
	"github.com/MerlinKodo/clash-rev/component/profile"
	"github.com/MerlinKodo/clash-rev/component/profile/cachefile"
	"github.com/MerlinKodo/clash-rev/component/resolver"
	"github.com/MerlinKodo/clash-rev/component/trie"
	"github.com/MerlinKodo/clash-rev/config"
	C "github.com/MerlinKodo/clash-rev/constant"
	"github.com/MerlinKodo/clash-rev/constant/provider"
	"github.com/MerlinKodo/clash-rev/dns"
	"github.com/MerlinKodo/clash-rev/listener"
	authStore "github.com/MerlinKodo/clash-rev/listener/auth"
	"github.com/MerlinKodo/clash-rev/log"
	"github.com/MerlinKodo/clash-rev/tunnel"
)

var mux sync.Mutex

func readConfig(path string) ([]byte, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("configuration file %s is empty", path)
	}

	return data, err
}

// Parse config with default config path
func Parse() (*config.Config, error) {
	return ParseWithPath(C.Path.Config())
}

// ParseWithPath parse config with custom config path
func ParseWithPath(path string) (*config.Config, error) {
	buf, err := readConfig(path)
	if err != nil {
		return nil, err
	}

	return ParseWithBytes(buf)
}

// ParseWithBytes config with buffer
func ParseWithBytes(buf []byte) (*config.Config, error) {
	return config.Parse(buf)
}

// ApplyConfig dispatch configure to all parts
func ApplyConfig(cfg *config.Config, force bool) {
	mux.Lock()
	defer mux.Unlock()

	updateUsers(cfg.Users)
	updateProxies(cfg.Proxies, cfg.Providers)
	updateRules(cfg.Rules)
	updateHosts(cfg.Hosts)
	updateProfile(cfg)
	updateGeneral(cfg.General, force)
	updateInbounds(cfg.Inbounds, force)
	updateDNS(cfg.DNS)
	updateExperimental(cfg)
	updateTunnels(cfg.Tunnels)
}

func GetGeneral() *config.General {
	ports := listener.GetPorts()
	authenticator := []string{}
	if auth := authStore.Authenticator(); auth != nil {
		authenticator = auth.Users()
	}

	general := &config.General{
		LegacyInbound: config.LegacyInbound{
			Port:        ports.Port,
			SocksPort:   ports.SocksPort,
			RedirPort:   ports.RedirPort,
			TProxyPort:  ports.TProxyPort,
			MixedPort:   ports.MixedPort,
			AllowLan:    listener.AllowLan(),
			BindAddress: listener.BindAddress(),
		},
		Authentication: authenticator,
		Mode:           tunnel.Mode(),
		LogLevel:       log.Level(),
		IPv6:           !resolver.DisableIPv6,
	}

	return general
}

func updateExperimental(c *config.Config) {
	tunnel.UDPFallbackMatch.Store(c.Experimental.UDPFallbackMatch)
}

func updateDNS(c *config.DNS) {
	if !c.Enable {
		resolver.DefaultResolver = nil
		resolver.DefaultHostMapper = nil
		dns.ReCreateServer("", nil, nil)
		return
	}

	cfg := dns.Config{
		Main:         c.NameServer,
		Fallback:     c.Fallback,
		IPv6:         c.IPv6,
		EnhancedMode: c.EnhancedMode,
		Pool:         c.FakeIPRange,
		Hosts:        c.Hosts,
		FallbackFilter: dns.FallbackFilter{
			GeoIP:     c.FallbackFilter.GeoIP,
			GeoIPCode: c.FallbackFilter.GeoIPCode,
			IPCIDR:    c.FallbackFilter.IPCIDR,
			Domain:    c.FallbackFilter.Domain,
		},
		Default:       c.DefaultNameserver,
		Policy:        c.NameServerPolicy,
		SearchDomains: c.SearchDomains,
	}

	r := dns.NewResolver(cfg)
	m := dns.NewEnhancer(cfg)

	// reuse cache of old host mapper
	if old := resolver.DefaultHostMapper; old != nil {
		m.PatchFrom(old.(*dns.ResolverEnhancer))
	}

	resolver.DefaultResolver = r
	resolver.DefaultHostMapper = m

	dns.ReCreateServer(c.Listen, r, m)
}

func updateHosts(tree *trie.DomainTrie) {
	resolver.DefaultHosts = tree
}

func updateProxies(proxies map[string]C.Proxy, providers map[string]provider.ProxyProvider) {
	tunnel.UpdateProxies(proxies, providers)
}

func updateRules(rules []C.Rule) {
	tunnel.UpdateRules(rules)
}

func updateTunnels(tunnels []config.Tunnel) {
	listener.PatchTunnel(tunnels, tunnel.TCPIn(), tunnel.UDPIn())
}

func updateInbounds(inbounds []C.Inbound, force bool) {
	if !force {
		return
	}
	tcpIn := tunnel.TCPIn()
	udpIn := tunnel.UDPIn()

	listener.ReCreateListeners(inbounds, tcpIn, udpIn)
}

func updateGeneral(general *config.General, force bool) {
	log.SetLevel(general.LogLevel)
	tunnel.SetMode(general.Mode)
	resolver.DisableIPv6 = !general.IPv6

	dialer.DefaultInterface.Store(general.Interface)
	dialer.DefaultRoutingMark.Store(int32(general.RoutingMark))

	iface.FlushCache()

	if !force {
		return
	}

	allowLan := general.AllowLan
	listener.SetAllowLan(allowLan)

	bindAddress := general.BindAddress
	listener.SetBindAddress(bindAddress)

	ports := listener.Ports{
		Port:       general.Port,
		SocksPort:  general.SocksPort,
		RedirPort:  general.RedirPort,
		TProxyPort: general.TProxyPort,
		MixedPort:  general.MixedPort,
	}
	listener.ReCreatePortsListeners(ports, tunnel.TCPIn(), tunnel.UDPIn())
}

func updateUsers(users []auth.AuthUser) {
	authenticator := auth.NewAuthenticator(users)
	authStore.SetAuthenticator(authenticator)
	if authenticator != nil {
		log.Infoln("Authentication of local server updated")
	}
}

func updateProfile(cfg *config.Config) {
	profileCfg := cfg.Profile

	profile.StoreSelected.Store(profileCfg.StoreSelected)
	if profileCfg.StoreSelected {
		patchSelectGroup(cfg.Proxies)
	}
}

func patchSelectGroup(proxies map[string]C.Proxy) {
	mapping := cachefile.Cache().SelectedMap()
	if mapping == nil {
		return
	}

	for name, proxy := range proxies {
		outbound, ok := proxy.(*adapter.Proxy)
		if !ok {
			continue
		}

		selector, ok := outbound.ProxyAdapter.(*outboundgroup.Selector)
		if !ok {
			continue
		}

		selected, exist := mapping[name]
		if !exist {
			continue
		}

		selector.Set(selected)
	}
}
