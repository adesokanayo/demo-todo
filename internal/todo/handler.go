package todo

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

func Register(mux *http.ServeMux, store Repository) {
	h := &handler{store: store}
	mux.HandleFunc("GET /{$}", h.serviceInfo)
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /todos", h.listTodos)
	mux.HandleFunc("POST /todos", h.createTodo)
	mux.HandleFunc("GET /todos/{id}", h.getTodo)
	mux.HandleFunc("POST /todos/{id}", h.updateTodo)
}

type handler struct {
	store Repository
}

// serviceInfo godoc
// @Summary     API index
// @Tags        meta
// @Produce     json
// @Success     200 {object} ServiceInfo
// @Router      / [get]
func (h *handler) serviceInfo(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, ServiceInfo{
		Name: "demo-todo",
		Docs: "/swagger/index.html",
	})
}

// health godoc
// @Summary     Health check
// @Description Reports whether the API can reach Postgres.
// @Tags        meta
// @Produce     json
// @Success     200 {object} Health
// @Failure     503 {object} Health
// @Router      /health [get]
func (h *handler) health(w http.ResponseWriter, r *http.Request) {
	if err := h.store.Ping(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, Health{Status: "down", Database: "down"})
		return
	}
	writeJSON(w, http.StatusOK, Health{Status: "ok", Database: "up"})
}

// listTodos godoc
// @Summary     List todos
// @Tags        todos
// @Produce     json
// @Success     200 {object} TodoList
// @Failure     500 {object} ErrorBody
// @Router      /todos [get]
func (h *handler) listTodos(w http.ResponseWriter, r *http.Request) {
	todos, err := h.store.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not list todos")
		return
	}
	writeJSON(w, http.StatusOK, TodoList{Todos: todos})
}

// createTodo godoc
// @Summary     Create a todo
// @Tags        todos
// @Accept      json
// @Produce     json
// @Param       todo body CreateTodoRequest true "Todo to create"
// @Success     201 {object} Todo
// @Failure     400 {object} ErrorBody
// @Failure     500 {object} ErrorBody
// @Router      /todos [post]
func (h *handler) createTodo(w http.ResponseWriter, r *http.Request) {
	var body CreateTodoRequest
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	title := strings.TrimSpace(body.Title)
	if title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	if len(title) > 200 {
		writeError(w, http.StatusBadRequest, "title must be 200 characters or fewer")
		return
	}
	item, err := h.store.Create(r.Context(), title)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not create todo")
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

// getTodo godoc
// @Summary     Get a todo
// @Tags        todos
// @Produce     json
// @Param       id path int true "Todo id"
// @Success     200 {object} Todo
// @Failure     400 {object} ErrorBody
// @Failure     404 {object} ErrorBody
// @Failure     500 {object} ErrorBody
// @Router      /todos/{id} [get]
func (h *handler) getTodo(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	item, err := h.store.Get(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "todo not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not get todo")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

// updateTodo godoc
// @Summary     Update a todo
// @Description Changes the title, the done flag, or both. Send only the fields you want to change.
// @Tags        todos
// @Accept      json
// @Produce     json
// @Param       id path int true "Todo id"
// @Param       todo body UpdateTodoRequest true "Fields to change"
// @Success     200 {object} Todo
// @Failure     400 {object} ErrorBody
// @Failure     404 {object} ErrorBody
// @Failure     500 {object} ErrorBody
// @Router      /todos/{id} [post]
func (h *handler) updateTodo(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	var body UpdateTodoRequest
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if body.Title == nil && body.Done == nil {
		writeError(w, http.StatusBadRequest, "title or done is required")
		return
	}
	if body.Title != nil {
		trimmed := strings.TrimSpace(*body.Title)
		if trimmed == "" {
			writeError(w, http.StatusBadRequest, "title cannot be empty")
			return
		}
		if len(trimmed) > 200 {
			writeError(w, http.StatusBadRequest, "title must be 200 characters or fewer")
			return
		}
		body.Title = &trimmed
	}
	item, err := h.store.Update(r.Context(), id, body)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "todo not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not update todo")
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorBody{Error: message})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
