package telegram

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrBotTokenInvalid  = errors.New("telegram bot token is invalid")
	ErrBotTokenRequired = fmt.Errorf("%w: token is empty or whitespace", ErrBotTokenInvalid)
	ErrBadResponse      = errors.New("telegram API returned bad response")
)

// Bot is a service-layer contract DTO for Telegram bot data.
type Bot struct {
	ID       int64
	Name     string
	Username string
}

type Provider interface {
	FetchBot(ctx context.Context, token string) (*Bot, error)
}

// // User is a service-layer contract DTO for Telegram user data.
// type User struct {
// 	ID        int64
// 	FirstName string
// 	LastName  *string
// 	Username  *string
// }

// type AuthData struct {
// 	User     User
// 	AuthDate time.Time
// 	Raw      string
// 	Hash     string
// }

// type AuthService interface {
// 	Verify(raw map[string]any, token string) (*AuthData, error)
// }
