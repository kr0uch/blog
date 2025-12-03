package errors

import "errors"

var (
	ErrUserAlreadyExists      = errors.New("user already exists")
	ErrEmailInvalid           = errors.New("invalid email")
	ErrRoleInvalid            = errors.New("invalid role")
	ErrInvalidEmailOrPassword = errors.New("invalid email or password")
)
