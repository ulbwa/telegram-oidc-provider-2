package bot

import (
	"context"
	"errors"
	"testing"

	"github.com/ulbwa/telegram-oidc-provider/internal/model"
	"github.com/ulbwa/telegram-oidc-provider/internal/service/persistence"
	"github.com/ulbwa/telegram-oidc-provider/internal/service/telegram"
)

type transactionMock struct {
	commitErr   error
	rollbackErr error
	closeErr    error

	commits   int
	rollbacks int
	closes    int
}

func (m *transactionMock) Savepoint(ctx context.Context, name string) error {
	return nil
}

func (m *transactionMock) Commit(ctx context.Context) error {
	m.commits++
	if m.commitErr != nil {
		return m.commitErr
	}
	return nil
}

func (m *transactionMock) Rollback(ctx context.Context) error {
	m.rollbacks++
	if m.rollbackErr != nil {
		return m.rollbackErr
	}
	return nil
}

func (m *transactionMock) RollbackSavepoint(ctx context.Context, name string) error {
	return nil
}

func (m *transactionMock) Close() error {
	m.closes++
	if m.closeErr != nil {
		return m.closeErr
	}
	return nil
}

type repositoryMock struct {
	storedBot *model.Bot
	savedBot  *model.Bot
	saveErr   error
	getErr    error
	beginErr  error
	tx        *transactionMock
}

func (m *repositoryMock) Begin(ctx context.Context) (context.Context, persistence.Transaction, error) {
	if m.beginErr != nil {
		return nil, nil, m.beginErr
	}
	if m.tx == nil {
		m.tx = &transactionMock{}
	}
	return ctx, m.tx, nil
}

func (m *repositoryMock) GetByID(ctx context.Context, id int64) (*model.Bot, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}

	if m.storedBot == nil || m.storedBot.ID != id {
		return nil, persistence.ErrNotFound
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

type telegramProviderMock struct {
	profile *telegram.Bot
	err     error
}

func (m *telegramProviderMock) FetchBot(ctx context.Context, token string) (*telegram.Bot, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.profile == nil {
		return nil, errors.New("profile is nil")
	}

	return &telegram.Bot{
		ID:       m.profile.ID,
		Name:     m.profile.Name,
		Username: m.profile.Username,
	}, nil
}

func TestNewServiceValidation(t *testing.T) {
	t.Parallel()

	provider := &telegramProviderMock{profile: &telegram.Bot{ID: 1, Name: "Bot", Username: "bot_name"}}
	repo := &repositoryMock{}

	_, err := NewService(nil, provider)
	if !errors.Is(err, ErrRepositoryNil) {
		t.Fatalf("expected error %v, got %v", ErrRepositoryNil, err)
	}

	_, err = NewService(repo, nil)
	if !errors.Is(err, ErrTelegramProviderNil) {
		t.Fatalf("expected error %v, got %v", ErrTelegramProviderNil, err)
	}
}

func TestServiceSyncByTokenCreate(t *testing.T) {
	t.Parallel()

	tx := &transactionMock{}
	repository := &repositoryMock{tx: tx}
	provider := &telegramProviderMock{
		profile: &telegram.Bot{ID: 777, Name: "OIDC Bot", Username: "oidc_login_bot"},
	}

	svc, err := NewService(repository, provider)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	botEntity, err := svc.SyncByToken(context.Background(), "777:VALID")
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

	if tx.commits != 1 {
		t.Fatalf("expected 1 commit, got %d", tx.commits)
	}

	if tx.rollbacks != 0 {
		t.Fatalf("expected 0 rollbacks, got %d", tx.rollbacks)
	}
}

func TestServiceSyncByTokenUpdate(t *testing.T) {
	t.Parallel()

	existingBot, err := model.NewBot(777, "Old Name", "old_bot", "777:OLD")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tx := &transactionMock{}
	repository := &repositoryMock{storedBot: existingBot, tx: tx}
	provider := &telegramProviderMock{
		profile: &telegram.Bot{ID: 777, Name: "New Name", Username: "new_bot"},
	}

	svc, err := NewService(repository, provider)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	botEntity, err := svc.SyncByToken(context.Background(), "777:NEW")
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

	if tx.commits != 1 {
		t.Fatalf("expected 1 commit, got %d", tx.commits)
	}
}

func TestServiceSyncByTokenErrors(t *testing.T) {
	t.Parallel()

	t.Run("telegram provider error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("telegram failed")
		svc, err := NewService(&repositoryMock{}, &telegramProviderMock{err: expectedErr})
		if err != nil {
			t.Fatalf("expected no constructor error, got %v", err)
		}

		_, gotErr := svc.SyncByToken(context.Background(), "777:BAD")
		if !errors.Is(gotErr, expectedErr) {
			t.Fatalf("expected error %v, got %v", expectedErr, gotErr)
		}
	})

	t.Run("begin error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("begin failed")
		repository := &repositoryMock{beginErr: expectedErr}
		provider := &telegramProviderMock{profile: &telegram.Bot{ID: 777, Name: "OIDC Bot", Username: "oidc_login_bot"}}
		svc, err := NewService(repository, provider)
		if err != nil {
			t.Fatalf("expected no constructor error, got %v", err)
		}

		_, gotErr := svc.SyncByToken(context.Background(), "777:VALID")
		if !errors.Is(gotErr, expectedErr) {
			t.Fatalf("expected error %v, got %v", expectedErr, gotErr)
		}
	})

	t.Run("repository save error triggers rollback", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("save failed")
		tx := &transactionMock{}
		repository := &repositoryMock{saveErr: expectedErr, tx: tx}
		provider := &telegramProviderMock{profile: &telegram.Bot{ID: 777, Name: "OIDC Bot", Username: "oidc_login_bot"}}
		svc, err := NewService(repository, provider)
		if err != nil {
			t.Fatalf("expected no constructor error, got %v", err)
		}

		_, gotErr := svc.SyncByToken(context.Background(), "777:VALID")
		if !errors.Is(gotErr, expectedErr) {
			t.Fatalf("expected error %v, got %v", expectedErr, gotErr)
		}

		if tx.rollbacks != 1 {
			t.Fatalf("expected 1 rollback, got %d", tx.rollbacks)
		}

		if tx.commits != 0 {
			t.Fatalf("expected 0 commits, got %d", tx.commits)
		}
	})
}
