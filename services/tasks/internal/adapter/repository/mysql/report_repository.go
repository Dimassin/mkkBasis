package mysql

import (
	"context"
	"database/sql"
	"tasks/internal/domain"
)

type ReportRepository struct {
	db *sql.DB
}

func NewReportRepository(db *sql.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

func (r *ReportRepository) GetTeamStats(ctx context.Context) ([]*domain.TeamStats, error) {
	query := `
		SELECT
			t.id,
			t.name,
			COUNT(DISTINCT tm.user_id) AS members_count,
			COUNT(DISTINCT CASE
				WHEN tk.status = 'done' AND tk.updated_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
				THEN tk.id
			END) AS done_tasks_last_7_days
		FROM teams t
		LEFT JOIN team_members tm ON tm.team_id = t.id
		LEFT JOIN tasks tk ON tk.team_id = t.id
		GROUP BY t.id, t.name
		ORDER BY t.id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*domain.TeamStats
	for rows.Next() {
		item := &domain.TeamStats{}
		if err := rows.Scan(&item.TeamID, &item.TeamName, &item.MembersCount, &item.DoneTasksLast7Days); err != nil {
			return nil, err
		}
		stats = append(stats, item)
	}

	return stats, nil
}

func (r *ReportRepository) GetTopCreatorsByTeam(ctx context.Context) ([]*domain.TopCreator, error) {
	query := `
		SELECT team_id, user_id, username, tasks_created, rank_in_team
		FROM (
			SELECT
				tk.team_id,
				tk.created_by AS user_id,
				u.username,
				COUNT(*) AS tasks_created,
				ROW_NUMBER() OVER (
					PARTITION BY tk.team_id
					ORDER BY COUNT(*) DESC, tk.created_by ASC
				) AS rank_in_team
			FROM tasks tk
			INNER JOIN users u ON u.id = tk.created_by
			WHERE tk.created_at >= DATE_SUB(NOW(), INTERVAL 1 MONTH)
			GROUP BY tk.team_id, tk.created_by, u.username
		) ranked
		WHERE rank_in_team <= 3
		ORDER BY team_id, rank_in_team
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var creators []*domain.TopCreator
	for rows.Next() {
		item := &domain.TopCreator{}
		if err := rows.Scan(&item.TeamID, &item.UserID, &item.Username, &item.TasksCreated, &item.Rank); err != nil {
			return nil, err
		}
		creators = append(creators, item)
	}

	return creators, nil
}

func (r *ReportRepository) GetInvalidAssigneeTasks(ctx context.Context) ([]*domain.InvalidAssigneeTask, error) {
	query := `
		SELECT tk.id, tk.title, tk.team_id, tk.assignee_id
		FROM tasks tk
		WHERE tk.assignee_id IS NOT NULL
		AND NOT EXISTS (
			SELECT 1
			FROM team_members tm
			WHERE tm.team_id = tk.team_id AND tm.user_id = tk.assignee_id
		)
		ORDER BY tk.id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.InvalidAssigneeTask
	for rows.Next() {
		item := &domain.InvalidAssigneeTask{}
		if err := rows.Scan(&item.TaskID, &item.Title, &item.TeamID, &item.AssigneeID); err != nil {
			return nil, err
		}
		tasks = append(tasks, item)
	}

	return tasks, nil
}
