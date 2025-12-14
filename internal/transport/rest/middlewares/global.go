package middlewares

import "net/http"

//TODO: сделать логирование и обработку исключений

func GlobalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Accept", "application/json")
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
