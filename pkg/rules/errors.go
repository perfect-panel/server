package rules

import "errors"

var (
	ErrRuleTypeNotFound   = errors.New("rule type not found")
	ErrRuleTargetNotFound = errors.New("rule target not found")
)
