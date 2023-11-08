package constant

const (
	RuleConfigDomain        RuleConfig = "DOMAIN"
	RuleConfigDomainSuffix  RuleConfig = "DOMAIN-SUFFIX"
	RuleConfigDomainKeyword RuleConfig = "DOMAIN-KEYWORD"
	RuleConfigGeoIP         RuleConfig = "GEOIP"
	RuleConfigIPCIDR        RuleConfig = "IP-CIDR"
	RuleConfigIPCIDR6       RuleConfig = "IP-CIDR6"
	RuleConfigSrcIPCIDR     RuleConfig = "SRC-IP-CIDR"
	RuleConfigSrcPort       RuleConfig = "SRC-PORT"
	RuleConfigDstPort       RuleConfig = "DST-PORT"
	RuleConfigInboundPort   RuleConfig = "INBOUND-PORT"
	RuleConfigProcessName   RuleConfig = "PROCESS-NAME"
	RuleConfigProcessPath   RuleConfig = "PROCESS-PATH"
	RuleConfigIPSet         RuleConfig = "IPSET"
	RuleConfigRuleSet       RuleConfig = "RULE-SET"
	RuleConfigScript        RuleConfig = "SCRIPT"
	RuleConfigMatch         RuleConfig = "MATCH"
)

// Rule Config Type String represents a rule type in configuration files.
type RuleConfig string

// Rule Type
const (
	Domain RuleType = iota
	DomainSuffix
	DomainKeyword
	GEOIP
	IPCIDR
	SrcIPCIDR
	SrcPort
	DstPort
	InboundPort
	Process
	ProcessPath
	IPSet
	MATCH
)

type RuleType int

func (rt RuleType) String() string {
	switch rt {
	case Domain:
		return "Domain"
	case DomainSuffix:
		return "DomainSuffix"
	case DomainKeyword:
		return "DomainKeyword"
	case GEOIP:
		return "GeoIP"
	case IPCIDR:
		return "IPCIDR"
	case SrcIPCIDR:
		return "SrcIPCIDR"
	case SrcPort:
		return "SrcPort"
	case DstPort:
		return "DstPort"
	case InboundPort:
		return "InboundPort"
	case Process:
		return "Process"
	case ProcessPath:
		return "ProcessPath"
	case IPSet:
		return "IPSet"
	case MATCH:
		return "Match"
	default:
		return "Unknown"
	}
}

type Rule interface {
	RuleType() RuleType
	Match(metadata *Metadata) bool
	Adapter() string
	Payload() string
	ShouldResolveIP() bool
	ShouldFindProcess() bool
}
