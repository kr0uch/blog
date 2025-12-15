package errors

import (
	"errors"
)

var (
	ErrInternalServerError = errors.New("internal server error")

	ErrFailedOpenDB        = errors.New("failed to open database")
	ErrFailedCheckDBExists = errors.New("failed to check if database exists")
	ErrFailedCreateDB      = errors.New("failed to create database")
	ErrFailedConnectDB     = errors.New("failed to connect to database")

	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidEmail      = errors.New("invalid email")
	ErrInvalidRole       = errors.New("invalid role")

	ErrInvalidEmailOrPassword = errors.New("invalid email or password")
	ErrInvalidIdempotencyKey  = errors.New("invalid idempotency key")

	ErrInvalidAccessToken  = errors.New("invalid access token")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")

	ErrNoPermission  = errors.New("no permission")
	ErrIncorrectData = errors.New("incorrect data")

	ErrPostNotFound        = errors.New("post not found")
	ErrPostOrImageNotFound = errors.New("post or image not found")

	ErrInvalidPostId     = errors.New("invalid post id")
	ErrInvalidPostStatus = errors.New("invalid post status")
	ErrInvalidUser       = errors.New("invalid user")
	ErrInvalidUserId     = errors.New("invalid user id")

	ErrMinioBucketNotExists    = errors.New("minio bucket does not exist")
	ErrMinioMakeBucket         = errors.New("minio cant make bucket")
	ErrMinioPutObject          = errors.New("minio cant put object")
	ErrMinioPresignedGetObject = errors.New("minio cant presigned get object")
	ErrMinioGetObject          = errors.New("minio cant get object")
	ErrMinioRemoveObject       = errors.New("minio cant remove object")

	ErrInvalidImageId = errors.New("invalid image id")
)
