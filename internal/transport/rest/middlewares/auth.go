package middlewares

import (
	"blog/internal/models/entities"
	"blog/pkg/consts"
	"context"
	"net/http"
	"strings"
)

type AuthService interface {
	AuthorizeUser(token string) (*entities.User, error)
}
type AuthMiddlewareHandler struct {
	srv AuthService
}

func NewAuthMiddlewareHandler(srv AuthService) *AuthMiddlewareHandler {
	return &AuthMiddlewareHandler{
		srv: srv,
	}
}

func (m *AuthMiddlewareHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			//w.WriteHeader(http.StatusUnauthorized)
			//unauthorized
			return
		}
		token := strings.Split(header, " ")
		if len(token) != 2 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if token[0] != "Bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		user, err := m.srv.AuthorizeUser(token[1])
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), consts.CtxUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
