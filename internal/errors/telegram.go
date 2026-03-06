package errors

import "errors"

var (
	ErrTelegramBotTokenRequired    = errors.New("telegram bot token is required")
	ErrTelegramBotTokenInvalid     = errors.New("telegram bot token is invalid")
	ErrTelegramBotUsernameRequired = errors.New("telegram bot username is required")
	ErrTelegramAPIResponse         = errors.New("telegram api returned invalid response")
)
