package domain

type TeamStats struct {
	TeamID             int    `json:"team_id"`
	TeamName           string `json:"team_name"`
	MembersCount       int    `json:"members_count"`
	DoneTasksLast7Days int    `json:"done_tasks_last_7_days"`
}

type TopCreator struct {
	TeamID       int    `json:"team_id"`
	UserID       int    `json:"user_id"`
	Username     string `json:"username"`
	TasksCreated int    `json:"tasks_created"`
	Rank         int    `json:"rank"`
}

type InvalidAssigneeTask struct {
	TaskID     int    `json:"task_id"`
	Title      string `json:"title"`
	TeamID     int    `json:"team_id"`
	AssigneeID int    `json:"assignee_id"`
}
