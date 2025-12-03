package servers

import (
	"blog/api"
	_ "blog/docs"
	"blog/internal/database/postgre"
	"blog/internal/repository"
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
	authRouter, _ := routers.NewAuthRouter(repo)

	mainRouter.Handle("/auth/", authRouter)

	mainRouter.Handle("/api/", http.StripPrefix("/api", mainRouter))
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
