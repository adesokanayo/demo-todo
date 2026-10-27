package todo

import "time"

// Todo is one item in the list.
type Todo struct {
	ID        int64     `json:"id" example:"1"`
	Title     string    `json:"title" example:"Buy milk"`
	Done      bool      `json:"done" example:"false"`
	CreatedAt time.Time `json:"created_at" example:"2026-10-27T12:00:00Z"`
}

// CreateTodoRequest is the body for POST /todos.
type CreateTodoRequest struct {
	Title string `json:"title" example:"Buy milk"`
}

// UpdateTodoRequest is a partial update for POST /todos/{id}.
// Omit a field to leave it unchanged.
type UpdateTodoRequest struct {
	Title *string `json:"title" example:"Buy oat milk"`
	Done  *bool   `json:"done" example:"true"`
}

// TodoList is the body of GET /todos.
type TodoList struct {
	Todos []Todo `json:"todos"`
}

// Health is the body of GET /health.
type Health struct {
	Status   string `json:"status" example:"ok"`
	Database string `json:"database" example:"up"`
}

// ErrorBody is returned for failed requests.
type ErrorBody struct {
	Error string `json:"error" example:"title is required"`
}

// ServiceInfo points callers at the API docs.
type ServiceInfo struct {
	Name string `json:"name" example:"demo-todo"`
	Docs string `json:"docs" example:"/swagger/index.html"`
}
