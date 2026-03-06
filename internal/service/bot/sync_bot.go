package bot

import (
	"context"
	"errors"

	"github.com/rs/zerolog"

	"github.com/ulbwa/telegram-oidc-provider/internal/model"
	"github.com/ulbwa/telegram-oidc-provider/internal/service/persistence"
	"github.com/ulbwa/telegram-oidc-provider/internal/service/telegram"
)

func (s *service) createFromTelegram(profile *telegram.Bot, token string) (*model.Bot, error) {
	return model.NewBot(profile.ID, profile.Name, profile.Username, token)
}

func (s *service) updateFromTelegram(bot *model.Bot, profile *telegram.Bot, token string) error {
	err := bot.UpdateProfile(profile.Name, profile.Username)
	if err != nil {
		return err
	}

	err = bot.SetToken(token)
	if err != nil {
		return err
	}

	return nil
}

func (s *service) SyncByToken(ctx context.Context, token string) (bot *model.Bot, err error) {
	profile, err := s.tgProvider.FetchBot(ctx, token)
	if err != nil {
		return nil, err
	}

	ctx, tx, err := s.repository.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := tx.Close(); closeErr != nil {
			zerolog.Ctx(ctx).Error().Err(closeErr).Msg("failed to close transaction in SyncByToken")
		}
	}()

	err = persistence.WithinTransaction(ctx, tx, func() error {
		bot, err = s.repository.GetByID(ctx, profile.ID)
		if err != nil {
			if !errors.Is(err, persistence.ErrNotFound) {
				return err
			}

			bot, err = s.createFromTelegram(profile, token)
			if err != nil {
				return err
			}
		} else {
			err = s.updateFromTelegram(bot, profile, token)
			if err != nil {
				return err
			}
		}

		return s.repository.Save(ctx, bot)
	})
	if err != nil {
		return nil, err
	}

	return bot, nil
}
