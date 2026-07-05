package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"tasks/internal/domain"
)

type mockTaskRepo struct {
	tasks      map[int]*domain.Task
	nextID     int
	createErr  error
	listResult []*domain.Task
	listTotal  int
	listErr    error
	getResult  *domain.Task
	history    []*domain.TaskHistory
}

func (m *mockTaskRepo) Create(ctx context.Context, task *domain.Task) error {
	if m.createErr != nil {
		return m.createErr
	}
	if m.tasks == nil {
		m.tasks = make(map[int]*domain.Task)
	}
	m.nextID++
	task.ID = m.nextID
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskRepo) GetByID(ctx context.Context, id int) (*domain.Task, error) {
	if m.tasks != nil {
		if task, ok := m.tasks[id]; ok {
			return task, nil
		}
	}
	if m.getResult != nil {
		return m.getResult, nil
	}
	return nil, nil
}

func (m *mockTaskRepo) List(ctx context.Context, filter domain.TaskFilter) ([]*domain.Task, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	return m.listResult, m.listTotal, nil
}

func (m *mockTaskRepo) Update(ctx context.Context, task *domain.Task) error {
	if m.tasks == nil {
		m.tasks = make(map[int]*domain.Task)
	}
	m.tasks[task.ID] = task
	return nil
}

func (m *mockTaskRepo) AddHistory(ctx context.Context, taskID, changedBy int, fieldName string, oldValue, newValue *string) error {
	m.history = append(m.history, &domain.TaskHistory{
		TaskID:    taskID,
		ChangedBy: changedBy,
		FieldName: fieldName,
		OldValue:  oldValue,
		NewValue:  newValue,
		ChangedAt: time.Now(),
	})
	return nil
}

func (m *mockTaskRepo) GetHistory(ctx context.Context, taskID int) ([]*domain.TaskHistory, error) {
	return m.history, nil
}

type mockTeamRepo struct {
	members map[string]bool
}

func teamMemberKey(teamID, userID int) string {
	return fmt.Sprintf("%d:%d", teamID, userID)
}

func (m *mockTeamRepo) IsMember(ctx context.Context, teamID, userID int) (bool, error) {
	if m.members == nil {
		return false, nil
	}
	return m.members[teamMemberKey(teamID, userID)], nil
}

type mockCache struct {
	data map[string]*domain.TaskListResponse
}

func (m *mockCache) Get(ctx context.Context, key string) (*domain.TaskListResponse, error) {
	if m.data == nil {
		return nil, nil
	}
	return m.data[key], nil
}

func (m *mockCache) Set(ctx context.Context, key string, value *domain.TaskListResponse, ttl time.Duration) error {
	if m.data == nil {
		m.data = make(map[string]*domain.TaskListResponse)
	}
	m.data[key] = value
	return nil
}

func (m *mockCache) DeleteByTeam(ctx context.Context, teamID int) error {
	return nil
}

