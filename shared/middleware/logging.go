package middleware

import (
	"log"
	"net/http"
	"time"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rid := r.Header.Get(RequestIDHeader)
		log.Printf("[%s] --> %s %s", rid, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("[%s] <-- %s %s (%s)", rid, r.Method, r.URL.Path, time.Since(start))
	})
}
