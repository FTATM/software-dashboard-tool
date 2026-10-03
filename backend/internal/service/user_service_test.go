package service_test

import (
	"context"
	"testing"

	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/FTATM/software-dashboard-tool/internal/service"
	"github.com/alexedwards/argon2id"
	"github.com/jackc/pgx/v5"
)

// Mock transaction primitives
type MockTransaction struct {
	ctx          context.Context
	CommitFunc   func(ctx context.Context) error
	RollbackFunc func(ctx context.Context) error
}

func (m *MockTransaction) Commit(ctx context.Context) error {
	if m.CommitFunc != nil {
		return m.CommitFunc(ctx)
	}
	return nil
}

func (m *MockTransaction) Rollback(ctx context.Context) error {
	if m.RollbackFunc != nil {
		return m.RollbackFunc(ctx)
	}
	return nil
}

func (m *MockTransaction) Context() context.Context {
	if m.ctx != nil {
		return m.ctx
	}
	return context.Background()
}

// MockTransactionManager implements model.TransactionManager
type MockTransactionManager struct {
	BeginFunc   func(ctx context.Context) (model.Transaction, error)
	BeginTxFunc func(ctx context.Context, opts pgx.TxOptions) (model.Transaction, error)
}

func (m *MockTransactionManager) Begin(ctx context.Context) (model.Transaction, error) {
	if m.BeginFunc != nil {
		return m.BeginFunc(ctx)
	}
	return &MockTransaction{ctx: ctx}, nil
}

func (m *MockTransactionManager) BeginTx(ctx context.Context, opts pgx.TxOptions) (model.Transaction, error) {
	if m.BeginTxFunc != nil {
		return m.BeginTxFunc(ctx, opts)
	}
	return &MockTransaction{ctx: ctx}, nil
}

// Mock repositories
type mockUserRepo struct {
	model.UserRepository
	countValidateFn func(ctx context.Context, user *model.User) (int, error)
	createFn        func(ctx context.Context, user *model.User) error
}

func (m *mockUserRepo) CountValidate(ctx context.Context, user *model.User) (int, error) {
	return m.countValidateFn(ctx, user)
}

func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	return m.createFn(ctx, user)
}

type mockAuditLogRepo struct {
	model.AuditLogRepository
	createdLogs []model.AuditLog
}

func (m *mockAuditLogRepo) Create(ctx context.Context, logs []model.AuditLog) error {
	m.createdLogs = append(m.createdLogs, logs...)
	return nil
}

func TestUserService_CreateUser(t *testing.T) {
	t.Run("fails when duplicate exists", func(t *testing.T) {
		repo := &mockUserRepo{
			countValidateFn: func(ctx context.Context, user *model.User) (int, error) {
				return 1, nil // Already exists
			},
		}

		// FIXED: Changed &mockTxManager{} to &MockTransactionManager{}
		svc := service.NewUserService(&MockTransactionManager{}, repo, []byte("secret"), nil, &mockAuditLogRepo{}, nil)

		req := &model.CreateUser{
			Username: "existing_user",
			Password: "secure_password",
		}

		_, err := svc.CreateUser(context.Background(), req, 1)
		if err == nil {
			t.Fatalf("expected duplicate error, got nil")
		}
	})

	t.Run("successfully creates user, hashes password, and creates audit log", func(t *testing.T) {
		var createdUser *model.User
		repo := &mockUserRepo{
			countValidateFn: func(ctx context.Context, user *model.User) (int, error) {
				return 0, nil
			},
			createFn: func(ctx context.Context, user *model.User) error {
				user.UserId = 99
				createdUser = user
				return nil
			},
		}
		auditRepo := &mockAuditLogRepo{}

		// FIXED: Changed &mockTxManager{} to &MockTransactionManager{}
		svc := service.NewUserService(&MockTransactionManager{}, repo, []byte("secret"), nil, auditRepo, nil)

		req := &model.CreateUser{
			FirstName: "Jane",
			LastName:  "Doe",
			Username:  "janed",
			Password:  "plainTextPass",
		}

		res, err := svc.CreateUser(context.Background(), req, 42)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// 1. Verify user was created
		if res.UserId != 99 {
			t.Errorf("expected user ID 99, got %d", res.UserId)
		}

		// 2. Verify password was properly hashed with Argon2id
		match, err := argon2id.ComparePasswordAndHash("plainTextPass", createdUser.PasswordHash)
		if err != nil || !match {
			t.Errorf("expected password hash to match plaintext password")
		}

		// 3. Verify audit log entry was created
		if len(auditRepo.createdLogs) != 1 {
			t.Fatalf("expected 1 audit log entry, got %d", len(auditRepo.createdLogs))
		}
		if auditRepo.createdLogs[0].ChangedBy != 42 {
			t.Errorf("expected audit changedBy 42, got %d", auditRepo.createdLogs[0].ChangedBy)
		}
	})
}
