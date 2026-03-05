package config

// LoggerConfig logging configuration.
type LoggerConfig struct {
	// Global log level (trace, debug, info, warn, error, fatal, panic).
	Level string `yaml:"level" validate:"required,oneof=trace debug info warn error fatal panic"`

	// Time format (unix, unixms, unixmicro, rfc3339, rfc3339nano).
	TimeFormat string `yaml:"time_format" validate:"omitempty,oneof=unix unixms unixmicro rfc3339 rfc3339nano"`

	// Console output settings.
	Console ConsoleLogConfig `yaml:"console"`

	// List of file log handlers.
	Files []FileLogConfig `yaml:"files" validate:"omitempty,dive"`
}

// ConsoleLogConfig console logging configuration.
type ConsoleLogConfig struct {
	// Enable or disable console output.
	Enabled bool `yaml:"enabled"`

	// Minimum log level (uses global level if empty).
	Level string `yaml:"level" validate:"omitempty,oneof=trace debug info warn error fatal panic"`

	// Maximum log level (optional, for filtering).
	MaxLevel string `yaml:"max_level" validate:"omitempty,oneof=trace debug info warn error fatal panic"`

	// Use colored output.
	Colored bool `yaml:"colored"`

	// Format logs in a human-readable form (not JSON).
	Pretty bool `yaml:"pretty"`
}

// FileLogConfig file logging configuration.
type FileLogConfig struct {
	// Path to the log file.
	Path string `yaml:"path" validate:"required"`

	// Minimum log level (uses global level if empty).
	Level string `yaml:"level" validate:"omitempty,oneof=trace debug info warn error fatal panic"`

	// Maximum log level (optional, for filtering).
	MaxLevel string `yaml:"max_level" validate:"omitempty,oneof=trace debug info warn error fatal panic"`

	// Log rotation settings.
	Rotate RotateConfig `yaml:"rotate"`
}

// RotateConfig log file rotation configuration.
type RotateConfig struct {
	// Enable log rotation.
	Enabled bool `yaml:"enabled"`

	// Maximum file size in megabytes.
	MaxSize *int `yaml:"max_size" validate:"required_if=Enabled true,omitnil,gte=1"`

	// Maximum file age in days.
	MaxAge *int `yaml:"max_age" validate:"required_if=Enabled true,omitnil,gte=1"`

	// Maximum number of old log files to keep.
	MaxBackups *int `yaml:"max_backups" validate:"required_if=Enabled true,omitnil,gte=0"`

	// Compress old log files.
	Compress bool `yaml:"compress"`
}
