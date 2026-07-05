package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"teams/internal/domain"
)

type mockTeamRepo struct {
	teams        map[int]*domain.Team
	userTeams    []*domain.Team
	getTeamsErr  error
	nextTeamID   int
	createErr    error
	addMemberErr error
	memberRoles  map[string]string
	teamExists   map[int]bool
}

func memberKey(teamID, userID int) string {
	return fmt.Sprintf("%d:%d", teamID, userID)
}

func (m *mockTeamRepo) Create(ctx context.Context, team *domain.Team) error {
	if m.createErr != nil {
		return m.createErr
	}
	if m.teams == nil {
		m.teams = make(map[int]*domain.Team)
	}
	m.nextTeamID++
	team.ID = m.nextTeamID
	team.CreatedAt = time.Now()
	m.teams[team.ID] = team
	if m.teamExists == nil {
		m.teamExists = make(map[int]bool)
	}
	m.teamExists[team.ID] = true
	return nil
}

func (m *mockTeamRepo) AddMember(ctx context.Context, teamID, userID int, role string) error {
	if m.addMemberErr != nil {
		return m.addMemberErr
	}
	if m.memberRoles == nil {
		m.memberRoles = make(map[string]string)
	}
	m.memberRoles[memberKey(teamID, userID)] = role
	return nil
}

func (m *mockTeamRepo) GetUserTeams(ctx context.Context, userID int) ([]*domain.Team, error) {
	if m.getTeamsErr != nil {
		return nil, m.getTeamsErr
	}
	return m.userTeams, nil
}

func (m *mockTeamRepo) TeamExists(ctx context.Context, teamID int) (bool, error) {
	return m.teamExists[teamID], nil
}

func (m *mockTeamRepo) GetMemberRole(ctx context.Context, teamID, userID int) (string, error) {
	if m.memberRoles == nil {
		return "", nil
	}
	return m.memberRoles[memberKey(teamID, userID)], nil
}

type mockUserRepo struct {
	byID    map[string]*domain.User
	byEmail map[string]*domain.User
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	if m.byID == nil {
		return nil, nil
	}
	return m.byID[id], nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.byEmail == nil {
		return nil, nil
	}
	return m.byEmail[email], nil
}

type mockEmailService struct {
	err error
}

func (m *mockEmailService) SendTeamInvite(ctx context.Context, email string, teamID, inviterID int) error {
	return m.err
}

