package bot

import "context"

// TelegramBotProfileDTO is a service-layer contract DTO for Telegram bot profile data.
type TelegramBotProfileDTO struct {
	ID       int64
	Name     string
	Username *string
}

// TelegramProvider is a service-layer port for Telegram API access.
type TelegramProvider interface {
	FetchBotProfile(ctx context.Context, token string) (*TelegramBotProfileDTO, error)
}
