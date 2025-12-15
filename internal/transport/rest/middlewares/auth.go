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
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
		token := strings.Split(header, " ")
		if len(token) != 2 {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}

		if token[0] != "Bearer" {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}

		user, err := m.srv.AuthorizeUser(token[1])
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), consts.CtxUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
