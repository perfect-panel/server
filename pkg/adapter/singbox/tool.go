package singbox

import "encoding/json"

func mergeOptions(target map[string]any, options any) error {
	optionsJSON, err := json.Marshal(options)
	if err != nil {
		return err
	}
	return json.Unmarshal(optionsJSON, &target)
}
