package bot

import (
	"context"
	"errors"

	errs "github.com/ulbwa/telegram-oidc-provider/internal/errors"
	"github.com/ulbwa/telegram-oidc-provider/internal/model"
	txsvc "github.com/ulbwa/telegram-oidc-provider/internal/service/tx"
)

type Repository interface {
	GetByID(ctx context.Context, id int64) (*model.Bot, error)
	Save(ctx context.Context, bot *model.Bot) error
}

type Service struct {
	repository      Repository
	telegramAdapter TelegramProvider
	txManager       txsvc.Manager
}

func NewService(repository Repository, telegramAdapter TelegramProvider, txManager txsvc.Manager) *Service {
	if txManager == nil {
		txManager = txsvc.NoOpManager{}
	}

	return &Service{
		repository:      repository,
		telegramAdapter: telegramAdapter,
		txManager:       txManager,
	}
}

// SyncByToken fetches bot info from Telegram by token and performs create-or-update in storage.
func (s *Service) SyncByToken(ctx context.Context, token string) (*model.Bot, error) {
	var syncedBot *model.Bot

	err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		botEntity, err := s.syncInTransaction(txCtx, token)
		if err != nil {
			return err
		}

		syncedBot = botEntity

		return nil
	})
	if err != nil {
		return nil, err
	}

	return syncedBot, nil
}

func (s *Service) syncInTransaction(ctx context.Context, token string) (*model.Bot, error) {
	profile, err := s.telegramAdapter.FetchBotProfile(ctx, token)
	if err != nil {
		return nil, err
	}

	botEntity, err := s.resolveBotEntity(ctx, profile, token)
	if err != nil {
		return nil, err
	}

	if err := s.repository.Save(ctx, botEntity); err != nil {
		return nil, err
	}

	return botEntity, nil
}

func (s *Service) resolveBotEntity(ctx context.Context, profile *TelegramBotProfileDTO, token string) (*model.Bot, error) {
	botEntity, err := s.repository.GetByID(ctx, profile.ID)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return model.NewBot(profile.ID, profile.Name, profile.Username, token)
		}

		return nil, err
	}

	if err := botEntity.UpdateProfile(profile.Name, profile.Username); err != nil {
		return nil, err
	}

	if err := botEntity.SetToken(token); err != nil {
		return nil, err
	}

	return botEntity, nil
}
