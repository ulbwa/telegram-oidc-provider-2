package persistence

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/rs/zerolog"
)

var ErrTransactionAlreadyExists = errors.New("transaction already exists and cannot be created again")

type TransactionFactory interface {
	Begin(ctx context.Context) (context.Context, Transaction, error)
}

var (
	ErrTransactionDead              = errors.New("transaction is dead and operation cannot be performed")
	ErrTransactionAlreadyCommitted  = errors.New("transaction already committed and cannot be rolled back")
	ErrTransactionAlreadyRolledBack = errors.New("transaction already rolled back and cannot be committed")
	ErrSavepointNotFound            = errors.New("savepoint with given name not found")
)

type Transaction interface {
	io.Closer
	Savepoint(ctx context.Context, name string) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
	RollbackSavepoint(ctx context.Context, name string) error
}

func WithinTransaction(ctx context.Context, tx Transaction, fn func() error) (err error) {
	defer func() {
		if closeErr := tx.Close(); closeErr != nil {
			zerolog.Ctx(ctx).Error().Err(closeErr).Msg("failed to close transaction")
		}
	}()

	if err = fn(); err != nil {
		rollbackCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		rollbackErr := tx.Rollback(rollbackCtx)
		if rollbackErr != nil {
			return errors.Join(err, rollbackErr)
		}
		return err
	}

	return tx.Commit(ctx)
}
