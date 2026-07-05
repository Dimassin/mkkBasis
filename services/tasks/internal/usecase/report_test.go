package usecase

import (
	"context"
	"errors"
	"testing"

	"tasks/internal/domain"
)

type mockReportRepo struct {
	stats    []*domain.TeamStats
	creators []*domain.TopCreator
	invalid  []*domain.InvalidAssigneeTask
	err      error
}

func (m *mockReportRepo) GetTeamStats(ctx context.Context) ([]*domain.TeamStats, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.stats, nil
}

func (m *mockReportRepo) GetTopCreatorsByTeam(ctx context.Context) ([]*domain.TopCreator, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.creators, nil
}

func (m *mockReportRepo) GetInvalidAssigneeTasks(ctx context.Context) ([]*domain.InvalidAssigneeTask, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.invalid, nil
}

func TestReportUsecase_GetTeamStats(t *testing.T) {
	uc := NewReportUsecase(&mockReportRepo{
		stats: []*domain.TeamStats{{TeamID: 1, TeamName: "Dev", MembersCount: 2}},
	})

	resp, err := uc.GetTeamStats(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestReportUsecase_GetInvalidAssigneeTasks(t *testing.T) {
	uc := NewReportUsecase(&mockReportRepo{
		invalid: []*domain.InvalidAssigneeTask{{TaskID: 1, Title: "bad"}},
	})

	resp, err := uc.GetInvalidAssigneeTasks(context.Background())
	if err != nil || len(resp) != 1 {
		t.Fatalf("unexpected response: %+v err=%v", resp, err)
	}
}

func TestReportUsecase_Error(t *testing.T) {
	uc := NewReportUsecase(&mockReportRepo{err: errors.New("db error")})

	_, err := uc.GetTopCreatorsByTeam(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}
