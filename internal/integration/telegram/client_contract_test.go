package telegram

import (
	"testing"

	servicebot "github.com/ulbwa/telegram-oidc-provider/internal/service/bot"
)

func TestAdapterImplementsProvider(t *testing.T) {
	t.Parallel()

	var _ servicebot.TelegramProvider = (*GoTgBotClientAdapter)(nil)
}
