package middleware

import (
	"encoding/json"
	"net/http"
)

func CSRFProtection(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// safe methods — пропускаем без проверки
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}

		cookie, err := r.Cookie("csrf_token")
		header := r.Header.Get("X-CSRF-Token")

		// все три условия должны выполняться: cookie есть, header есть, они равны.
		if err != nil || cookie.Value == "" || header == "" || cookie.Value != header {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "CSRF token invalid"})
			return
		}

		next.ServeHTTP(w, r)
	})
}
