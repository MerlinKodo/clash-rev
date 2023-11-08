package rules

import (
	"fmt"

	C "github.com/MerlinKodo/clash-rev/constant"
)

func ParseRule(tp, payload, target string, params []string) (C.Rule, error) {
	var (
		parseErr error
		parsed   C.Rule
	)

	ruleConfigType := C.RuleConfig(tp)

	switch ruleConfigType {
	case C.RuleConfigDomain:
		parsed = NewDomain(payload, target)
	case C.RuleConfigDomainSuffix:
		parsed = NewDomainSuffix(payload, target)
	case C.RuleConfigDomainKeyword:
		parsed = NewDomainKeyword(payload, target)
	case C.RuleConfigGeoIP:
		noResolve := HasNoResolve(params)
		parsed = NewGEOIP(payload, target, noResolve)
	case C.RuleConfigIPCIDR, C.RuleConfigIPCIDR6:
		noResolve := HasNoResolve(params)
		parsed, parseErr = NewIPCIDR(payload, target, WithIPCIDRNoResolve(noResolve))
	case C.RuleConfigSrcIPCIDR:
		parsed, parseErr = NewIPCIDR(payload, target, WithIPCIDRSourceIP(true), WithIPCIDRNoResolve(true))
	case C.RuleConfigSrcPort:
		parsed, parseErr = NewPort(payload, target, PortTypeSrc)
	case C.RuleConfigDstPort:
		parsed, parseErr = NewPort(payload, target, PortTypeDest)
	case C.RuleConfigInboundPort:
		parsed, parseErr = NewPort(payload, target, PortTypeInbound)
	case C.RuleConfigProcessName:
		parsed, parseErr = NewProcess(payload, target, true)
	case C.RuleConfigProcessPath:
		parsed, parseErr = NewProcess(payload, target, false)
	case C.RuleConfigIPSet:
		noResolve := HasNoResolve(params)
		parsed, parseErr = NewIPSet(payload, target, noResolve)
	case C.RuleConfigMatch:
		parsed = NewMatch(target)
	case C.RuleConfigRuleSet, C.RuleConfigScript:
		parseErr = fmt.Errorf("unsupported rule type %s", tp)
	default:
		parseErr = fmt.Errorf("unsupported rule type %s", tp)
	}

	return parsed, parseErr
}
