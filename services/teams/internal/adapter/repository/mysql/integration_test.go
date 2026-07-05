//go:build integration

package mysql_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"teams/internal/adapter/repository/mysql"
	"teams/internal/domain"

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

func TestTeamRepository_CreateAndAddMember(t *testing.T) {
	db, cleanup := setupMySQL(t)
	defer cleanup()

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `INSERT INTO users (email, password_hash, username) VALUES ('o@test.com', 'hash', 'owner')`)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	repo := mysql.NewTeamRepository(db)
	team := &domain.Team{Name: "Dev", Description: "team", CreatedBy: 1}
	if err := repo.Create(ctx, team); err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := repo.AddMember(ctx, team.ID, 1, "owner"); err != nil {
		t.Fatalf("add member: %v", err)
	}

	role, err := repo.GetMemberRole(ctx, team.ID, 1)
	if err != nil {
		t.Fatalf("get role: %v", err)
	}
	if role != "owner" {
		t.Fatalf("expected owner, got %s", role)
	}
}
