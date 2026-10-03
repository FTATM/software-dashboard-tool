package repo_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/FTATM/software-dashboard-tool/internal/repo"
	"github.com/jackc/pgx/v5/pgxpool"
)

// setupDB initializes the real database connection pool for tests.
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func setupDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	user := getEnv("DB_USER", "postgres")
	pass := getEnv("DB_PASSWORD", "admin1234")
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	name := getEnv("DB_NAME", "dashboardTest")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, name)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Verify connection is alive
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	return pool
}

// hardDelete cleans up rows after each test so future test runs don't collide on unique keys.
func hardDelete(t *testing.T, pool *pgxpool.Pool, userId int) {
	t.Helper()
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM "user" WHERE user_id = $1`, userId)
	})
}

func TestUserRepo_Create_And_GetById(t *testing.T) {
	pool := setupDB(t)
	r := repo.NewUserRepository(pool)
	ctx := context.Background()

	newUser := &model.User{
		FirstName:    "Alice",
		LastName:     "Tester",
		Username:     "alicet_unique",
		PasswordHash: "some_argon2_hash",
		Active:       true,
		RoleId:       1,
		Email:        "alice_unique@example.com",
		Tel:          "0811112222",
	}

	// 1. Test Create
	err := r.Create(ctx, newUser)
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	if newUser.UserId == 0 {
		t.Fatalf("Create() expected returning UserId to be populated, got 0")
	}

	// Ensure hard delete runs when the test completes
	hardDelete(t, pool, newUser.UserId)

	// 2. Test GetById
	found, err := r.GetById(ctx, newUser.UserId)
	if err != nil {
		t.Fatalf("GetById() unexpected error: %v", err)
	}

	if found.FirstName != newUser.FirstName || found.Email != newUser.Email {
		t.Errorf("GetById() got %+v, want %+v", found, newUser)
	}
}

func TestUserRepo_UpdateLineUserToken_Duplicate(t *testing.T) {
	pool := setupDB(t)
	r := repo.NewUserRepository(pool)
	ctx := context.Background()

	// Seed user 1
	u1 := &model.User{
		FirstName: "User", LastName: "One", Username: "u1_token_test",
		PasswordHash: "hash", Active: true, RoleId: 1, Email: "u1_token@test.com", Tel: "01",
	}
	if err := r.Create(ctx, u1); err != nil {
		t.Fatalf("failed to seed u1: %v", err)
	}
	hardDelete(t, pool, u1.UserId)

	// Seed user 2
	u2 := &model.User{
		FirstName: "User", LastName: "Two", Username: "u2_token_test",
		PasswordHash: "hash", Active: true, RoleId: 1, Email: "u2_token@test.com", Tel: "02",
	}
	if err := r.Create(ctx, u2); err != nil {
		t.Fatalf("failed to seed u2: %v", err)
	}
	hardDelete(t, pool, u2.UserId)

	token := "shared_unique_line_token"

	// Assign token to User 1 (should succeed)
	if err := r.UpdateLineUserToken(ctx, u1.UserId, token); err != nil {
		t.Fatalf("UpdateLineUserToken() for u1 failed: %v", err)
	}

	// Assign same token to User 2 (must trigger duplicate constraint error -> model.ErrDuplicate)
	err := r.UpdateLineUserToken(ctx, u2.UserId, token)
	if err == nil {
		t.Fatalf("expected ErrDuplicate for duplicate token, got nil")
	}

	if !errors.Is(err, model.ErrDuplicate) {
		t.Errorf("expected error to wrap model.ErrDuplicate, got: %v", err)
	}
}

func TestUserRepo_GetAll(t *testing.T) {
	pool := setupDB(t)
	r := repo.NewUserRepository(pool)
	ctx := context.Background()

	// Seed active user
	activeUser := &model.User{
		FirstName: "Active", LastName: "User", Username: "act1_test",
		PasswordHash: "h", Active: true, RoleId: 1, Email: "act1_test@test.com", Tel: "01",
	}
	if err := r.Create(ctx, activeUser); err != nil {
		t.Fatalf("failed to seed active user: %v", err)
	}
	hardDelete(t, pool, activeUser.UserId)

	// Seed soft-deleted user
	deletedUser := &model.User{
		FirstName: "Deleted", LastName: "User", Username: "del1_test",
		PasswordHash: "h", Active: true, RoleId: 1, Email: "del1_test@test.com", Tel: "02",
	}
	if err := r.Create(ctx, deletedUser); err != nil {
		t.Fatalf("failed to seed deleted user: %v", err)
	}
	hardDelete(t, pool, deletedUser.UserId)

	// Soft-delete user 2
	if err := r.Delete(ctx, deletedUser.UserId); err != nil {
		t.Fatalf("failed to soft delete user: %v", err)
	}

	// active=true should not return soft-deleted users
	usersOnlyActive, err := r.GetAll(ctx, true)
	if err != nil {
		t.Fatalf("GetAll(true) error: %v", err)
	}
	for _, u := range usersOnlyActive {
		if u.UserId == deletedUser.UserId {
			t.Errorf("GetAll(true) returned soft-deleted user ID %d", u.UserId)
		}
	}

	// active=false should return all users including soft-deleted ones
	usersAll, err := r.GetAll(ctx, false)
	if err != nil {
		t.Fatalf("GetAll(false) error: %v", err)
	}
	foundDeleted := false
	for _, u := range usersAll {
		if u.UserId == deletedUser.UserId {
			foundDeleted = true
			break
		}
	}
	if !foundDeleted {
		t.Errorf("GetAll(false) did not return the soft-deleted user")
	}
}
