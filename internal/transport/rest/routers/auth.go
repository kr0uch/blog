package routers

import (
	"blog/internal/repository"
	"blog/internal/service"
	"blog/internal/transport/rest/controllers"
	"net/http"
)

func NewAuthRouter(repo *repository.BlogRepository, secret string) (*http.ServeMux, *service.AuthService) {
	srv := service.NewAuthService(repo, secret)
	controller := controllers.NewAuthController(srv)
	router := http.NewServeMux()

	router.HandleFunc("POST /auth/register", controller.RegistrateUser)
	router.HandleFunc("POST /auth/login", controller.LoginUser)
	router.HandleFunc("POST /auth/refresh-token", controller.RefreshUserToken)

	return router, srv
}
