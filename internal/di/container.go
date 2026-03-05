package di

import (
	"github.com/samber/do/v2"

	"github.com/ulbwa/telegram-oidc-provider/internal/config"
)

func NewContainer(cfg *config.Config) do.Injector {
	injector := do.New()

	// Config
	do.ProvideValue(injector, cfg)

	provideZerolog(injector)

	return injector
}
