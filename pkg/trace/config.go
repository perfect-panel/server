package trace

// TraceName represents the tracing name.
const TraceName = "ppanel"

// A Config is an opentelemetry config.
type Config struct {
	Name     string  `yaml:"Name"`
	Endpoint string  `yaml:"Endpoint"`
	Sampler  float64 `yaml:"Sampler" default:"1.0"`
	Batcher  string  `yaml:"Batcher" default:"jaeger"`
	// OtlpHeaders represents the headers for OTLP gRPC or HTTP transport.
	// For example:
	//  uptrace-dsn: 'http://project2_secret_token@localhost:14317/2'
	OtlpHeaders map[string]string `yaml:"OtlpHeaders"`
	// OtlpHttpPath represents the path for OTLP HTTP transport.
	// For example
	// /v1/traces
	OtlpHttpPath string `yaml:"OtlpHttpPath"`
	// OtlpHttpSecure represents the scheme to use for OTLP HTTP transport.
	OtlpHttpSecure bool `yaml:"OtlpHttpSecure"`
	// Disabled indicates whether StartAgent starts the agent.
	Disabled bool `yaml:"Disabled"`
}
