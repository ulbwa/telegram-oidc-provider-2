package telegram

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	gotgbot "github.com/PaulSonOfLars/gotgbot/v2"

	errs "github.com/ulbwa/telegram-oidc-provider/internal/errors"
	servicebot "github.com/ulbwa/telegram-oidc-provider/internal/service/bot"
	"github.com/ulbwa/telegram-oidc-provider/pkg/utils"
)

type GoTgBotClientAdapter struct {
	gotgbot.BotClient
}

func NewGoTgBotClientAdapter(botClient gotgbot.BotClient) *GoTgBotClientAdapter {
	if botClient == nil {
		botClient = &gotgbot.BaseBotClient{Client: http.Client{}}
	}

	return &GoTgBotClientAdapter{BotClient: botClient}
}

func (a *GoTgBotClientAdapter) FetchBotProfile(ctx context.Context, token string) (*servicebot.TelegramBotProfileDTO, error) {
	if strings.TrimSpace(token) == "" {
		return nil, errs.ErrTelegramBotTokenInvalid
	}

	if a.BotClient == nil {
		a.BotClient = &gotgbot.BaseBotClient{Client: http.Client{}}
	}

	bot := &gotgbot.Bot{Token: token, BotClient: a.BotClient}

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

	var username *string
	if strings.TrimSpace(botUser.Username) != "" && botUser.Username != "<missing>" {
		username = utils.Ptr(botUser.Username)
	}

	return &servicebot.TelegramBotProfileDTO{
		ID:       botUser.Id,
		Name:     botUser.FirstName,
		Username: username,
	}, nil
}

func isUnauthorizedError(err error) bool {
	var telegramError *gotgbot.TelegramError
	if errors.As(err, &telegramError) {
		return strings.Contains(strings.ToLower(telegramError.Description), "unauthorized")
	}

	return strings.Contains(strings.ToLower(err.Error()), "unauthorized")
}
