package routers

import (
	"blog/internal/repository"
	"blog/internal/service"
	"blog/internal/transport/rest/controllers"
	"net/http"
)

func NewPostsRouter(repo *repository.BlogRepository) *http.ServeMux {
	srv := service.NewPostsService(repo, "secret")
	controller := controllers.NewPostsController(srv)
	router := http.NewServeMux()

	router.HandleFunc("POST /posts", controller.CreatePost)
	router.HandleFunc("POST /posts/{postId}/images", controller.AddImageToPost)
	router.HandleFunc("PUT /posts/{postId}", controller.EditPost)
	router.HandleFunc("DELETE /posts/{postId}/images/{imageId}", controller.DeleteImageFromPost)
	router.HandleFunc("PATCH /posts/{postId}/status", controller.PublishPost)

	return router
}
