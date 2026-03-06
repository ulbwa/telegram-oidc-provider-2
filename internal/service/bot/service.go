package bot

import (
	"context"
	"errors"

	"github.com/ulbwa/telegram-oidc-provider/internal/model"
	"github.com/ulbwa/telegram-oidc-provider/internal/service/telegram"
)

type Service interface {
	SyncByToken(ctx context.Context, token string) (*model.Bot, error)
}

type service struct {
	repository Repository
	tgProvider telegram.Provider
}

var (
	ErrRepositoryNil       = errors.New("repository is nil")
	ErrTelegramProviderNil = errors.New("telegram provider is nil")
)

func NewService(repository Repository, telegramProvider telegram.Provider) (*service, error) {
	if repository == nil {
		return nil, ErrRepositoryNil
	}
	if telegramProvider == nil {
		return nil, ErrTelegramProviderNil
	}
	return &service{
		repository: repository,
		tgProvider: telegramProvider,
	}, nil
}
