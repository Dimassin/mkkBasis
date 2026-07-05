//go:build integration

package mysql_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"tasks/internal/adapter/repository/mysql"
	"tasks/internal/domain"

	_ "github.com/go-sql-driver/mysql"
	"github.com/testcontainers/testcontainers-go"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupMySQL(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	ctx := context.Background()

	container, err := tcmysql.Run(ctx, "mysql:8.0",
		tcmysql.WithDatabase("test"),
		tcmysql.WithUsername("root"),
		tcmysql.WithPassword("pass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("port: 3306  MySQL Community Server").WithStartupTimeout(2*time.Minute),
		),
	)
	if err != nil {
		t.Fatalf("start mysql: %v", err)
	}

	dsn, err := container.ConnectionString(ctx, "parseTime=true")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	statements := []string{
		`CREATE TABLE users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			email VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			username VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE teams (
			id INT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			created_by INT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE team_members (
			id INT AUTO_INCREMENT PRIMARY KEY,
			team_id INT NOT NULL,
			user_id INT NOT NULL,
			role ENUM('owner', 'admin', 'member') NOT NULL DEFAULT 'member',
			joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE KEY unique_team_user (team_id, user_id)
		)`,
		`CREATE TABLE tasks (
			id INT AUTO_INCREMENT PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			status ENUM('new', 'in_progress', 'done', 'cancelled') NOT NULL DEFAULT 'new',
			priority ENUM('low', 'medium', 'high', 'critical') NOT NULL DEFAULT 'medium',
			assignee_id INT,
			team_id INT NOT NULL,
			created_by INT NOT NULL,
			due_date DATE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE task_history (
			id INT AUTO_INCREMENT PRIMARY KEY,
			task_id INT NOT NULL,
			changed_by INT NOT NULL,
			field_name VARCHAR(100) NOT NULL,
			old_value TEXT,
			new_value TEXT,
			changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, stmt := range statements {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("migrate: %v", err)
		}
	}

	cleanup := func() {
		db.Close()
		_ = container.Terminate(ctx)
	}

	return db, cleanup
}

func TestTaskRepository_CreateListAndHistory(t *testing.T) {
	db, cleanup := setupMySQL(t)
	defer cleanup()

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `INSERT INTO users (email, password_hash, username) VALUES ('u@test.com', 'hash', 'user')`)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	_, err = db.ExecContext(ctx, `INSERT INTO teams (name, description, created_by) VALUES ('Dev', 'team', 1)`)
	if err != nil {
		t.Fatalf("seed team: %v", err)
	}
	_, err = db.ExecContext(ctx, `INSERT INTO team_members (team_id, user_id, role) VALUES (1, 1, 'owner')`)
	if err != nil {
		t.Fatalf("seed member: %v", err)
	}

	repo := mysql.NewTaskRepository(db)
	task := &domain.Task{
		Title:     "Integration task",
		Status:    "new",
		Priority:  "medium",
		TeamID:    1,
		CreatedBy: 1,
	}
	if err := repo.Create(ctx, task); err != nil {
		t.Fatalf("create: %v", err)
	}

	oldVal := "new"
	newVal := "in_progress"
	if err := repo.AddHistory(ctx, task.ID, 1, "status", &oldVal, &newVal); err != nil {
		t.Fatalf("history: %v", err)
	}

	tasks, total, err := repo.List(ctx, domain.TaskFilter{TeamID: 1, Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 || len(tasks) != 1 {
		t.Fatalf("expected 1 task, got total=%d len=%d", total, len(tasks))
	}

	history, err := repo.GetHistory(ctx, task.ID)
	if err != nil {
		t.Fatalf("get history: %v", err)
	}
	if len(history) != 1 || history[0].FieldName != "status" {
		t.Fatalf("unexpected history: %+v", history)
	}
}

func TestTaskRepository_CursorPagination(t *testing.T) {
	db, cleanup := setupMySQL(t)
	defer cleanup()

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `INSERT INTO users (email, password_hash, username) VALUES ('u2@test.com', 'hash', 'user')`)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	_, err = db.ExecContext(ctx, `INSERT INTO teams (name, description, created_by) VALUES ('QA', 'team', 1)`)
	if err != nil {
		t.Fatalf("seed team: %v", err)
	}

	repo := mysql.NewTaskRepository(db)
	for i := 1; i <= 3; i++ {
		task := &domain.Task{
			Title:     "Task",
			Status:    "new",
			Priority:  "low",
			TeamID:    1,
			CreatedBy: 1,
		}
		if err := repo.Create(ctx, task); err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
	}

	firstPage, _, err := repo.List(ctx, domain.TaskFilter{TeamID: 1, Limit: 2, Cursor: 0, Page: 1})
	if err != nil {
		t.Fatalf("list page: %v", err)
	}
	if len(firstPage) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(firstPage))
	}

	cursor := firstPage[len(firstPage)-1].ID
	secondPage, _, err := repo.List(ctx, domain.TaskFilter{TeamID: 1, Limit: 2, Cursor: cursor})
	if err != nil {
		t.Fatalf("list cursor: %v", err)
	}
	if len(secondPage) != 1 {
		t.Fatalf("expected 1 task on cursor page, got %d", len(secondPage))
	}
}
