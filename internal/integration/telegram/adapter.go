package telegram

import (
	"context"
	"errors"
	"strings"
	"time"

	gotgbot "github.com/PaulSonOfLars/gotgbot/v2"

	errs "github.com/ulbwa/telegram-oidc-provider/internal/errors"
	servicebot "github.com/ulbwa/telegram-oidc-provider/internal/service/bot"
)

type goTgBotClientAdapter struct {
	botClient gotgbot.BotClient
}

func NewGoTgBotClientAdapter(botClient gotgbot.BotClient) (servicebot.TelegramProvider, error) {
	if botClient == nil {
		return nil, errors.New("telegram bot client is nil")
	}

	return &goTgBotClientAdapter{botClient: botClient}, nil
}

func (a *goTgBotClientAdapter) FetchBotProfile(ctx context.Context, token string) (*servicebot.TelegramBotProfileDTO, error) {
	if strings.TrimSpace(token) == "" {
		return nil, errs.ErrTelegramBotTokenRequired
	}
	if strings.TrimSpace(token) != token {
		return nil, errs.ErrTelegramBotTokenInvalid
	}

	bot := &gotgbot.Bot{Token: token, BotClient: a.botClient}

	botUser, err := bot.GetMeWithContext(ctx, &gotgbot.GetMeOpts{RequestOpts: &gotgbot.RequestOpts{Timeout: 10 * time.Second}})
	if err != nil {
		if isUnauthorizedError(err) || errors.Is(err, gotgbot.ErrInvalidTokenFormat) {
			return nil, errs.ErrTelegramBotTokenInvalid
		}

		return nil, err
	}

	if !botUser.IsBot || botUser.Id <= 0 || strings.TrimSpace(botUser.FirstName) == "" {
		return nil, errs.ErrTelegramAPIResponse
	}

	if strings.TrimSpace(botUser.Username) == "" {
		return nil, errs.ErrTelegramBotUsernameRequired
	}

	return &servicebot.TelegramBotProfileDTO{
		ID:       botUser.Id,
		Name:     botUser.FirstName,
		Username: botUser.Username,
	}, nil
}

func isUnauthorizedError(err error) bool {
	var telegramError *gotgbot.TelegramError
	if errors.As(err, &telegramError) {
		return strings.Contains(strings.ToLower(telegramError.Description), "unauthorized")
	}

	return strings.Contains(strings.ToLower(err.Error()), "unauthorized")
}
