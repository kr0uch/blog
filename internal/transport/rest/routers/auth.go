package routers

import (
	"blog/internal/repository"
	"blog/internal/service"
	"blog/internal/transport/rest/controllers"
	"net/http"
)

func NewAuthRouter(repo *repository.BlogRepository) (*http.ServeMux, *service.AuthService) {
	srv := service.NewAuthService(repo, "secret")
	controller := controllers.NewAuthRouter(srv)
	router := http.NewServeMux()

	router.HandleFunc("POST /auth/register", controller.RegistrateUser)

	return router, srv
}
