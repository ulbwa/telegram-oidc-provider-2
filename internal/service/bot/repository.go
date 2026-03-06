package bot

import (
	"context"

	"github.com/ulbwa/telegram-oidc-provider/internal/model"
	"github.com/ulbwa/telegram-oidc-provider/internal/service/persistence"
)

type Repository interface {
	persistence.TransactionFactory
	GetByID(ctx context.Context, id int64) (*model.Bot, error)
	Save(ctx context.Context, bot *model.Bot) error
}
