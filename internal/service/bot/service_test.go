package bot

import (
	"context"
	"errors"
	"testing"

	errdefs "github.com/ulbwa/telegram-oidc-provider/internal/errors"
	"github.com/ulbwa/telegram-oidc-provider/internal/model"
)

type repositoryMock struct {
	storedBot *model.Bot
	savedBot  *model.Bot
	saveErr   error
	getErr    error
}

func (m *repositoryMock) GetByID(ctx context.Context, id int64) (*model.Bot, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}

	if m.storedBot == nil {
		return nil, errdefs.ErrNotFound
	}

	if m.storedBot.ID != id {
		return nil, errdefs.ErrNotFound
	}

	return m.storedBot, nil
}

func (m *repositoryMock) Save(ctx context.Context, bot *model.Bot) error {
	m.savedBot = bot
	if m.saveErr != nil {
		return m.saveErr
	}

	m.storedBot = bot

	return nil
}

type telegramAdapterMock struct {
	profile *TelegramBotProfileDTO
	err     error
}

type transactionManagerMock struct {
	called bool
	err    error
}

func (m *transactionManagerMock) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	m.called = true
	if m.err != nil {
		return m.err
	}

	return fn(ctx)
}

func (m *telegramAdapterMock) FetchBotProfile(ctx context.Context, token string) (*TelegramBotProfileDTO, error) {
	if m.err != nil {
		return nil, m.err
	}

	return &TelegramBotProfileDTO{
		ID:       m.profile.ID,
		Name:     m.profile.Name,
		Username: m.profile.Username,
	}, nil
}

func TestServiceSyncByTokenCreate(t *testing.T) {
	t.Parallel()

	repository := &repositoryMock{}
	adapter := &telegramAdapterMock{
		profile: &TelegramBotProfileDTO{ID: 777, Name: "OIDC Bot", Username: "oidc_login_bot"},
	}
	txManager := &transactionManagerMock{}

	service := NewService(repository, adapter, txManager)
	botEntity, err := service.SyncByToken(context.Background(), "777:VALID")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if botEntity.ID != 777 {
		t.Fatalf("expected id 777, got %d", botEntity.ID)
	}

	if botEntity.Token != "777:VALID" {
		t.Fatalf("expected token 777:VALID, got %q", botEntity.Token)
	}

	if repository.savedBot == nil {
		t.Fatalf("expected bot to be saved")
	}

	if !txManager.called {
		t.Fatalf("expected transaction manager to be called")
	}
}

func TestServiceSyncByTokenUpdate(t *testing.T) {
	t.Parallel()

	existingBot, err := model.NewBot(777, "Old Name", "old_bot", "777:OLD")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	repository := &repositoryMock{storedBot: existingBot}
	adapter := &telegramAdapterMock{
		profile: &TelegramBotProfileDTO{ID: 777, Name: "New Name", Username: "new_bot"},
	}
	txManager := &transactionManagerMock{}

	service := NewService(repository, adapter, txManager)
	botEntity, err := service.SyncByToken(context.Background(), "777:NEW")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if botEntity.Name != "New Name" {
		t.Fatalf("expected name New Name, got %q", botEntity.Name)
	}

	if botEntity.Username != "new_bot" {
		t.Fatalf("expected username new_bot, got %v", botEntity.Username)
	}

	if botEntity.Token != "777:NEW" {
		t.Fatalf("expected token 777:NEW, got %q", botEntity.Token)
	}

	if !txManager.called {
		t.Fatalf("expected transaction manager to be called")
	}
}

func TestServiceSyncByTokenErrors(t *testing.T) {
	t.Parallel()

	t.Run("telegram adapter error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("telegram failed")
		service := NewService(&repositoryMock{}, &telegramAdapterMock{err: expectedErr}, &transactionManagerMock{})

		_, err := service.SyncByToken(context.Background(), "777:BAD")
		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected error %v, got %v", expectedErr, err)
		}
	})

	t.Run("repository save error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("save failed")
		repository := &repositoryMock{saveErr: expectedErr}
		adapter := &telegramAdapterMock{profile: &TelegramBotProfileDTO{ID: 777, Name: "OIDC Bot", Username: "oidc_login_bot"}}
		service := NewService(repository, adapter, &transactionManagerMock{})

		_, err := service.SyncByToken(context.Background(), "777:VALID")
		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected error %v, got %v", expectedErr, err)
		}
	})

	t.Run("transaction manager error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("tx failed")
		txManager := &transactionManagerMock{err: expectedErr}
		service := NewService(&repositoryMock{}, &telegramAdapterMock{profile: &TelegramBotProfileDTO{ID: 777, Name: "OIDC Bot", Username: "oidc_login_bot"}}, txManager)

		_, err := service.SyncByToken(context.Background(), "777:VALID")
		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected error %v, got %v", expectedErr, err)
		}
	})
}
