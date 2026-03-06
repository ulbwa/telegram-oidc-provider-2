package telegram

import (
	"errors"
	"testing"

	gotgbot "github.com/PaulSonOfLars/gotgbot/v2"

	servicebot "github.com/ulbwa/telegram-oidc-provider/internal/service/bot"
)

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

	if !errors.Is(err, ErrTelegramBotClientNil) {
		t.Fatalf("expected errors.Is(err, ErrTelegramBotClientNil) to be true, got %v", err)
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
