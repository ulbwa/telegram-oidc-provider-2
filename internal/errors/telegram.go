package errors

import "errors"

var (
	ErrTelegramBotTokenInvalid = errors.New("telegram bot token is invalid")
	ErrTelegramAPIResponse     = errors.New("telegram api returned invalid response")
)
