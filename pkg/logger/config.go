package logger

// A LogConf is a logging config.
type LogConf struct {
	// ServiceName represents the service name.
	ServiceName string `yaml:"ServiceName" default:"PPanel"`
	// Mode represents the logging mode, default is `console`.
	// console: log to console.
	// file: log to file.
	// volume: used in k8s, prepend the hostname to the log file name.
	Mode string `yaml:"Mode" default:"file"`
	// Encoding represents the encoding type, default is `json`.
	// json: json encoding.
	// plain: plain text encoding, typically used in development.
	Encoding string `yaml:"Encoding" default:"json"`
	// TimeFormat represents the time format, default is `2006-01-02T15:04:05.000Z07:00`.
	TimeFormat string `yaml:"TimeFormat" default:"2006-01-02 15:04:05.000"`
	// Path represents the log file path, default is `logs`.
	Path string `yaml:"Path" default:"logs"`
	// Level represents the log level, default is `info`.
	Level string `yaml:"Level" default:"info"`
	// MaxContentLength represents the max content bytes, default is no limit.
	MaxContentLength uint32 `yaml:"MaxContentLength" default:"0"`
	// Compress represents whether to compress the log file, default is `false`.
	Compress bool `yaml:"Compress" default:"false"`
	// Stat represents whether to log statistics, default is `true`.
	Stat bool `yaml:"Stat" default:"true"`
	// KeepDays represents how many days the log files will be kept. Default to keep all files.
	// Only take effect when Mode is `file` or `volume`, both work when Rotation is `daily` or `size`.
	KeepDays int `yaml:"KeepDays" default:"0"`
	// StackCooldownMillis represents the cooldown time for stack logging, default is 100ms.
	StackCooldownMillis int `yaml:"StackCooldownMillis" default:"100"`
	// MaxBackups represents how many backup log files will be kept. 0 means all files will be kept forever.
	// Only take effect when RotationRuleType is `size`.
	// Even though `MaxBackups` sets 0, log files will still be removed
	// if the `KeepDays` limitation is reached.
	MaxBackups int `yaml:"MaxBackups" default:"0"`
	// MaxSize represents how much space the writing log file takes up. 0 means no limit. The unit is `MB`.
	// Only take effect when RotationRuleType is `size`
	MaxSize int `yaml:"MaxSize" default:"0"`
	// Rotation represents the type of log rotation rule. Default is `daily`.
	// daily: daily rotation.
	// size: size limited rotation.
	Rotation string `yaml:"Rotation" default:"daily"`
	// FileTimeFormat represents the time format for file name, default is `2006-01-02T15:04:05.000Z07:00`.
	FileTimeFormat string `yaml:"FileTimeFormat" default:"2006-01-02T15:04:05.000Z07:00"`
}
