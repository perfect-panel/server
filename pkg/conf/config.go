package conf

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

func MustLoad(file string, v any) {
	if err := Load(file, v); err != nil {
		log.Fatalf("error: config file %s, %s", file, err.Error())
	}
}

func Load(file string, v any) error {
	setDefaults(v)
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	// Unmarshal the YAML content directly into the target structure
	if err := yaml.Unmarshal(content, v); err != nil {
		return err
	}
	return nil
}
