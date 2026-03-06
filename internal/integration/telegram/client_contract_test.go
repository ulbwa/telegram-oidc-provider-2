package telegram

import (
	"testing"

	gotgbot "github.com/PaulSonOfLars/gotgbot/v2"

	servicebot "github.com/ulbwa/telegram-oidc-provider/internal/service/bot"
)

const errBotClientNilMessage = "telegram bot client is nil"

func TestAdapterImplementsProvider(t *testing.T) {
	t.Parallel()

	var _ servicebot.TelegramProvider = (*goTgBotClientAdapter)(nil)
}

func TestNewGoTgBotClientAdapterNilClient(t *testing.T) {
	t.Parallel()

	adapter, err := NewGoTgBotClientAdapter(nil)
	if err == nil {
		t.Fatalf("expected non-nil error")
	}

	if err.Error() != errBotClientNilMessage {
		t.Fatalf("expected error %q, got %q", errBotClientNilMessage, err.Error())
	}

	if adapter != nil {
		t.Fatalf("expected nil adapter when client is nil")
	}
}

func TestNewGoTgBotClientAdapterSuccess(t *testing.T) {
	t.Parallel()

	adapter, err := NewGoTgBotClientAdapter(&gotgbot.BaseBotClient{})
	if err == nil {
		if adapter == nil {
			t.Fatalf("expected non-nil adapter")
		}
		return
	}

	t.Fatalf("expected no error, got %v", err)
}
