package controllers

import (
	"blog/internal/models/dto"
	"blog/pkg/consts/errors"
	"encoding/json"
	stderr "errors"
	"log"
	"net/http"
)

type AuthService interface {
	RegistrateUser(user *dto.RegistrateUserRequest) (*dto.RegistrateUserResponse, error)
	LoginUser(user *dto.LoginUserRequest) (*dto.LoginUserResponse, error)
	RefreshUserToken(token *dto.RefreshUserTokenRequest) (*dto.RefreshUserTokenResponse, error)
}
type AuthController struct {
	srv AuthService
}

func NewAuthRouter(srv AuthService) *AuthController {
	return &AuthController{
		srv: srv,
	}
}

// RegistrateUser godoc
// @Summary Зарегистрировать пользователя
// @Tags Роли пользователей и аутентификация
// @Accept json
// @Produce json
// @Param request body dto.RegistrateUserRequest true "Данные пользователя"
// @Success 200 {object} dto.RegistrateUserResponse
// @Failure 403 {string} errors.ErrUserAlreadyExists "user already exists"
// @Failure 400 {string} errors.ErrInvalidEmail "invalid email"
// @Router /api/auth/register [post]
func (c *AuthController) RegistrateUser(w http.ResponseWriter, r *http.Request) {
	var request dto.RegistrateUserRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println(err)
		http.Error(w, errors.ErrIncorrectData.Error(), http.StatusBadRequest)
		return
	}

	response, err := c.srv.RegistrateUser(&request)
	if err != nil {
		switch {
		case stderr.Is(err, errors.ErrUserAlreadyExists):
			http.Error(w, err.Error(), http.StatusForbidden)
		case stderr.Is(err, errors.ErrInvalidEmail):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusForbidden)
		}
		return
	}

	w.Header().Set("Authorization", "Bearer "+response.AccessToken)

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println(err)
		http.Error(w, errors.ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}
}

// LoginUser godoc
// @Summary Залогинить пользователя
// @Tags Роли пользователей и аутентификация
// @Accept json
// @Produce json
// @Param request body dto.LoginUserRequest true "Данные пользователя"
// @Success 200 {object} dto.LoginUserResponse
// @Failure 403 {string} errors.ErrInvalidEmailOrPassword
// @Router /api/auth/login [post]
func (c *AuthController) LoginUser(w http.ResponseWriter, r *http.Request) {
	var request dto.LoginUserRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println(err)
		http.Error(w, errors.ErrIncorrectData.Error(), http.StatusBadRequest)
		return
	}

	response, err := c.srv.LoginUser(&request)
	if err != nil {
		switch {
		case stderr.Is(err, errors.ErrInvalidEmailOrPassword):
			http.Error(w, err.Error(), http.StatusForbidden)
		default:
			http.Error(w, err.Error(), http.StatusForbidden)
		}
		return
	}

	w.Header().Set("Authorization", "Bearer "+response.AccessToken)

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println(err)
		http.Error(w, errors.ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}
}

// RefreshUserToken godoc
// @Summary Обновить токен пользователя
// @Tags Роли пользователей и аутентификация
// @Accept json
// @Produce json
// @Param request body dto.RefreshUserTokenRequest true "Данные пользователя"
// @Success 200 {object} dto.RefreshUserTokenResponse
// @Failure 400 {string} errors.ErrInvalidRefreshToken
// @Router /api/auth/refresh-token [post]
func (c *AuthController) RefreshUserToken(w http.ResponseWriter, r *http.Request) {
	var request dto.RefreshUserTokenRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println(err)
		http.Error(w, errors.ErrIncorrectData.Error(), http.StatusBadRequest)
		return
	}

	response, err := c.srv.RefreshUserToken(&request)
	if err != nil {
		switch {
		case stderr.Is(err, errors.ErrInvalidRefreshToken):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusForbidden)
		}
		return
	}

	w.Header().Set("Authorization", "Bearer "+response.AccessToken)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println(err)
		http.Error(w, errors.ErrInternalServerError.Error(), http.StatusInternalServerError)
		return
	}
}
