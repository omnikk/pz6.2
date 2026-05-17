package httphandler

import (
	"context"
	"encoding/json"
	"errors" // <-- сюда
	"html"
	"log"
	"net/http"
	"strings"

	"github.com/omnikk/pz6/services/tasks/internal/client/authclient"
	"github.com/omnikk/pz6/services/tasks/internal/service"
	"github.com/omnikk/pz6/shared/middleware"
)

type AuthVerifier interface {
	Verify(ctx context.Context, token, sessionCookie, requestID string) (string, error)
}

type Handler struct {
	svc  *service.TaskService
	auth AuthVerifier
}

func NewWithAuth(svc *service.TaskService, auth AuthVerifier) *Handler {
	return &Handler{svc: svc, auth: auth}
}

func (h *Handler) Routes() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/tasks", h.tasksCollection)
	mux.HandleFunc("/v1/tasks/search", h.searchTasks)                     // Р±РµР·РѕРїР°СЃРЅС‹Р№ РїРѕРёСЃРє
	mux.HandleFunc("/v1/tasks/searchvulnerable", h.searchTasksVulnerable) // СѓСЏР·РІРёРјС‹Р№, РґР»СЏ РґРµРјРѕРЅСЃС‚СЂР°С†РёРё
	mux.HandleFunc("/v1/tasks/", h.taskItem)
	return mux
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// internalError вЂ” РµРґРёРЅР°СЏ С‚РѕС‡РєР° РІРѕР·РІСЂР°С‚Р° 500-РѕС€РёР±РѕРє.
// Р”РµС‚Р°Р»СЊ РїР°РґР°РµС‚ РІ Р»РѕРі, РєР»РёРµРЅС‚Сѓ вЂ” РѕР±С‰РµРµ СЃРѕРѕР±С‰РµРЅРёРµ (РЅРµ СЃРІРµС‚РёРј РІРЅСѓС‚СЂСЊ СЃРёСЃС‚РµРјС‹).
func internalError(w http.ResponseWriter, rid string, err error, where string) {
	log.Printf("[%s] %s: %v", rid, where, err)
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
}

func (h *Handler) checkAuth(ctx context.Context, r *http.Request) (string, int) {
	rid := r.Header.Get(middleware.RequestIDHeader)

	// 1) Bearer-токен (как в ПЗ 5)
	authHeader := r.Header.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		// префикса не было — значит это не Bearer
		token = ""
	}

	// 2) session cookie (ПЗ 6)
	var sessionCookie string
	if c, err := r.Cookie("session"); err == nil {
		sessionCookie = c.Value
	}

	if token == "" && sessionCookie == "" {
		return rid, http.StatusUnauthorized
	}

	subject, err := h.auth.Verify(ctx, token, sessionCookie, rid)
	if err != nil {
		if errors.Is(err, authclient.ErrUnauthorized) {
			return rid, http.StatusUnauthorized
		}
		log.Printf("[%s] auth service unavailable: %v", rid, err)
		return rid, http.StatusServiceUnavailable
	}
	log.Printf("[%s] auth ok, subject=%s", rid, subject)
	return rid, 0
}

func (h *Handler) tasksCollection(w http.ResponseWriter, r *http.Request) {
	rid, errStatus := h.checkAuth(r.Context(), r)
	if errStatus != 0 {
		writeJSON(w, errStatus, map[string]string{"error": http.StatusText(errStatus)})
		return
	}
	switch r.Method {
	case http.MethodPost:
		h.createTask(w, r, rid)
	case http.MethodGet:
		h.listTasks(w, r, rid)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) taskItem(w http.ResponseWriter, r *http.Request) {
	rid, errStatus := h.checkAuth(r.Context(), r)
	if errStatus != 0 {
		writeJSON(w, errStatus, map[string]string{"error": http.StatusText(errStatus)})
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/v1/tasks/")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing task id"})
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.getTask(w, r, rid, id)
	case http.MethodPatch:
		h.updateTask(w, r, rid, id)
	case http.MethodDelete:
		h.deleteTask(w, r, rid, id)
	default:
		http.NotFound(w, r)
	}
}

type createRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"`
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request, rid string) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Title == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	// description проходит через html.EscapeString: <script> → &lt;script&gt;
	// При выводе обратно в любой HTML-контекст браузер увидит безопасный текст.
	// (Это компромиссное решение: правильно экранировать на ВЫВОДЕ, а не на ВХОДЕ,
	//  но для учебной работы экранируем заранее.)
	safeDescription := html.EscapeString(req.Description)
	task, err := h.svc.Create(req.Title, safeDescription, req.DueDate)
	if err != nil {
		internalError(w, rid, err, "create task")
		return
	}
	log.Printf("[%s] task created: %s", rid, task.ID)
	writeJSON(w, http.StatusCreated, task)
}

func (h *Handler) listTasks(w http.ResponseWriter, r *http.Request, rid string) {
	tasks, err := h.svc.List()
	if err != nil {
		internalError(w, rid, err, "list tasks")
		return
	}
	log.Printf("[%s] list tasks: %d items", rid, len(tasks))
	writeJSON(w, http.StatusOK, tasks)
}

func (h *Handler) getTask(w http.ResponseWriter, r *http.Request, rid, id string) {
	task, ok, err := h.svc.Get(id)
	if err != nil {
		internalError(w, rid, err, "get task")
		return
	}
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}
	writeJSON(w, http.StatusOK, task)
}

type updateRequest struct {
	Title *string `json:"title"`
	Done  *bool   `json:"done"`
}

func (h *Handler) updateTask(w http.ResponseWriter, r *http.Request, rid, id string) {
	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	task, ok, err := h.svc.Update(id, req.Title, req.Done)
	if err != nil {
		internalError(w, rid, err, "update task")
		return
	}
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *Handler) deleteTask(w http.ResponseWriter, r *http.Request, rid, id string) {
	ok, err := h.svc.Delete(id)
	if err != nil {
		internalError(w, rid, err, "delete task")
		return
	}
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- РїРѕРёСЃРє ----

func (h *Handler) searchTasks(w http.ResponseWriter, r *http.Request) {
	rid, errStatus := h.checkAuth(r.Context(), r)
	if errStatus != 0 {
		writeJSON(w, errStatus, map[string]string{"error": http.StatusText(errStatus)})
		return
	}
	title := r.URL.Query().Get("title")
	tasks, err := h.svc.Search(title)
	if err != nil {
		internalError(w, rid, err, "search")
		return
	}
	log.Printf("[%s] search title=%q: %d items", rid, title, len(tasks))
	writeJSON(w, http.StatusOK, tasks)
}

func (h *Handler) searchTasksVulnerable(w http.ResponseWriter, r *http.Request) {
	rid, errStatus := h.checkAuth(r.Context(), r)
	if errStatus != 0 {
		writeJSON(w, errStatus, map[string]string{"error": http.StatusText(errStatus)})
		return
	}
	title := r.URL.Query().Get("title")
	tasks, err := h.svc.SearchVulnerable(title)
	if err != nil {
		internalError(w, rid, err, "search vulnerable")
		return
	}
	log.Printf("[%s] VULNERABLE search title=%q: %d items", rid, title, len(tasks))
	writeJSON(w, http.StatusOK, tasks)
}
