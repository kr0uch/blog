package routers

import (
	"blog/internal/repository"
	"blog/internal/service"
	"blog/internal/storage/minio"
	"blog/internal/transport/rest/controllers"
	"net/http"
)

// TODO: repo2 пофиксить

func NewPostsRouter(repo *repository.BlogRepository, minio *minio.MinioClient) *http.ServeMux {
	srv := service.NewPostsService(repo, minio, "secret", "data")
	controller := controllers.NewPostsController(srv)
	router := http.NewServeMux()

	router.HandleFunc("POST /posts", controller.CreatePost)
	router.HandleFunc("POST /posts/{postId}/images", controller.AddImageToPost)
	router.HandleFunc("PUT /posts/{postId}", controller.EditPost)
	router.HandleFunc("DELETE /posts/{postId}/images/{imageId}", controller.DeleteImageFromPost)
	router.HandleFunc("PATCH /posts/{postId}/status", controller.PublishPost)
	router.HandleFunc("GET /posts", controller.ViewPosts)

	return router
}
