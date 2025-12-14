package controllers

import (
	"blog/internal/models/dto"
	"encoding/json"
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
// @Router /api/auth/register [post]
func (c *AuthController) RegistrateUser(w http.ResponseWriter, r *http.Request) {
	var request dto.RegistrateUserRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response, err := c.srv.RegistrateUser(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		err = json.NewEncoder(w).Encode(dto.ErrorResponse{Error: err.Error()})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
		}
		return
	}

	w.Header().Set("Authorization", "Bearer "+response.AccessToken)

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
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
// @Router /api/auth/login [post]
func (c *AuthController) LoginUser(w http.ResponseWriter, r *http.Request) {
	var request dto.LoginUserRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response, err := c.srv.LoginUser(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		err = json.NewEncoder(w).Encode(dto.ErrorResponse{Error: err.Error()})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
		return
	}

	w.Header().Set("Authorization", "Bearer "+response.AccessToken)

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
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
// @Router /api/auth/refresh-token [post]
func (c *AuthController) RefreshUserToken(w http.ResponseWriter, r *http.Request) {
	var request dto.RefreshUserTokenRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response, err := c.srv.RefreshUserToken(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		err = json.NewEncoder(w).Encode(dto.ErrorResponse{Error: err.Error()})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
		return
	}

	w.Header().Set("Authorization", "Bearer "+response.AccessToken)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
