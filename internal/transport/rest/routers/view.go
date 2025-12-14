package routers

import (
	"blog/internal/repository"
	"blog/internal/service"
	"blog/internal/transport/rest/controllers"
	"net/http"
)

func NewViewRouter(repo *repository.BlogRepository) *http.ServeMux {
	srv := service.NewViewService(repo)
	controller := controllers.NewViewController(srv)
	router := http.NewServeMux()

	router.HandleFunc("GET /posts", controller.ViewPosts)

	return router
}
