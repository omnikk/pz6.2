package httphandler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/omnikk/pz6/services/auth/internal/service"
)

type Handler struct {
	svc *service.AuthService
}

func New(svc *service.AuthService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/auth/login", h.login)
	mux.HandleFunc("/v1/auth/verify", h.verify)
	return mux
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	CSRFToken   string `json:"csrf_token"` // дублируем для JS-клиентов, которые предпочитают читать из тела
}

type verifyResponse struct {
	Valid   bool   `json:"valid"`
	Subject string `json:"subject,omitempty"`
	Error   string `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	token, ok := h.svc.Login(req.Username, req.Password)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}

	// CSRF-токен — случайный UUID. Это и есть основной секрет для Double Submit Cookie.
	csrfToken := uuid.NewString()

	// session cookie: HttpOnly — JS не может прочитать (защита от кражи через XSS)
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600,
	})

	// csrf cookie: НЕ HttpOnly — клиентский JS должен её читать и слать в X-CSRF-Token
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    csrfToken,
		Path:     "/",
		HttpOnly: false,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600,
	})

	writeJSON(w, http.StatusOK, loginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		CSRFToken:   csrfToken,
	})
}

// verify принимает либо Authorization: Bearer ..., либо session cookie.
// Так старые клиенты (curl с Bearer из ПЗ 5) продолжают работать,
// а новые (через login + cookies) — тоже.
func (h *Handler) verify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	var token string
	// 1) пробуем заголовок
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		token = strings.TrimPrefix(h, "Bearer ")
	}
	// 2) если в заголовке нет — пробуем session cookie
	if token == "" {
		if c, err := r.Cookie("session"); err == nil {
			token = c.Value
		}
	}

	if token == "" {
		writeJSON(w, http.StatusUnauthorized, verifyResponse{Valid: false, Error: "unauthorized"})
		return
	}

	subject, valid := h.svc.Verify(token)
	if !valid {
		writeJSON(w, http.StatusUnauthorized, verifyResponse{Valid: false, Error: "unauthorized"})
		return
	}
	writeJSON(w, http.StatusOK, verifyResponse{Valid: true, Subject: subject})
}
