package errors

import "errors"

var (
	ErrUserFirstNameRequired       = errors.New("user first name is required")
	ErrUserFirstNameInvalid        = errors.New("user first name is invalid")
	ErrUserFirstNameHasOuterSpaces = errors.New("user first name cannot contain leading or trailing spaces")
	ErrUserLastNameInvalid         = errors.New("user last name is invalid")
	ErrUserLastNameHasOuterSpaces  = errors.New("user last name cannot contain leading or trailing spaces")
	ErrUserUsernameInvalid         = errors.New("user username is invalid")
	ErrUserUsernameHasOuterSpaces  = errors.New("user username cannot contain leading or trailing spaces")
	ErrUserPhotoURLInvalid         = errors.New("user photo url is invalid")
)
