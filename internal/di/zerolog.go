package di

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/do/v2"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/ulbwa/go-backend-template/internal/config"
)

func provideZerolog(injector do.Injector) {
	do.Provide(injector, func(i do.Injector) (*zerolog.Logger, error) {
		cfg, err := do.Invoke[*config.Config](i)
		if err != nil {
			// Fallback to default configuration
			logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
			return &logger, nil
		}

		globalLevel := parseLevel(cfg.Logger.Level, zerolog.InfoLevel)
		zerolog.SetGlobalLevel(globalLevel)

		configureTimeFormat(cfg.Logger.TimeFormat)

		var writers []io.Writer

		// Console output
		if cfg.Logger.Console.Enabled {
			consoleWriter := createConsoleWriter(cfg.Logger.Console)

			if cfg.Logger.Console.Level != "" || cfg.Logger.Console.MaxLevel != "" {
				consoleLevel := parseLevel(cfg.Logger.Console.Level, globalLevel)
				maxConsoleLevel := parseLevel(cfg.Logger.Console.MaxLevel, zerolog.PanicLevel)
				writers = append(writers, &levelRangeWriter{
					writer:   consoleWriter,
					minLevel: consoleLevel,
					maxLevel: maxConsoleLevel,
				})
			} else {
				writers = append(writers, consoleWriter)
			}
		}

		// File outputs
		for _, fileCfg := range cfg.Logger.Files {
			if fileCfg.Path == "" {
				continue
			}

			fileWriter := createFileWriter(fileCfg)

			fileLevel := parseLevel(fileCfg.Level, globalLevel)
			maxLevel := parseLevel(fileCfg.MaxLevel, zerolog.PanicLevel)

			if fileCfg.Level != "" || fileCfg.MaxLevel != "" {
				writers = append(writers, &levelRangeWriter{
					writer:   fileWriter,
					minLevel: fileLevel,
					maxLevel: maxLevel,
				})
			} else {
				writers = append(writers, fileWriter)
			}
		}

		// Default to stdout if no writers configured
		if len(writers) == 0 {
			writers = append(writers, os.Stdout)
		}

		var output io.Writer
		if len(writers) == 1 {
			output = writers[0]
		} else {
			output = zerolog.MultiLevelWriter(writers...)
		}

		logger := zerolog.New(output).With().Timestamp().Logger()
		logger = logger.Level(globalLevel)

		return &logger, nil
	})
}

func createConsoleWriter(cfg config.ConsoleLogConfig) io.Writer {
	if cfg.Pretty {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05",
			NoColor:    !cfg.Colored,
		}
		return consoleWriter
	}
	return os.Stdout
}

// createFileWriter creates a file writer with optional rotation support.
func createFileWriter(cfg config.FileLogConfig) io.Writer {
	// Ensure log directory exists
	dir := filepath.Dir(cfg.Path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return os.Stderr
		}
	}

	if cfg.Rotate.Enabled {
		maxSize := 100
		if cfg.Rotate.MaxSize != nil {
			maxSize = *cfg.Rotate.MaxSize
		}

		maxAge := 0
		if cfg.Rotate.MaxAge != nil {
			maxAge = *cfg.Rotate.MaxAge
		}

		maxBackups := 0
		if cfg.Rotate.MaxBackups != nil {
			maxBackups = *cfg.Rotate.MaxBackups
		}

		return &lumberjack.Logger{
			Filename:   cfg.Path,
			MaxSize:    maxSize,
			MaxAge:     maxAge,
			MaxBackups: maxBackups,
			Compress:   cfg.Rotate.Compress,
		}
	}

	// Simple file without rotation
	file, err := os.OpenFile(cfg.Path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		// Fallback to stderr on error
		return os.Stderr
	}
	return file
}

func parseLevel(levelStr string, defaultLevel zerolog.Level) zerolog.Level {
	if levelStr == "" {
		return defaultLevel
	}

	switch strings.ToLower(levelStr) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	case "disabled":
		return zerolog.Disabled
	default:
		return defaultLevel
	}
}

func configureTimeFormat(format string) {
	switch strings.ToLower(format) {
	case "unix":
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	case "unixms":
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	case "unixmicro":
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMicro
	case "rfc3339":
		zerolog.TimeFieldFormat = time.RFC3339
	case "rfc3339nano":
		zerolog.TimeFieldFormat = time.RFC3339Nano
	default:
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	}
}

// levelRangeWriter filters log messages by level range.
type levelRangeWriter struct {
	writer   io.Writer
	minLevel zerolog.Level
	maxLevel zerolog.Level
}

func (lw *levelRangeWriter) Write(p []byte) (n int, err error) {
	return lw.writer.Write(p)
}

func (lw *levelRangeWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	if level >= lw.minLevel && level <= lw.maxLevel {
		return lw.writer.Write(p)
	}
	return len(p), nil
}
