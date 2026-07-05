package domain

import "time"

type Task struct {
	ID          int
	Title       string
	Description string
	Status      string
	Priority    string
	AssigneeID  *int
	TeamID      int
	CreatedBy   int
	DueDate     *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	TeamID      int    `json:"team_id"`
	AssigneeID  *int   `json:"assignee_id"`
	Priority    string `json:"priority"`
	Status      string `json:"status"`
	DueDate     string `json:"due_date"`
}

type UpdateTaskRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	Priority    *string `json:"priority"`
	AssigneeID  *int    `json:"assignee_id"`
	DueDate     *string `json:"due_date"`
}

type TaskResponse struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	AssigneeID  *int       `json:"assignee_id"`
	TeamID      int        `json:"team_id"`
	CreatedBy   int        `json:"created_by"`
	DueDate     *time.Time `json:"due_date"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type TaskListResponse struct {
	Items      []*TaskResponse `json:"items"`
	Total      int             `json:"total"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	NextCursor int             `json:"next_cursor,omitempty"`
}

type TaskHistory struct {
	ID        int       `json:"id"`
	TaskID    int       `json:"task_id"`
	ChangedBy int       `json:"changed_by"`
	FieldName string    `json:"field_name"`
	OldValue  *string   `json:"old_value"`
	NewValue  *string   `json:"new_value"`
	ChangedAt time.Time `json:"changed_at"`
}

type TaskFilter struct {
	TeamID     int
	Status     string
	AssigneeID *int
	Page       int
	Limit      int
	Cursor     int
}
