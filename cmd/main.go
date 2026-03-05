package main

import (
	"context"
	"flag"
	"time"

	"github.com/rs/zerolog"
	"github.com/samber/do/v2"

	"github.com/ulbwa/go-backend-template/internal/config"
	"github.com/ulbwa/go-backend-template/internal/di"
)

func loadConfig(path string) *config.Config {
	cfg, err := config.Read(path)
	if err != nil {
		panic(err)
	}
	return cfg
}

func shutdown(injector do.Injector) {
	logger := do.MustInvoke[*zerolog.Logger](injector)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
	defer cancel()
	shutdownReport := injector.ShutdownWithContext(ctx)
	if len(shutdownReport.Errors) > 0 {
		for service, err := range shutdownReport.Errors {
			// Log the error using a fallback logger since the main logger might be part of the shutdown
			logger.Error().Err(err).Interface("service", service).Msg("Error during shutdown")
		}
	}
}

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg := loadConfig(*configPath)

	// Initialize the DI container with the configuration
	injector := di.NewContainer(cfg)

	logger := do.MustInvoke[*zerolog.Logger](injector)
	logger.Info().Msg("Hello, World!")

	shutdown(injector)
}
