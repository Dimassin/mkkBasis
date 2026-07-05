package mysql

import (
	"context"
	"database/sql"
	"errors"
	"teams/internal/domain"
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

func (r *TeamRepository) GetUserTeams(ctx context.Context, userID int) ([]*domain.Team, error) {
	query := `
		SELECT t.id, t.name, t.description, t.created_by, t.created_at, t.updated_at
		FROM teams t
		JOIN team_members tm ON t.id = tm.team_id
		WHERE tm.user_id = ?
		ORDER BY t.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*domain.Team
	for rows.Next() {
		var team domain.Team
		err := rows.Scan(
			&team.ID,
			&team.Name,
			&team.Description,
			&team.CreatedBy,
			&team.CreatedAt,
			&team.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		teams = append(teams, &team)
	}

	return teams, nil
}

func (r *TeamRepository) AddMember(ctx context.Context, teamID, userID int, role string) error {
	query := `INSERT INTO team_members (team_id, user_id, role) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, teamID, userID, role)
	return err
}

func (r *TeamRepository) TeamExists(ctx context.Context, teamID int) (bool, error) {
	var id int
	err := r.db.QueryRowContext(ctx, `SELECT id FROM teams WHERE id = ?`, teamID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *TeamRepository) GetMemberRole(ctx context.Context, teamID, userID int) (string, error) {
	var role string
	err := r.db.QueryRowContext(ctx, `SELECT role FROM team_members WHERE team_id = ? AND user_id = ?`, teamID, userID).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return role, nil
}
