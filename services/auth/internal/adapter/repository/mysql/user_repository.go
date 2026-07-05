package mysql

import (
	"auth/internal/domain"
	"context"
	"database/sql"
	"errors"
	"strconv"
)

type UserRepository struct {
	db *sql.DB
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (email, password_hash, username) VALUES (?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query, user.Email, user.Password, user.Username)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = strconv.Itoa(int(id))
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT id, email, password_hash, username, created_at, updated_at FROM users WHERE email = ?`
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.Username, &user.CreatedAt, &user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	//TODO implement me
	panic("implement me")
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}