func TestCreateTeam_Success(t *testing.T) {
	userRepo := &mockUserRepo{byID: map[string]*domain.User{"1": {ID: "1", Email: "o@test.com"}}}
	teamRepo := &mockTeamRepo{}
	uc := NewTeamUsecase(teamRepo, userRepo, &mockEmailService{})

	resp, err := uc.CreateTeam(context.Background(), 1, &domain.CreateTeamRequest{Name: "Dev", Description: "desc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != 1 || resp.Name != "Dev" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestCreateTeam_UserNotFound(t *testing.T) {
	uc := NewTeamUsecase(&mockTeamRepo{}, &mockUserRepo{}, &mockEmailService{})

	_, err := uc.CreateTeam(context.Background(), 1, &domain.CreateTeamRequest{Name: "Dev"})
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestInviteMember_Success(t *testing.T) {
	teamRepo := &mockTeamRepo{
		teamExists: map[int]bool{1: true},
		memberRoles: map[string]string{
			memberKey(1, 1): "owner",
		},
	}
	userRepo := &mockUserRepo{
		byEmail: map[string]*domain.User{"m@test.com": {ID: "2", Email: "m@test.com"}},
	}
	uc := NewTeamUsecase(teamRepo, userRepo, &mockEmailService{})

	resp, err := uc.InviteMember(context.Background(), 1, 1, &domain.InviteMemberRequest{Email: "m@test.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.UserID != 2 || resp.Role != "member" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestInviteMember_Forbidden(t *testing.T) {
	teamRepo := &mockTeamRepo{
		teamExists: map[int]bool{1: true},
		memberRoles: map[string]string{
			memberKey(1, 2): "member",
		},
	}
	uc := NewTeamUsecase(teamRepo, &mockUserRepo{}, &mockEmailService{})

	_, err := uc.InviteMember(context.Background(), 2, 1, &domain.InviteMemberRequest{Email: "m@test.com"})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

func TestInviteMember_AlreadyMember(t *testing.T) {
	teamRepo := &mockTeamRepo{
		teamExists: map[int]bool{1: true},
		memberRoles: map[string]string{
			memberKey(1, 1): "owner",
			memberKey(1, 2): "member",
		},
	}
	userRepo := &mockUserRepo{
		byEmail: map[string]*domain.User{"m@test.com": {ID: "2", Email: "m@test.com"}},
	}
	uc := NewTeamUsecase(teamRepo, userRepo, &mockEmailService{})

	_, err := uc.InviteMember(context.Background(), 1, 1, &domain.InviteMemberRequest{Email: "m@test.com"})
	if !errors.Is(err, domain.ErrAlreadyMember) {
		t.Fatalf("expected ErrAlreadyMember, got %v", err)
	}
}

func TestInviteMember_TeamNotFound(t *testing.T) {
	uc := NewTeamUsecase(&mockTeamRepo{}, &mockUserRepo{}, &mockEmailService{})

	_, err := uc.InviteMember(context.Background(), 1, 99, &domain.InviteMemberRequest{Email: "m@test.com"})
	if !errors.Is(err, domain.ErrTeamNotFound) {
		t.Fatalf("expected ErrTeamNotFound, got %v", err)
	}
}

func TestInviteMember_AddMemberError(t *testing.T) {
	teamRepo := &mockTeamRepo{
		teamExists:   map[int]bool{1: true},
		memberRoles:  map[string]string{memberKey(1, 1): "admin"},
		addMemberErr: errors.New("db"),
	}
	userRepo := &mockUserRepo{
		byEmail: map[string]*domain.User{"m@test.com": {ID: "2", Email: "m@test.com"}},
	}
	uc := NewTeamUsecase(teamRepo, userRepo, &mockEmailService{})

	_, err := uc.InviteMember(context.Background(), 1, 1, &domain.InviteMemberRequest{Email: "m@test.com"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInviteMember_InvalidUserID(t *testing.T) {
	teamRepo := &mockTeamRepo{
		teamExists:  map[int]bool{1: true},
		memberRoles: map[string]string{memberKey(1, 1): "owner"},
	}
	userRepo := &mockUserRepo{
		byEmail: map[string]*domain.User{"m@test.com": {ID: "bad", Email: "m@test.com"}},
	}
	uc := NewTeamUsecase(teamRepo, userRepo, &mockEmailService{})

	_, err := uc.InviteMember(context.Background(), 1, 1, &domain.InviteMemberRequest{Email: "m@test.com"})
	if !errors.Is(err, domain.ErrInternalServer) {
		t.Fatalf("expected ErrInternalServer, got %v", err)
	}
}

func TestGetUserTeams_Error(t *testing.T) {
	teamRepo := &mockTeamRepo{getTeamsErr: errors.New("db")}
	uc := NewTeamUsecase(teamRepo, &mockUserRepo{}, &mockEmailService{})

	_, err := uc.GetUserTeams(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCreateTeam_AddMemberError(t *testing.T) {
	userRepo := &mockUserRepo{byID: map[string]*domain.User{"1": {ID: "1"}}}
	teamRepo := &mockTeamRepo{addMemberErr: errors.New("db error")}
	uc := NewTeamUsecase(teamRepo, userRepo, &mockEmailService{})

	_, err := uc.CreateTeam(context.Background(), 1, &domain.CreateTeamRequest{Name: "Dev"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInviteMember_UserNotFound(t *testing.T) {
	teamRepo := &mockTeamRepo{
		teamExists: map[int]bool{1: true},
		memberRoles: map[string]string{
			memberKey(1, 1): "owner",
		},
	}
	uc := NewTeamUsecase(teamRepo, &mockUserRepo{}, &mockEmailService{})

	_, err := uc.InviteMember(context.Background(), 1, 1, &domain.InviteMemberRequest{Email: "missing@test.com"})
	if !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestCreateTeam_RepoError(t *testing.T) {
	userRepo := &mockUserRepo{byID: map[string]*domain.User{"1": {ID: "1"}}}
	teamRepo := &mockTeamRepo{createErr: errors.New("db error")}
	uc := NewTeamUsecase(teamRepo, userRepo, &mockEmailService{})

	_, err := uc.CreateTeam(context.Background(), 1, &domain.CreateTeamRequest{Name: "Dev"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetUserTeams_Success(t *testing.T) {
	teamRepo := &mockTeamRepo{
		userTeams: []*domain.Team{{ID: 1, Name: "A", CreatedBy: 1, CreatedAt: time.Now()}},
	}
	uc := NewTeamUsecase(teamRepo, &mockUserRepo{}, &mockEmailService{})

	resp, err := uc.GetUserTeams(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 1 || resp[0].Name != "A" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
