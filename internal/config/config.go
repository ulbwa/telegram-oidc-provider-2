package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"go.yaml.in/yaml/v3"
)

type Config struct {
	// Logging configuration.
	Logger LoggerConfig `yaml:"logger" validate:"required"`
}

var defaultConfig = Config{
	Logger: LoggerConfig{
		Level:      "info",
		TimeFormat: "rfc3339",
		Console: ConsoleLogConfig{
			Enabled: true,
			Colored: true,
			Pretty:  true,
		},
	},
}

var configValidator = validator.New()

func Read(path string) (*Config, error) {
	//nolint:gosec // G304: configuration path is explicit runtime input (CLI/entrypoint).
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	expandedData := os.ExpandEnv(string(data))

	c := defaultConfig
	if err := yaml.Unmarshal([]byte(expandedData), &c); err != nil {
		return nil, err
	}

	if err := configValidator.Struct(c); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &c, nil
}
