package rules

import "fmt"

func (r *Rule) Validate() error {
	if r.Type == "" || r.Payload == "" || r.Target == "" {
		return fmt.Errorf("invalid rule: %+v", r)
	}
	return nil
}
