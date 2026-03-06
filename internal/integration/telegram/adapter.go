package telegram

import (
	"context"
	"errors"
	"strings"
	"time"

	gotgbot "github.com/PaulSonOfLars/gotgbot/v2"

	"github.com/ulbwa/telegram-oidc-provider/internal/service/telegram"
)

type goTgBotClientAdapter struct {
	botClient gotgbot.BotClient
}

var ErrTelegramBotClientNil = errors.New("telegram bot client is nil")

func NewGoTgBotClientAdapter(botClient gotgbot.BotClient) (telegram.Provider, error) {
	if botClient == nil {
		return nil, ErrTelegramBotClientNil
	}

	return &goTgBotClientAdapter{botClient: botClient}, nil
}

func (a *goTgBotClientAdapter) FetchBot(ctx context.Context, token string) (*telegram.Bot, error) {
	if strings.TrimSpace(token) == "" {
		return nil, telegram.ErrBotTokenRequired
	}
	if strings.TrimSpace(token) != token {
		return nil, telegram.ErrBotTokenInvalid
	}

	bot := &gotgbot.Bot{Token: token, BotClient: a.botClient}

	botUser, err := bot.GetMeWithContext(ctx, &gotgbot.GetMeOpts{RequestOpts: &gotgbot.RequestOpts{Timeout: 10 * time.Second}})
	if err != nil {
		if isUnauthorizedError(err) || errors.Is(err, gotgbot.ErrInvalidTokenFormat) {
			return nil, telegram.ErrBotTokenInvalid
		}

		return nil, err
	}

	if !botUser.IsBot || botUser.Id <= 0 || strings.TrimSpace(botUser.FirstName) == "" {
		return nil, telegram.ErrBadResponse
	}

	if strings.TrimSpace(botUser.Username) == "" {
		return nil, telegram.ErrBadResponse
	}

	return &telegram.Bot{
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
