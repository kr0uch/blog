package service

import (
	"blog/internal/models"
	"blog/pkg/consts"
	"blog/pkg/consts/errors"
	"blog/pkg/utils/hash"
	"blog/pkg/utils/jwt"
	"blog/pkg/utils/mail"
	"time"
)

type BlogRepository interface {
	CreateUser(email, passwordHash, role, refreshToken string, refreshTokenExpiryTime time.Time) (*models.User, error)
	GetUser(email string) (*models.User, error)
	UpdateRefreshToken(userId, refreshToken string) error
}

type AuthService struct {
	repo   BlogRepository
	secret string
}

func NewAuthService(repo BlogRepository, secret string) *AuthService {
	return &AuthService{
		repo:   repo,
		secret: secret,
	}
}

func (s *AuthService) RegistrateUser(user *models.RegistrateUserRequest) (*models.RegistrateUserResponse, error) {
	if !mail.IsValidEmail(user.Email) {
		return nil, errors.ErrEmailInvalid
	}

	if user.Role != consts.AuthorRole && user.Role != consts.ReaderRole {
		return nil, errors.ErrRoleInvalid
	}

	passwordHash, err := hash.HashString(user.Password)
	if err != nil {
		return nil, err
	}

	refreshToken, err := jwt.NewRefreshToken(user.Email, s.secret)
	if err != nil {
		return nil, err
	}

	refreshTokenExpiryTime := time.Now().Add(time.Hour * 24 * 7)

	newUser, err := s.repo.CreateUser(user.Email, passwordHash, user.Role, refreshToken, refreshTokenExpiryTime)
	if err != nil {
		return nil, err
	}

	accessToken := jwt.NewAccessToken(newUser.UserId, s.secret)

	responseUser := &models.RegistrateUserResponse{
		Id:           newUser.UserId,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return responseUser, nil
}

func (s *AuthService) LoginUser(user *models.LoginUserRequest) (*models.LoginUserResponse, error) {
	if !mail.IsValidEmail(user.Email) {
		return nil, errors.ErrEmailInvalid
	}

	newUser, err := s.repo.GetUser(user.Email)
	if err != nil {
		return nil, err
	}

	success, err := hash.CompareHashString(user.Password, newUser.PasswordHash)
	if err != nil {
		return nil, errors.ErrInvalidEmailOrPassword
	}
	if !success {
		return nil, errors.ErrInvalidEmailOrPassword
	}

	refreshToken, err := jwt.NewRefreshToken(user.Email, s.secret)
	if err != nil {
		return nil, err
	}

	if err = s.repo.UpdateRefreshToken(newUser.UserId, refreshToken); err != nil {
		return nil, err
	}

	accessToken := jwt.NewAccessToken(newUser.UserId, s.secret)

	responseUser := &models.LoginUserResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return responseUser, nil
}
