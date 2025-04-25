package singbox

import (
	"fmt"
	"testing"
)

func TestAdapterToSingboxRule(t *testing.T) {
	rules := []string{
		"DOMAIN,example.com,DIRECT",
		"DOMAIN-SUFFIX,google.com,智能线路",
	}
	result := adapterToSingboxRule(rules)
	fmt.Printf("TestAdapterToSingboxRule: result: %+v\n", result)
}
