package errors

import "errors"

var (
	ErrUserAlreadyExists      = errors.New("user already exists")
	ErrEmailInvalid           = errors.New("invalid email")
	ErrRoleInvalid            = errors.New("invalid role")
	ErrInvalidEmailOrPassword = errors.New("invalid email or password")
	ErrInvalidIdempotencyKey  = errors.New("invalid idempotency key")
	ErrInvalidAccessToken     = errors.New("invalid access token")
	ErrInvalidPostId          = errors.New("invalid post id")
	ErrInvalidPostState       = errors.New("invalid post status")
	ErrInvalidUser            = errors.New("invalid user")
	ErrMinioBucketNotExists   = errors.New("minio bucket does not exist")
	ErrInvalidImageId         = errors.New("invalid image id")
)
