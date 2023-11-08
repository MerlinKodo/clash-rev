package rules

import (
	"github.com/MerlinKodo/clash-rev/component/ipset"
	C "github.com/MerlinKodo/clash-rev/constant"
	"github.com/MerlinKodo/clash-rev/log"
)

// Implements C.Rule
var _ C.Rule = (*IPSet)(nil)

type IPSet struct {
	name        string
	adapter     string
	noResolveIP bool
}

func (f *IPSet) RuleType() C.RuleType {
	return C.IPSet
}

func (f *IPSet) Match(metadata *C.Metadata) bool {
	exist, err := ipset.Test(f.name, metadata.DstIP)
	if err != nil {
		log.Warnln("check ipset '%s' failed: %s", f.name, err.Error())
		return false
	}
	return exist
}

func (f *IPSet) Adapter() string {
	return f.adapter
}

func (f *IPSet) Payload() string {
	return f.name
}

func (f *IPSet) ShouldResolveIP() bool {
	return !f.noResolveIP
}

func (f *IPSet) ShouldFindProcess() bool {
	return false
}

func NewIPSet(name string, adapter string, noResolveIP bool) (*IPSet, error) {
	if err := ipset.Verify(name); err != nil {
		return nil, err
	}

	return &IPSet{
		name:        name,
		adapter:     adapter,
		noResolveIP: noResolveIP,
	}, nil
}
