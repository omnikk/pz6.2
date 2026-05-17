package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get(RequestIDHeader)
		if rid == "" {
			rid = uuid.NewString()
		}
		w.Header().Set(RequestIDHeader, rid)
		r.Header.Set(RequestIDHeader, rid)
		next.ServeHTTP(w, r)
	})
}
