package rules

type RuleType int

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
	Unknown
)

var ruleTypeMap = map[RuleType]string{
	Domain:        "DOMAIN",
	DomainSuffix:  "DOMAIN-SUFFIX",
	DomainKeyword: "DOMAIN-KEYWORD",
	GEOIP:         "GEOIP",
	IPCIDR:        "IP-CIDR",
	SrcIPCIDR:     "SRC-IP-CIDR",
	SrcPort:       "SRC-PORT",
	DstPort:       "DST-PORT",
	InboundPort:   "INBOUND-PORT",
	Process:       "PROCESS-NAME",
	ProcessPath:   "PROCESS-PATH",
	IPSet:         "IPSET",
	MATCH:         "MATCH",
	Unknown:       "UNKNOWN",
}

func (rt RuleType) String() string {
	if str, ok := ruleTypeMap[rt]; ok {
		return str
	}
	return "UNKNOWN"
}
