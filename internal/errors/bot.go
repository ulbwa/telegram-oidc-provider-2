package errors

import "errors"

var (
	ErrBotInvalidID                = errors.New("bot id must be greater than zero")
	ErrBotNameRequired             = errors.New("bot name is required")
	ErrBotNameInvalid              = errors.New("bot name is invalid")
	ErrBotNameHasOuterSpaces       = errors.New("bot name cannot contain leading or trailing spaces")
	ErrBotUsernameRequired         = errors.New("bot username is required")
	ErrBotUsernameInvalid          = errors.New("bot username is invalid")
	ErrBotUsernameOuterSpaces      = errors.New("bot username cannot contain leading or trailing spaces")
	ErrBotTokenRequired            = errors.New("bot token is required")
	ErrBotTokenInvalid             = errors.New("bot token is invalid")
	ErrBotTokenOuterSpaces         = errors.New("bot token cannot contain leading or trailing spaces")
	ErrBotOAuthClientIDInvalid     = errors.New("bot oauth client id is invalid")
	ErrBotOAuthClientIDOuterSpaces = errors.New("bot oauth client id cannot contain leading or trailing spaces")
	ErrBotCreatedAtInvalid         = errors.New("bot created time is invalid")
	ErrBotUpdatedAtInvalid         = errors.New("bot updated time is invalid")
	ErrBotUpdatedBeforeCreate      = errors.New("bot updated time cannot be before created time")
)
