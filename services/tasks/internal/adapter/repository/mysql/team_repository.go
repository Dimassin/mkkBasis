package mysql

import (
	"context"
	"database/sql"
	"errors"
)

type TeamRepository struct {
	db *sql.DB
}

func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) IsMember(ctx context.Context, teamID, userID int) (bool, error) {
	var id int
	err := r.db.QueryRowContext(ctx, `SELECT id FROM team_members WHERE team_id = ? AND user_id = ?`, teamID, userID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
