package servers

import (
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

	repo := repository.NewBlogRepository(db.DB)
	authRouter, _ := routers.NewAuthRouter(repo)

	mainRouter.Handle("/auth/", authRouter)

	http.DefaultServeMux.Handle("/api/", http.StripPrefix("/api", mainRouter))

	server := &http.Server{Addr: fmt.Sprintf(":%s", cfg.Port)}

	return &BlogServer{
		cfg:    cfg,
		server: server,
	}
}

func (srv *BlogServer) Start() error {
	log.Printf("Starting server on port %s", srv.server.Addr)
	return srv.server.ListenAndServe()
}
