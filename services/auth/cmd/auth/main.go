package main

import (
	"log"
	"net/http"
	"os"

	httphandler "github.com/omnikk/pz6/services/auth/internal/http"
	"github.com/omnikk/pz6/services/auth/internal/service"
	"github.com/omnikk/pz6/shared/middleware"
)

func main() {
	httpPort := os.Getenv("AUTH_PORT")
	if httpPort == "" {
		httpPort = "8081"
	}

	svc := service.New()
	handler := httphandler.New(svc)

	// auth-сервису CSRF не нужен (login сам выдаёт токен), а заголовки безопасности — да
	var mux http.Handler = handler.Routes()
	mux = middleware.SecurityHeaders(mux)
	mux = middleware.Logging(mux)
	mux = middleware.RequestID(mux)

	log.Printf("Auth HTTP server starting on :%s", httpPort)
	if err := http.ListenAndServe(":"+httpPort, mux); err != nil {
		log.Fatalf("Auth HTTP server failed: %v", err)
	}
}
