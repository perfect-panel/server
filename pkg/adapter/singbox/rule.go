package singbox

import (
	"strconv"

	"github.com/perfect-panel/ppanel-server/pkg/logger"
	"github.com/perfect-panel/ppanel-server/pkg/rules"
)

type Rule struct {
	Outbound      string   `json:"outbound,omitempty"`
	ClashMode     string   `json:"clash_mode,omitempty"`
	RuleSet       []string `json:"rule_set,omitempty"`
	Domain        []string `json:"domain,omitempty"`
	DomainSuffix  []string `json:"domain_suffix,omitempty"`
	DomainKeyword []string `json:"domain_keyword,omitempty"`
	DomainRegex   []string `json:"domain_regex,omitempty"`
	GeoIP         []string `json:"geoip,omitempty"`
	IPCIDR        []string `json:"ip_cidr,omitempty"`
	IPIsPrivate   bool     `json:"ip_is_private,omitempty"`
	SourceIPCIDR  []string `json:"source_ip_cidr,omitempty"`
	ProcessName   []string `json:"process_name,omitempty"`
	ProcessPath   []string `json:"process_path,omitempty"`
	SourcePort    []uint16 `json:"source_port,omitempty"`
	Protocol      []string `json:"protocol,omitempty"`
	Port          []uint16 `json:"port,omitempty"`
	Action        string   `json:"action,omitempty"`
	Inbound       []string `json:"inbound,omitempty"`
	Rules         []Rule   `json:"rules,omitempty"`
	Type          string   `json:"type,omitempty"`
	Mode          string   `json:"mode,omitempty"`
}

type RuleSet struct {
	Tag            string `json:"tag,omitempty"`
	Type           string `json:"type,omitempty"`
	Format         string `json:"format,omitempty"`
	URL            string `json:"url,omitempty"`
	DownloadDetour string `json:"download_detour,omitempty"`
}

func adapterToSingboxRule(texts []string) []Rule {
	var rulesList []Rule
	for _, rule := range texts {
		r := rules.NewRule(rule, "")
		if r == nil {
			continue
		}
		rulesList = addRuleToItem(rulesList, r.Target, *r)
	}
	return rulesList
}

func addRuleToItem(group []Rule, outbound string, rule rules.Rule) []Rule {
	for i := range group {
		if group[i].Outbound == outbound {
			switch rules.ParseRuleType(rule.Type) {
			case rules.Domain:
				group[i].Domain = append(group[i].Domain, rule.Payload)
				return group
			case rules.DomainSuffix:
				group[i].DomainSuffix = append(group[i].DomainSuffix, rule.Payload)
				return group
			case rules.DomainKeyword:
				group[i].DomainKeyword = append(group[i].DomainKeyword, rule.Payload)
				return group
			case rules.IPCIDR:
				group[i].IPCIDR = append(group[i].IPCIDR, rule.Payload)
				return group
			case rules.SrcIPCIDR:
				group[i].SourceIPCIDR = append(group[i].SourceIPCIDR, rule.Payload)
				return group
			case rules.SrcPort:
				port, err := strconv.ParseUint(rule.Payload, 10, 16)
				if err != nil {
					logger.Errorf("[adapterToSingboxRule] failed to parse port %s to uint16", rule.Payload)
					return group
				}
				group[i].SourcePort = append(group[i].SourcePort, uint16(port))
				return group
			case rules.GEOIP:
				group[i].GeoIP = append(group[i].GeoIP, rule.Payload)
				return group
			case rules.Process:
				group[i].ProcessName = append(group[i].ProcessName, rule.Payload)
				return group
			case rules.ProcessPath:
				group[i].ProcessPath = append(group[i].ProcessPath, rule.Payload)
				return group
			default:
				logger.Errorf("[adapterToSingboxRule] unknown rule type %s", rule.Type)
				return group
			}
		}
	}
	newRule := Rule{
		Outbound: outbound,
	}

	switch rules.ParseRuleType(rule.Type) {
	case rules.Domain:
		newRule.Domain = []string{rule.Payload}
	case rules.DomainSuffix:
		newRule.DomainSuffix = []string{rule.Payload}
	case rules.DomainKeyword:
		newRule.DomainKeyword = []string{rule.Payload}
	case rules.IPCIDR:
		newRule.IPCIDR = []string{rule.Payload}
	case rules.SrcIPCIDR:
		newRule.SourceIPCIDR = []string{rule.Payload}
	case rules.SrcPort:
		port, err := strconv.ParseUint(rule.Payload, 10, 16)
		if err != nil {
			logger.Errorf("[adapterToSingboxRule] failed to parse port %s to uint16", rule.Payload)
			return group
		}
		newRule.SourcePort = []uint16{uint16(port)}
	case rules.GEOIP:
		newRule.GeoIP = []string{rule.Payload}
	case rules.Process:
		newRule.ProcessName = []string{rule.Payload}
	case rules.ProcessPath:
		newRule.ProcessPath = []string{rule.Payload}
	default:
		logger.Errorf("[adapterToSingboxRule] unknown rule type %s", rule.Type)
		return group
	}
	group = append(group, newRule)
	return group
}