func TestCreateTask_Success(t *testing.T) {
	teamRepo := &mockTeamRepo{members: map[string]bool{teamMemberKey(1, 1): true}}
	taskRepo := &mockTaskRepo{}
	uc := NewTaskUsecase(taskRepo, teamRepo, nil)

	resp, err := uc.CreateTask(context.Background(), 1, &domain.CreateTaskRequest{
		Title:  "Task",
		TeamID: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Title != "Task" || resp.Status != "new" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestCreateTask_NotMember(t *testing.T) {
	uc := NewTaskUsecase(&mockTaskRepo{}, &mockTeamRepo{}, nil)

	_, err := uc.CreateTask(context.Background(), 1, &domain.CreateTaskRequest{Title: "Task", TeamID: 1})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestCreateTask_InvalidStatus(t *testing.T) {
	teamRepo := &mockTeamRepo{members: map[string]bool{teamMemberKey(1, 1): true}}
	uc := NewTaskUsecase(&mockTaskRepo{}, teamRepo, nil)

	_, err := uc.CreateTask(context.Background(), 1, &domain.CreateTaskRequest{
		Title:  "Task",
		TeamID: 1,
		Status: "invalid",
	})
	if !errors.Is(err, domain.ErrInvalidStatus) {
		t.Fatalf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestListTasks_FromRepo(t *testing.T) {
	teamRepo := &mockTeamRepo{members: map[string]bool{teamMemberKey(1, 1): true}}
	taskRepo := &mockTaskRepo{
		listResult: []*domain.Task{{ID: 1, Title: "DB", TeamID: 1, Status: "new", Priority: "medium", CreatedBy: 1}},
		listTotal:  1,
	}
	uc := NewTaskUsecase(taskRepo, teamRepo, nil)

	resp, err := uc.ListTasks(context.Background(), 1, domain.TaskFilter{TeamID: 1, Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Title != "DB" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestListTasks_TeamNotFound(t *testing.T) {
	uc := NewTaskUsecase(&mockTaskRepo{}, &mockTeamRepo{}, nil)

	_, err := uc.ListTasks(context.Background(), 1, domain.TaskFilter{})
	if !errors.Is(err, domain.ErrTeamNotFound) {
		t.Fatalf("expected ErrTeamNotFound, got %v", err)
	}
}

func TestListTasks_InvalidStatus(t *testing.T) {
	teamRepo := &mockTeamRepo{members: map[string]bool{teamMemberKey(1, 1): true}}
	uc := NewTaskUsecase(&mockTaskRepo{}, teamRepo, nil)

	_, err := uc.ListTasks(context.Background(), 1, domain.TaskFilter{TeamID: 1, Status: "bad", Page: 1, Limit: 20})
	if !errors.Is(err, domain.ErrInvalidStatus) {
		t.Fatalf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestListTasks_CursorNextPage(t *testing.T) {
	teamRepo := &mockTeamRepo{members: map[string]bool{teamMemberKey(1, 1): true}}
	taskRepo := &mockTaskRepo{
		listResult: []*domain.Task{
			{ID: 10, Title: "A", TeamID: 1, Status: "new", Priority: "low"},
			{ID: 9, Title: "B", TeamID: 1, Status: "new", Priority: "low"},
		},
	}
	uc := NewTaskUsecase(taskRepo, teamRepo, nil)

	resp, err := uc.ListTasks(context.Background(), 1, domain.TaskFilter{TeamID: 1, Cursor: 11, Limit: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.NextCursor != 9 {
		t.Fatalf("expected next cursor 9, got %d", resp.NextCursor)
	}
}

func TestCreateTask_AssigneeNotMember(t *testing.T) {
	assignee := 2
	teamRepo := &mockTeamRepo{members: map[string]bool{teamMemberKey(1, 1): true}}
	uc := NewTaskUsecase(&mockTaskRepo{}, teamRepo, nil)

	_, err := uc.CreateTask(context.Background(), 1, &domain.CreateTaskRequest{
		Title:      "Task",
		TeamID:     1,
		AssigneeID: &assignee,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestCreateTask_InvalidPriority(t *testing.T) {
	teamRepo := &mockTeamRepo{members: map[string]bool{teamMemberKey(1, 1): true}}
	uc := NewTaskUsecase(&mockTaskRepo{}, teamRepo, nil)

	_, err := uc.CreateTask(context.Background(), 1, &domain.CreateTaskRequest{
		Title:    "Task",
		TeamID:   1,
		Priority: "bad",
	})
	if !errors.Is(err, domain.ErrInvalidPriority) {
		t.Fatalf("expected ErrInvalidPriority, got %v", err)
	}
}

func TestUpdateTask_Forbidden(t *testing.T) {
	taskRepo := &mockTaskRepo{
		tasks: map[int]*domain.Task{1: {ID: 1, TeamID: 1, Status: "new"}},
	}
	uc := NewTaskUsecase(taskRepo, &mockTeamRepo{}, nil)

	_, err := uc.UpdateTask(context.Background(), 1, 1, &domain.UpdateTaskRequest{})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestUpdateTask_TitleAndPriority(t *testing.T) {
	title := "New title"
	priority := "high"
	teamRepo := &mockTeamRepo{members: map[string]bool{teamMemberKey(1, 1): true}}
	taskRepo := &mockTaskRepo{
		tasks: map[int]*domain.Task{
			1: {ID: 1, Title: "Old", Status: "new", Priority: "low", TeamID: 1},
		},
	}
	uc := NewTaskUsecase(taskRepo, teamRepo, nil)

	resp, err := uc.UpdateTask(context.Background(), 1, 1, &domain.UpdateTaskRequest{
		Title:    &title,
		Priority: &priority,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Title != "New title" || resp.Priority != "high" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestGetTaskHistory_Forbidden(t *testing.T) {
	taskRepo := &mockTaskRepo{
		tasks: map[int]*domain.Task{1: {ID: 1, TeamID: 1}},
	}
	uc := NewTaskUsecase(taskRepo, &mockTeamRepo{}, nil)

	_, err := uc.GetTaskHistory(context.Background(), 1, 1)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestGetTaskHistory_NotFound(t *testing.T) {
	teamRepo := &mockTeamRepo{members: map[string]bool{teamMemberKey(1, 1): true}}
	uc := NewTaskUsecase(&mockTaskRepo{}, teamRepo, nil)

	_, err := uc.GetTaskHistory(context.Background(), 1, 99)
	if !errors.Is(err, domain.ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestCreateTask_WithDueDate(t *testing.T) {
	teamRepo := &mockTeamRepo{members: map[string]bool{teamMemberKey(1, 1): true}}
	taskRepo := &mockTaskRepo{}
	uc := NewTaskUsecase(taskRepo, teamRepo, nil)

	resp, err := uc.CreateTask(context.Background(), 1, &domain.CreateTaskRequest{
		Title:   "Task",
		TeamID:  1,
		DueDate: "2026-07-01",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.DueDate == nil {
		t.Fatal("expected due date")
	}
}

func TestUpdateTask_DescriptionAssigneeDueDate(t *testing.T) {
	desc := "updated"
	assignee := 1
	due := "2026-08-01"
	zero := 0
	teamRepo := &mockTeamRepo{members: map[string]bool{teamMemberKey(1, 1): true}}
	taskRepo := &mockTaskRepo{
		tasks: map[int]*domain.Task{
			1: {
				ID:          1,
				Title:       "Task",
				Description: "old",
				Status:      "new",
				Priority:    "low",
				TeamID:      1,
				AssigneeID:  &assignee,
			},
		},
	}
	uc := NewTaskUsecase(taskRepo, teamRepo, &mockCache{})

	resp, err := uc.UpdateTask(context.Background(), 1, 1, &domain.UpdateTaskRequest{
		Description: &desc,
		DueDate:     &due,
		AssigneeID:  &zero,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Description != "updated" || resp.AssigneeID != nil {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestUpdateTask_InvalidStatus(t *testing.T) {
	bad := "bad"
	teamRepo := &mockTeamRepo{members: map[string]bool{teamMemberKey(1, 1): true}}
	taskRepo := &mockTaskRepo{
		tasks: map[int]*domain.Task{1: {ID: 1, TeamID: 1, Status: "new"}},
	}
	uc := NewTaskUsecase(taskRepo, teamRepo, nil)

	_, err := uc.UpdateTask(context.Background(), 1, 1, &domain.UpdateTaskRequest{Status: &bad})
	if !errors.Is(err, domain.ErrInvalidStatus) {
		t.Fatalf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestUpdateTask_AssigneeNotMember(t *testing.T) {
	assignee := 2
	teamRepo := &mockTeamRepo{members: map[string]bool{teamMemberKey(1, 1): true}}
	taskRepo := &mockTaskRepo{
		tasks: map[int]*domain.Task{1: {ID: 1, TeamID: 1, Status: "new"}},
	}
	uc := NewTaskUsecase(taskRepo, teamRepo, nil)

	_, err := uc.UpdateTask(context.Background(), 1, 1, &domain.UpdateTaskRequest{AssigneeID: &assignee})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestListTasks_TodoStatus(t *testing.T) {
	teamRepo := &mockTeamRepo{members: map[string]bool{teamMemberKey(1, 1): true}}
	taskRepo := &mockTaskRepo{listResult: []*domain.Task{}, listTotal: 0}
	uc := NewTaskUsecase(taskRepo, teamRepo, nil)

	_, err := uc.ListTasks(context.Background(), 1, domain.TaskFilter{TeamID: 1, Status: "todo", Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateTask_TodoStatus(t *testing.T) {
	status := "todo"
	teamRepo := &mockTeamRepo{members: map[string]bool{teamMemberKey(1, 1): true}}
	taskRepo := &mockTaskRepo{
		tasks: map[int]*domain.Task{1: {ID: 1, TeamID: 1, Status: "new", Priority: "low"}},
	}
	uc := NewTaskUsecase(taskRepo, teamRepo, nil)

	resp, err := uc.UpdateTask(context.Background(), 1, 1, &domain.UpdateTaskRequest{Status: &status})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "new" {
		t.Fatalf("expected new status, got %s", resp.Status)
	}
}

func TestCreateTask_InvalidDueDate(t *testing.T) {
	teamRepo := &mockTeamRepo{members: map[string]bool{teamMemberKey(1, 1): true}}
	uc := NewTaskUsecase(&mockTaskRepo{}, teamRepo, nil)

	_, err := uc.CreateTask(context.Background(), 1, &domain.CreateTaskRequest{
		Title:   "Task",
		TeamID:  1,
		DueDate: "bad-date",
	})
	if !errors.Is(err, domain.ErrInternalServer) {
		t.Fatalf("expected ErrInternalServer, got %v", err)
	}
}

func TestListTasks_CacheHit(t *testing.T) {
	teamRepo := &mockTeamRepo{members: map[string]bool{teamMemberKey(1, 1): true}}
	cache := &mockCache{
		data: map[string]*domain.TaskListResponse{
			buildTaskListCacheKey(domain.TaskFilter{TeamID: 1, Page: 1, Limit: 20, Status: "all"}): {
				Items: []*domain.TaskResponse{{ID: 99, Title: "cached"}},
				Total: 1,
				Page:  1,
				Limit: 20,
			},
		},
	}
	uc := NewTaskUsecase(&mockTaskRepo{}, teamRepo, cache)

	resp, err := uc.ListTasks(context.Background(), 1, domain.TaskFilter{TeamID: 1, Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].Title != "cached" {
		t.Fatalf("expected cached result, got %+v", resp)
	}
}

func TestListTasks_Forbidden(t *testing.T) {
	uc := NewTaskUsecase(&mockTaskRepo{}, &mockTeamRepo{}, nil)

	_, err := uc.ListTasks(context.Background(), 1, domain.TaskFilter{TeamID: 1, Page: 1, Limit: 20})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestUpdateTask_Success(t *testing.T) {
	status := "in_progress"
	teamRepo := &mockTeamRepo{members: map[string]bool{teamMemberKey(1, 1): true}}
	taskRepo := &mockTaskRepo{
		tasks: map[int]*domain.Task{
			1: {ID: 1, Title: "Old", Status: "new", TeamID: 1},
		},
	}
	uc := NewTaskUsecase(taskRepo, teamRepo, nil)

	resp, err := uc.UpdateTask(context.Background(), 1, 1, &domain.UpdateTaskRequest{Status: &status})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "in_progress" {
		t.Fatalf("unexpected status: %s", resp.Status)
	}
	if len(taskRepo.history) == 0 {
		t.Fatal("expected history entry")
	}
}

func TestUpdateTask_NotFound(t *testing.T) {
	teamRepo := &mockTeamRepo{members: map[string]bool{teamMemberKey(1, 1): true}}
	uc := NewTaskUsecase(&mockTaskRepo{}, teamRepo, nil)

	_, err := uc.UpdateTask(context.Background(), 1, 99, &domain.UpdateTaskRequest{})
	if !errors.Is(err, domain.ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestGetTaskHistory_Success(t *testing.T) {
	teamRepo := &mockTeamRepo{members: map[string]bool{teamMemberKey(1, 1): true}}
	taskRepo := &mockTaskRepo{
		getResult: &domain.Task{ID: 1, TeamID: 1},
		history:   []*domain.TaskHistory{{ID: 1, TaskID: 1, FieldName: "status"}},
	}
	uc := NewTaskUsecase(taskRepo, teamRepo, nil)

	history, err := uc.GetTaskHistory(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("expected history, got %+v", history)
	}
}

func TestIsValidStatus(t *testing.T) {
	if !isValidStatus("new") || isValidStatus("bad") {
		t.Fatal("unexpected validation result")
	}
}

func TestIsValidPriority(t *testing.T) {
	if !isValidPriority("high") || isValidPriority("bad") {
		t.Fatal("unexpected validation result")
	}
}
