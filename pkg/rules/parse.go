package rules

func ParseRuleType(ruleType string) RuleType {
	for k, v := range ruleTypeMap {
		if v == ruleType {
			return k
		}
	}
	return Unknown
}
