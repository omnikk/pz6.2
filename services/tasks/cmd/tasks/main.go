package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/omnikk/pz6/services/tasks/internal/client/authclient"
	httphandler "github.com/omnikk/pz6/services/tasks/internal/http"
	"github.com/omnikk/pz6/services/tasks/internal/repository"
	"github.com/omnikk/pz6/services/tasks/internal/service"
	"github.com/omnikk/pz6/shared/middleware"
)

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	port := env("TASKS_PORT", "8082")

	// DSN РґР»СЏ PostgreSQL СЃРѕР±РёСЂР°РµРј РёР· РїРµСЂРµРјРµРЅРЅС‹С… РѕРєСЂСѓР¶РµРЅРёСЏ.
	// Р’ docker-compose РѕРЅРё РїСЂРѕРєРёРЅСѓС‚С‹, Р»РѕРєР°Р»СЊРЅРѕ РјРѕР¶РЅРѕ РїРѕСЃС‚Р°РІРёС‚СЊ .env РёР»Рё СЌРєСЃРїРѕСЂС‚РёСЂРѕРІР°С‚СЊ РІСЂСѓС‡РЅСѓСЋ.
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		env("DB_HOST", "localhost"),
		env("DB_PORT", "5432"),
		env("DB_USER", "tasks_user"),
		env("DB_PASSWORD", "tasks_pass"),
		env("DB_NAME", "tasks_db"),
		env("DB_SSLMODE", "disable"),
	)

	repo, err := repository.NewPostgres(dsn)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer repo.Close()
	log.Printf("connected to postgres: %s@%s/%s",
		env("DB_USER", "tasks_user"),
		env("DB_HOST", "localhost"),
		env("DB_NAME", "tasks_db"))

	authURL := env("AUTH_BASE_URL", "http://localhost:8081")
	auth := authclient.New(authURL)

	svc := service.New(repo)
	handler := httphandler.NewWithAuth(svc, auth)

	// Порядок middleware важен (читать снизу вверх — наружный применяется последним):
	// 1) RequestID  — присваивает X-Request-ID, должен быть самым внешним для всех логов
	// 2) Logging    — пишет access-log с request-id
	// 3) Security   — добавляет CSP/X-Frame-Options ко ВСЕМ ответам (даже к 401/403)
	// 4) CSRF       — проверяет токен для POST/PATCH/DELETE
	var mux http.Handler = handler.Routes()
	mux = middleware.CSRFProtection(mux)
	mux = middleware.SecurityHeaders(mux)
	mux = middleware.Logging(mux)
	mux = middleware.RequestID(mux)

	log.Printf("Tasks service starting on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("Tasks service failed: %v", err)
	}
}
