package servers

import (
	"blog/api"
	_ "blog/docs"
	"blog/internal/database/postgre"
	"blog/internal/repository"
	"blog/internal/transport/rest/middlewares"
	"blog/internal/transport/rest/routers"
	"fmt"
	"log"
	"net/http"
)

type BlogServerConfig struct {
	Port string `env:"PORT" env-default:"8080"`
}

type BlogServer struct {
	cfg    BlogServerConfig
	server *http.Server
}

func NewBlogServer(cfg BlogServerConfig, db *postgre.DB) *BlogServer {
	mainRouter := http.NewServeMux()

	swagger := api.NewSwagger()
	swagger.Setup()

	repo := repository.NewBlogRepository(db.DB)

	authRouter, authService := routers.NewAuthRouter(repo)
	postsRouter := routers.NewPostsRouter(repo)
	viewRouter := routers.NewViewRouter(repo)

	authMiddleware := middlewares.NewAuthMiddlewareHandler(authService).AuthMiddleware
	globalMiddleware := middlewares.GlobalMiddleware

	postsRouter.Handle("/", viewRouter) //т.к. необходимо дважды / роут
	// и в целом оно ложится т.к. тут тоже работа с постом

	mainRouter.Handle("/auth/", authRouter)
	mainRouter.Handle("/", authMiddleware(postsRouter)) //т.к. /posts не совместим с /posts/{id}

	mainRouter.Handle("/api/", http.StripPrefix("/api", globalMiddleware(mainRouter)))
	mainRouter.Handle("/swagger/", swagger.Router)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: mainRouter,
	}

	return &BlogServer{
		cfg:    cfg,
		server: server,
	}
}

func (srv *BlogServer) Start() error {
	log.Printf("Starting server on port %s", srv.server.Addr)
	return srv.server.ListenAndServe()
}
