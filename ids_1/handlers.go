package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "irvs/docs"
)

type CreateTaskRequest struct {
	Title string `json:"title" example:"сдать домашку"`
}

type PatchTaskRequest struct {
	Done *bool `json:"done" example:"true"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"task not found"`
}

type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

type API struct {
	store *Store
	log   *slog.Logger
}

func NewRouter(store *Store, log *slog.Logger) http.Handler {
	api := &API{store: store, log: log}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", api.health)
	mux.HandleFunc("GET /tasks", api.listTasks)
	mux.HandleFunc("POST /tasks", api.createTask)
	mux.HandleFunc("GET /tasks/{id}", api.getTask)
	mux.HandleFunc("PATCH /tasks/{id}", api.patchTask)
	mux.HandleFunc("DELETE /tasks/{id}", api.deleteTask)
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	return logRequests(log, mux)
}

func logRequests(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("request", "method", r.Method, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// health проверяет не только процесс, но и соединение с БД —
// его же использует healthcheck в docker compose.
//
//	@Summary		Проверка живости
//	@Description	Проверяет процесс и соединение с БД
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	HealthResponse
//	@Failure		503	{object}	ErrorResponse
//	@Router			/health [get]
func (a *API) health(w http.ResponseWriter, r *http.Request) {
	if err := a.store.Ping(r.Context()); err != nil {
		a.fail(w, http.StatusServiceUnavailable, "database is unavailable")
		return
	}
	writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}

// listTasks
//
//	@Summary	Список задач
//	@Tags		tasks
//	@Produce	json
//	@Success	200	{array}		Task
//	@Failure	500	{object}	ErrorResponse
//	@Router		/tasks [get]
func (a *API) listTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := a.store.List(r.Context())
	if err != nil {
		a.internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

// createTask
//
//	@Summary	Создать задачу
//	@Tags		tasks
//	@Accept		json
//	@Produce	json
//	@Param		request	body		CreateTaskRequest	true	"Новая задача"
//	@Success	201		{object}	Task
//	@Failure	400		{object}	ErrorResponse
//	@Failure	500		{object}	ErrorResponse
//	@Router		/tasks [post]
func (a *API) createTask(w http.ResponseWriter, r *http.Request) {
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.fail(w, http.StatusBadRequest, "invalid json body")
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		a.fail(w, http.StatusBadRequest, "field 'title' is required")
		return
	}

	task, err := a.store.Create(r.Context(), req.Title)
	if err != nil {
		a.internal(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

// getTask
//
//	@Summary	Получить задачу
//	@Tags		tasks
//	@Produce	json
//	@Param		id	path		int	true	"ID задачи"	minimum(1)
//	@Success	200	{object}	Task
//	@Failure	400	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Failure	500	{object}	ErrorResponse
//	@Router		/tasks/{id} [get]
func (a *API) getTask(w http.ResponseWriter, r *http.Request) {
	id, ok := a.pathID(w, r)
	if !ok {
		return
	}
	task, err := a.store.Get(r.Context(), id)
	if err != nil {
		a.storeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

// patchTask
//
//	@Summary	Отметить задачу выполненной или невыполненной
//	@Tags		tasks
//	@Accept		json
//	@Produce	json
//	@Param		id		path		int					true	"ID задачи"	minimum(1)
//	@Param		request	body		PatchTaskRequest	true	"Новое состояние"
//	@Success	200		{object}	Task
//	@Failure	400		{object}	ErrorResponse
//	@Failure	404		{object}	ErrorResponse
//	@Failure	500		{object}	ErrorResponse
//	@Router		/tasks/{id} [patch]
func (a *API) patchTask(w http.ResponseWriter, r *http.Request) {
	id, ok := a.pathID(w, r)
	if !ok {
		return
	}
	var req PatchTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		a.fail(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.Done == nil {
		a.fail(w, http.StatusBadRequest, "field 'done' is required")
		return
	}

	task, err := a.store.SetDone(r.Context(), id, *req.Done)
	if err != nil {
		a.storeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

// deleteTask
//
//	@Summary	Удалить задачу
//	@Tags		tasks
//	@Produce	json
//	@Param		id	path	int	true	"ID задачи"	minimum(1)
//	@Success	204
//	@Failure	400	{object}	ErrorResponse
//	@Failure	404	{object}	ErrorResponse
//	@Failure	500	{object}	ErrorResponse
//	@Router		/tasks/{id} [delete]
func (a *API) deleteTask(w http.ResponseWriter, r *http.Request) {
	id, ok := a.pathID(w, r)
	if !ok {
		return
	}
	if err := a.store.Delete(r.Context(), id); err != nil {
		a.storeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		a.fail(w, http.StatusBadRequest, "id must be a positive integer")
		return 0, false
	}
	return id, true
}

func (a *API) storeError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrNotFound) {
		a.fail(w, http.StatusNotFound, "task not found")
		return
	}
	a.internal(w, err)
}

func (a *API) internal(w http.ResponseWriter, err error) {
	a.log.Error("request failed", "error", err)
	a.fail(w, http.StatusInternalServerError, "internal error")
}

func (a *API) fail(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, ErrorResponse{Error: message})
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}
