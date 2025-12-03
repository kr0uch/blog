package controllers

import (
	"blog/internal/models"
	"encoding/json"
	"log"
	"net/http"
)

type AuthService interface {
	RegistrateUser(user *models.RegistrateUserRequest) (*models.RegistrateUserResponse, error)
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
// @Tags Аутентификация
// @Accept json
// @Produce json
// @Param request body models.RegistrateUserRequest true "Данные пользователя"
// @Success 200 {object} models.RegistrateUserResponse
// @Router /api/auth/register [post]
func (c *AuthController) RegistrateUser(w http.ResponseWriter, r *http.Request) {
	var request models.RegistrateUserRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	response, err := c.srv.RegistrateUser(&request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		err = json.NewEncoder(w).Encode(models.ErrorResponse{Error: err.Error()})
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
