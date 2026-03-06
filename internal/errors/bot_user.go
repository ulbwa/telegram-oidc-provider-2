package errors

import "errors"

var (
	ErrBotUserInvalidBotID            = errors.New("bot user bot id must be greater than zero")
	ErrBotUserInvalidUserID           = errors.New("bot user id must be greater than zero")
	ErrBotUserLoginIPRequired         = errors.New("bot user login ip is required")
	ErrBotUserLoginIPInvalid          = errors.New("bot user login ip is invalid")
	ErrBotUserUserAgentInvalid        = errors.New("bot user user-agent is invalid")
	ErrBotUserUserAgentHasOuterSpaces = errors.New("bot user user-agent cannot contain leading or trailing spaces")
	ErrBotUserLanguageInvalid         = errors.New("bot user language is invalid")
	ErrBotUserLanguageHasOuterSpaces  = errors.New("bot user language cannot contain leading or trailing spaces")
	ErrBotUserCreatedAtInvalid        = errors.New("bot user created time is invalid")
	ErrBotUserUpdatedAtInvalid        = errors.New("bot user updated time is invalid")
	ErrBotUserUpdatedBeforeCreate     = errors.New("bot user updated time cannot be before created time")
	ErrBotUserLastLoginAtInvalid      = errors.New("bot user last login time is invalid")
	ErrBotUserLastLoginBeforeCreate   = errors.New("bot user last login time cannot be before created time")
)
