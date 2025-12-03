package api

import (
	_ "blog/docs"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

type Swagger struct {
	Router  *http.ServeMux
	enabled bool
}

func NewSwagger() *Swagger {
	return &Swagger{
		Router:  http.NewServeMux(),
		enabled: true,
	}
}

func (s *Swagger) Setup() {
	if !s.enabled {
		return
	}

	s.Router.HandleFunc("/", httpSwagger.WrapHandler)
}
