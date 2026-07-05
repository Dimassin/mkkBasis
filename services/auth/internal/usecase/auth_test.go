package usecase

import (
	"context"
	"errors"
	"testing"

	"auth/internal/domain"
)

type mockUserRepo struct {
	users      map[string]*domain.User
	createErr  error
	findResult *domain.User
	findErr    error
}

func (m *mockUserRepo) Create(ctx context.Context, user *domain.User) error {
	if m.createErr != nil {
		return m.createErr
	}
	if m.users == nil {
		m.users = make(map[string]*domain.User)
	}
	user.ID = "1"
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	if m.findResult != nil {
		return m.findResult, nil
	}
	if m.users != nil {
		return m.users[email], nil
	}
	return nil, nil
}

type mockHasher struct {
	hashErr    error
	compareErr error
}

func (m *mockHasher) Hash(ctx context.Context, password string) (string, error) {
	if m.hashErr != nil {
		return "", m.hashErr
	}
	return "hashed-" + password, nil
}

func (m *mockHasher) Compare(ctx context.Context, hashedPassword, password string) error {
	return m.compareErr
}

type mockTokenMgr struct {
	token  string
	genErr error
}

func (m *mockTokenMgr) GenerateAccessToken(ctx context.Context, user *domain.User) (string, error) {
	if m.genErr != nil {
		return "", m.genErr
	}
	if m.token != "" {
		return m.token, nil
	}
	return "token-" + user.ID, nil
}

func (m *mockTokenMgr) ValidateToken(ctx context.Context, token string) (*domain.UserClaims, error) {
	return nil, errors.New("not implemented")
}

func TestRegister_Success(t *testing.T) {
	uc := NewAuthUsecase(&mockUserRepo{}, &mockTokenMgr{}, &mockHasher{})

	resp, err := uc.Register(context.Background(), &domain.RegisterRequest{
		Email:    "a@test.com",
		Password: "secret12",
		Username: "alice",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken == "" || resp.Email != "a@test.com" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestRegister_UserAlreadyExists(t *testing.T) {
	repo := &mockUserRepo{
		users: map[string]*domain.User{
			"a@test.com": {Email: "a@test.com"},
		},
	}
	uc := NewAuthUsecase(repo, &mockTokenMgr{}, &mockHasher{})

	_, err := uc.Register(context.Background(), &domain.RegisterRequest{
		Email:    "a@test.com",
		Password: "secret12",
		Username: "alice",
	})
	if !errors.Is(err, domain.ErrUserAlreadyExists) {
		t.Fatalf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestRegister_WeakPassword(t *testing.T) {
	uc := NewAuthUsecase(&mockUserRepo{}, &mockTokenMgr{}, &mockHasher{})

	_, err := uc.Register(context.Background(), &domain.RegisterRequest{
		Email:    "a@test.com",
		Password: "123",
		Username: "alice",
	})
	if !errors.Is(err, domain.ErrWeakPassword) {
		t.Fatalf("expected ErrWeakPassword, got %v", err)
	}
}

func TestRegister_HashError(t *testing.T) {
	uc := NewAuthUsecase(&mockUserRepo{}, &mockTokenMgr{}, &mockHasher{hashErr: errors.New("hash failed")})

	_, err := uc.Register(context.Background(), &domain.RegisterRequest{
		Email:    "a@test.com",
		Password: "secret12",
		Username: "alice",
	})
	if !errors.Is(err, domain.ErrInternalServer) {
		t.Fatalf("expected ErrInternalServer, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	repo := &mockUserRepo{
		findResult: &domain.User{
			ID:       "1",
			Email:    "a@test.com",
			Password: "hashed",
			Username: "alice",
		},
	}
	uc := NewAuthUsecase(repo, &mockTokenMgr{token: "jwt"}, &mockHasher{})

	resp, err := uc.Login(context.Background(), &domain.LoginRequest{
		Email:    "a@test.com",
		Password: "secret12",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken != "jwt" {
		t.Fatalf("unexpected token: %s", resp.AccessToken)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	uc := NewAuthUsecase(&mockUserRepo{}, &mockTokenMgr{}, &mockHasher{})

	_, err := uc.Login(context.Background(), &domain.LoginRequest{
		Email:    "missing@test.com",
		Password: "secret12",
	})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_InvalidPassword(t *testing.T) {
	repo := &mockUserRepo{
		findResult: &domain.User{Email: "a@test.com", Password: "hashed"},
	}
	uc := NewAuthUsecase(repo, &mockTokenMgr{}, &mockHasher{compareErr: errors.New("mismatch")})

	_, err := uc.Login(context.Background(), &domain.LoginRequest{
		Email:    "a@test.com",
		Password: "wrong",
	})
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
