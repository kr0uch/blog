package service

import (
	"blog/internal/logger"
	"blog/internal/models/dto"
	"blog/internal/models/entities"
	"blog/pkg/consts"
	"blog/pkg/consts/errors"
	"blog/pkg/utils/hash"
	"blog/pkg/utils/jwt"
	"blog/pkg/utils/mail"
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AuthBlogRepository interface {
	CreateUser(email, passwordHash, role, refreshToken string, refreshTokenExpiryTime time.Time) (*entities.User, error)
	GetUserByEmail(email string) (*entities.User, error)
	GetUserByRefreshToken(refreshToken string) (*entities.User, error)
	GetUserById(userId string) (*entities.User, error)
	UpdateRefreshToken(userId, refreshToken string) error
}

type AuthService struct {
	repo   AuthBlogRepository
	secret string
}

func NewAuthService(repo AuthBlogRepository, secret string) *AuthService {
	return &AuthService{
		repo:   repo,
		secret: secret,
	}
}

func (s *AuthService) RegistrateUser(ctx context.Context, user *dto.RegistrateUserRequest) (*dto.RegistrateUserResponse, error) {
	reqLogger := logger.LoggerFromContext(ctx).WithFields(zap.String("operation", "RegistrateUser"))

	reqLogger.Info("Register user")

	if !mail.IsValidEmail(user.Email) {
		reqLogger.Error("Invalid email", zap.String("email", user.Email))
		return nil, errors.ErrInvalidEmail
	}

	if user.Role != consts.AuthorRole && user.Role != consts.ReaderRole {
		reqLogger.Error("Invalid role", zap.String("role", user.Role))
		return nil, errors.ErrInvalidRole
	}

	passwordHash, err := hash.HashString(user.Password)
	if err != nil {
		reqLogger.Error("Failed to hash password", zap.Error(err))
		return nil, err
	}

	refreshToken, err := jwt.NewRefreshToken(user.Email, s.secret)
	if err != nil {
		reqLogger.Error("Failed to generate refresh token", zap.Error(err))
		return nil, err
	}

	refreshTokenExpiryTime := time.Now().Add(time.Hour * 24 * 7)

	newUser, err := s.repo.CreateUser(user.Email, passwordHash, user.Role, refreshToken, refreshTokenExpiryTime)
	if err != nil {
		reqLogger.Error("Failed to create user", zap.Error(err))
		return nil, err
	}

	accessToken := jwt.NewAccessToken(newUser.UserId, s.secret)

	var message string
	if newUser != nil {
		message = "Registered successfully"
	}

	responseUser := &dto.RegistrateUserResponse{
		Message:      message,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	reqLogger.Info("Register user done")

	return responseUser, nil
}

func (s *AuthService) LoginUser(ctx context.Context, user *dto.LoginUserRequest) (*dto.LoginUserResponse, error) {
	reqLogger := logger.LoggerFromContext(ctx).WithFields(zap.String("operation", "LoginUser"))

	reqLogger.Info("Login user")

	if !mail.IsValidEmail(user.Email) {
		reqLogger.Error("Invalid email", zap.String("email", user.Email))
		return nil, errors.ErrInvalidEmail
	}

	newUser, err := s.repo.GetUserByEmail(user.Email)
	if err != nil {
		reqLogger.Error("Failed to get user by email", zap.Error(err))
		return nil, err
	}

	success, err := hash.CompareHashString(user.Password, newUser.PasswordHash)
	if err != nil {
		reqLogger.Error("Failed to compare password", zap.Error(err))
		return nil, errors.ErrInvalidEmailOrPassword
	}
	if !success {
		reqLogger.Error("Invalid email or password", zap.String("email", user.Email))
		return nil, errors.ErrInvalidEmailOrPassword
	}

	refreshToken, err := jwt.NewRefreshToken(user.Email, s.secret)
	if err != nil {
		reqLogger.Error("Failed to generate refresh token", zap.Error(err))
		return nil, err
	}

	if err = s.repo.UpdateRefreshToken(newUser.UserId, refreshToken); err != nil {
		reqLogger.Error("Failed to update refresh token", zap.Error(err))
		return nil, err
	}

	accessToken := jwt.NewAccessToken(newUser.UserId, s.secret)

	var message string
	if newUser != nil {
		message = "Logged in successfully"
	}

	responseUser := &dto.LoginUserResponse{
		Message:      message,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	reqLogger.Info("Login user done")

	return responseUser, nil
}

func (s *AuthService) RefreshUserToken(ctx context.Context, token *dto.RefreshUserTokenRequest) (*dto.RefreshUserTokenResponse, error) {
	reqLogger := logger.LoggerFromContext(ctx).WithFields(zap.String("operation", "RefreshUserToken"))

	reqLogger.Info("Refresh user token")

	_, err := jwt.ValidateToken(token.RefreshToken, s.secret)
	if err != nil {
		reqLogger.Error("Failed to validate refresh token", zap.Error(err))
		return nil, errors.ErrInvalidRefreshToken
	}

	newUser, err := s.repo.GetUserByRefreshToken(token.RefreshToken)
	if err != nil {
		reqLogger.Error("Failed to get user by refresh token", zap.Error(err))
		return nil, err
	}

	accessToken := jwt.NewAccessToken(newUser.UserId, s.secret)

	var message string
	if newUser != nil {
		message = "Refresh tokens successfully"
	}

	responseToken := &dto.RefreshUserTokenResponse{
		Message:      message,
		AccessToken:  accessToken,
		RefreshToken: token.RefreshToken,
	}

	reqLogger.Info("Refresh user token done")

	return responseToken, nil
}

func (s *AuthService) AuthorizeUser(ctx context.Context, token string) (*entities.User, error) {
	reqLogger := logger.LoggerFromContext(ctx).WithFields(zap.String("operation", "AuthorizeUser"))

	reqLogger.Info("Authorize user")

	claims, err := jwt.ValidateToken(token, s.secret)
	if err != nil {
		reqLogger.Error("Failed to validate token", zap.Error(err))
		return nil, err
	}

	sub, ok := (*claims)["sub"].(string)
	if !ok {
		reqLogger.Error("Failed to get sub from token", zap.String("token", token))
		return nil, errors.ErrInvalidAccessToken
	}

	id, err := uuid.Parse(sub)
	if err != nil {
		reqLogger.Error("Failed to parse sub from token", zap.String("token", token))
		return nil, err
	}
	user, err := s.repo.GetUserById(id.String())
	if err != nil {
		reqLogger.Error("Failed to get user by id", zap.String("id", id.String()))
		return nil, err
	}

	reqLogger.Info("Authorize user done")

	return user, nil
}
