package service

import (
	"blog/internal/models/dto"
	"blog/internal/models/entities"
	"blog/pkg/consts"
	"blog/pkg/consts/errors"
	"blog/pkg/utils/hash"
	"blog/pkg/utils/jwt"
	"blog/pkg/utils/mail"
	"time"

	"github.com/google/uuid"
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

func (s *AuthService) RegistrateUser(user *dto.RegistrateUserRequest) (*dto.RegistrateUserResponse, error) {
	if !mail.IsValidEmail(user.Email) {
		return nil, errors.ErrInvalidEmail
	}

	if user.Role != consts.AuthorRole && user.Role != consts.ReaderRole {
		return nil, errors.ErrInvalidRole
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

	var message string
	if newUser != nil {
		message = "Registered successfully"
	}

	responseUser := &dto.RegistrateUserResponse{
		Message:      message,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return responseUser, nil
}

func (s *AuthService) LoginUser(user *dto.LoginUserRequest) (*dto.LoginUserResponse, error) {
	if !mail.IsValidEmail(user.Email) {
		return nil, errors.ErrInvalidEmail
	}

	newUser, err := s.repo.GetUserByEmail(user.Email)
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

	var message string
	if newUser != nil {
		message = "Logged in successfully"
	}

	responseUser := &dto.LoginUserResponse{
		Message:      message,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return responseUser, nil
}

func (s *AuthService) RefreshUserToken(token *dto.RefreshUserTokenRequest) (*dto.RefreshUserTokenResponse, error) {
	_, err := jwt.ValidateToken(token.RefreshToken, s.secret)
	if err != nil {
		return nil, errors.ErrInvalidRefreshToken
	}

	newUser, err := s.repo.GetUserByRefreshToken(token.RefreshToken)
	if err != nil {
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

	return responseToken, nil
}

func (s *AuthService) AuthorizeUser(token string) (*entities.User, error) {
	claims, err := jwt.ValidateToken(token, s.secret)
	if err != nil {
		return nil, err
	}

	sub, ok := (*claims)["sub"].(string)
	if !ok {
		return nil, errors.ErrInvalidAccessToken
	}

	id, err := uuid.Parse(sub)
	if err != nil {
		return nil, err
	}
	user, err := s.repo.GetUserById(id.String())
	if err != nil {
		return nil, err
	}

	return user, nil
}
