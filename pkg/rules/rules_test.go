package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var text = `
DOMAIN,example.com
DOMAIN-SUFFIX,google.com,DIRECT
DOMAIN-KEYWORD,amazon,REJECT
IP-CIDR,192.168.0.0/16
`

func TestNewRule(t *testing.T) {
	var rs []string
	// parse validate rules
	ruleArr := strings.Split(text, "\n")
	if len(ruleArr) == 0 {
		t.Error("rules is empty")
	}
	ruleArr = trimArr(ruleArr)
	for _, s := range ruleArr {
		r := NewRule(s, "Test")
		if r == nil {
			t.Errorf("[CreateRuleGroup] rule %s is nil, len: %d", s, len(s))
			continue
		}
		if err := r.Validate(); err != nil {
			t.Errorf("[CreateRuleGroup] rule %s is invalid: %v", s, err)
			continue
		}
		rs = append(rs, r.String())
	}

	expected := []string{
		"DOMAIN,example.com,Test",
		"DOMAIN-SUFFIX,google.com,DIRECT",
		"DOMAIN-KEYWORD,amazon,REJECT",
		"IP-CIDR,192.168.0.0/16,Test,no-resolve",
	}

	for i, r := range rs {
		if r != expected[i] {
			t.Errorf("expected %s, got %s", expected[i], r)
		}
	}
	// Check if the rules are sorted
	assert.Equal(t, len(rs), len(expected))
}
