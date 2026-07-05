//go:build integration

package mysql_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"auth/internal/adapter/repository/mysql"
	"auth/internal/domain"

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

	schema := `
	CREATE TABLE users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		email VARCHAR(255) NOT NULL UNIQUE,
		password_hash VARCHAR(255) NOT NULL,
		username VARCHAR(255) NOT NULL,
		is_active BOOLEAN DEFAULT TRUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	) ENGINE=InnoDB;
	`
	if _, err := db.ExecContext(ctx, schema); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	cleanup := func() {
		db.Close()
		_ = container.Terminate(ctx)
	}

	return db, cleanup
}

func TestUserRepository_CreateAndFindByEmail(t *testing.T) {
	db, cleanup := setupMySQL(t)
	defer cleanup()

	repo := mysql.NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		Email:    fmt.Sprintf("user%d@test.com", time.Now().UnixNano()),
		Password: "hashed",
		Username: "tester",
	}
	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("create: %v", err)
	}
	if user.ID == "" {
		t.Fatal("expected user id")
	}

	found, err := repo.FindByEmail(ctx, user.Email)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if found == nil || found.Email != user.Email {
		t.Fatalf("unexpected user: %+v", found)
	}
}
