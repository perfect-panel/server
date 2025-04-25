package rules

import (
	"strings"
)

const noResolve = "no-resolve"

type Rule struct {
	Type    string
	Payload string
	Target  string
}

func NewRule(text, name string) *Rule {
	rule := trimArr(strings.Split(text, ","))
	var (
		payload string
		target  string
	)
	switch l := len(rule); {
	case l == 2:
		payload = rule[1]
		target = name
	case l == 3:
		payload = rule[1]
		target = rule[2]
	case l >= 4:
		payload = rule[1]
		target = rule[2]
	default:
		return nil
	}
	rule = trimArr(rule)
	return &Rule{
		Type:    rule[0],
		Payload: payload,
		Target:  target,
	}
}

func (r *Rule) String() string {
	text := r.Type + "," + r.Payload + "," + r.Target
	switch ParseRuleType(r.Type) {
	case IPCIDR, IPSet:
		return text + "," + noResolve
	default:
		return text
	}
}
