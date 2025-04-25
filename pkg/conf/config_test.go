package conf

import "testing"

type Server struct {
	Host string `yaml:"Host" default:"localhost"`
	Port int    `yaml:"Port" default:"8080"`
}

type Config struct {
	Server Server `yaml:"Server"`
}

func TestConfigLoad(t *testing.T) {
	var c Config
	MustLoad("./config_test.yaml", &c)
	t.Logf("config: %+v", c)
}
