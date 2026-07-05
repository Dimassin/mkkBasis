package mysql

import (
	"auth/internal/domain"
	"context"
	"database/sql"
)

type TeamRepository struct {
	db *sql.DB
}

func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Create(ctx context.Context, team *domain.Team) error {
	query := `INSERT INTO teams (name, description, created_by) VALUES (?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query, team.Name, team.Description, team.CreatedBy)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	team.ID = int(id)
	return nil
}

func (r *TeamRepository) AddMember(ctx context.Context, teamID, userID int, role string) error {
	query := `INSERT INTO team_members (team_id, user_id, role) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, teamID, userID, role)
	return err
}
