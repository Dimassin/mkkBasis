package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"tasks/internal/domain"
)

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	query := `INSERT INTO tasks (title, description, status, priority, assignee_id, team_id, created_by, due_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	var assigneeID sql.NullInt64
	if task.AssigneeID != nil {
		assigneeID = sql.NullInt64{Int64: int64(*task.AssigneeID), Valid: true}
	}

	var dueDate sql.NullTime
	if task.DueDate != nil {
		dueDate = sql.NullTime{Time: *task.DueDate, Valid: true}
	}

	result, err := r.db.ExecContext(ctx, query,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		assigneeID,
		task.TeamID,
		task.CreatedBy,
		dueDate,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	task.ID = int(id)
	return nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id int) (*domain.Task, error) {
	query := `SELECT id, title, description, status, priority, assignee_id, team_id, created_by, due_date, created_at, updated_at
		FROM tasks WHERE id = ?`

	task := &domain.Task{}
	var assigneeID sql.NullInt64
	var dueDate sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.Priority,
		&assigneeID,
		&task.TeamID,
		&task.CreatedBy,
		&dueDate,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if assigneeID.Valid {
		val := int(assigneeID.Int64)
		task.AssigneeID = &val
	}
	if dueDate.Valid {
		task.DueDate = &dueDate.Time
	}

	return task, nil
}

func (r *TaskRepository) List(ctx context.Context, filter domain.TaskFilter) ([]*domain.Task, int, error) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "team_id = ?")
	args = append(args, filter.TeamID)

	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.AssigneeID != nil {
		conditions = append(conditions, "assignee_id = ?")
		args = append(args, *filter.AssigneeID)
	}

	where := strings.Join(conditions, " AND ")

	var total int
	var listQuery string
	var listArgs []interface{}

	if filter.Cursor > 0 {
		conditions = append(conditions, "id < ?")
		cursorArgs := append(args, filter.Cursor)
		where = strings.Join(conditions, " AND ")
		listQuery = fmt.Sprintf(`SELECT id, title, description, status, priority, assignee_id, team_id, created_by, due_date, created_at, updated_at
			FROM tasks WHERE %s ORDER BY id DESC LIMIT ?`, where)
		listArgs = append(cursorArgs, filter.Limit)
	} else {
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tasks WHERE %s", where)
		if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
			return nil, 0, err
		}

		offset := (filter.Page - 1) * filter.Limit
		listQuery = fmt.Sprintf(`SELECT id, title, description, status, priority, assignee_id, team_id, created_by, due_date, created_at, updated_at
			FROM tasks WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?`, where)
		listArgs = append(args, filter.Limit, offset)
	}

	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		task := &domain.Task{}
		var assigneeID sql.NullInt64
		var dueDate sql.NullTime

		if err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.Priority,
			&assigneeID,
			&task.TeamID,
			&task.CreatedBy,
			&dueDate,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		if assigneeID.Valid {
			val := int(assigneeID.Int64)
			task.AssigneeID = &val
		}
		if dueDate.Valid {
			task.DueDate = &dueDate.Time
		}

		tasks = append(tasks, task)
	}

	return tasks, total, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *domain.Task) error {
	query := `UPDATE tasks SET title = ?, description = ?, status = ?, priority = ?, assignee_id = ?, due_date = ? WHERE id = ?`

	var assigneeID sql.NullInt64
	if task.AssigneeID != nil {
		assigneeID = sql.NullInt64{Int64: int64(*task.AssigneeID), Valid: true}
	}

	var dueDate sql.NullTime
	if task.DueDate != nil {
		dueDate = sql.NullTime{Time: *task.DueDate, Valid: true}
	}

	_, err := r.db.ExecContext(ctx, query,
		task.Title,
		task.Description,
		task.Status,
		task.Priority,
		assigneeID,
		dueDate,
		task.ID,
	)
	return err
}

func (r *TaskRepository) AddHistory(ctx context.Context, taskID, changedBy int, fieldName string, oldValue, newValue *string) error {
	query := `INSERT INTO task_history (task_id, changed_by, field_name, old_value, new_value) VALUES (?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, taskID, changedBy, fieldName, oldValue, newValue)
	return err
}

func (r *TaskRepository) GetHistory(ctx context.Context, taskID int) ([]*domain.TaskHistory, error) {
	query := `SELECT id, task_id, changed_by, field_name, old_value, new_value, changed_at
		FROM task_history WHERE task_id = ? ORDER BY changed_at ASC`

	rows, err := r.db.QueryContext(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*domain.TaskHistory
	for rows.Next() {
		entry := &domain.TaskHistory{}
		if err := rows.Scan(
			&entry.ID,
			&entry.TaskID,
			&entry.ChangedBy,
			&entry.FieldName,
			&entry.OldValue,
			&entry.NewValue,
			&entry.ChangedAt,
		); err != nil {
			return nil, err
		}
		history = append(history, entry)
	}

	return history, nil
}
