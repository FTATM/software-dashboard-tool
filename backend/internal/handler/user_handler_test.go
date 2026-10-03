package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FTATM/software-dashboard-tool/internal/auth"
	"github.com/FTATM/software-dashboard-tool/internal/handler"
	"github.com/FTATM/software-dashboard-tool/internal/model"
)

// Minimal mock using interface embedding
type mockUserService struct {
	model.UserService
	createUserFn func(ctx context.Context, u *model.CreateUser, authId int) (*model.User, error)
}

func (m *mockUserService) CreateUser(ctx context.Context, u *model.CreateUser, authId int) (*model.User, error) {
	return m.createUserFn(ctx, u, authId)
}

type mockRoleService struct {
	model.RoleService
	accessFn func(ctx context.Context, acc *model.Access) (bool, error)
}

func (m *mockRoleService) Access(ctx context.Context, acc *model.Access) (bool, error) {
	return m.accessFn(ctx, acc)
}

func TestUserHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		authUserId     any
		hasAccess      bool
		requestBody    model.CreateUser
		mockCreateUser func(ctx context.Context, u *model.CreateUser, authId int) (*model.User, error)
		expectedStatus int
	}{
		{
			name:       "unauthorized when auth context missing",
			authUserId: nil, // Simulates unauthenticated request
			hasAccess:  true,
			requestBody: model.CreateUser{
				FirstName: "John", LastName: "Doe", Username: "johnd", Password: "secretPassword",
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:       "forbidden when user lacks permission",
			authUserId: 42,
			hasAccess:  false, // RBAC check fails
			requestBody: model.CreateUser{
				FirstName: "John", LastName: "Doe", Username: "johnd", Password: "secretPassword",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "success when valid payload and permitted",
			authUserId: 42,
			hasAccess:  true,
			requestBody: model.CreateUser{
				FirstName: "John", LastName: "Doe", Username: "johnd", Password: "secretPassword",
			},
			mockCreateUser: func(ctx context.Context, u *model.CreateUser, authId int) (*model.User, error) {
				return &model.User{UserId: 10, Username: u.Username}, nil
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			userSvc := &mockUserService{createUserFn: tc.mockCreateUser}
			roleSvc := &mockRoleService{
				accessFn: func(ctx context.Context, acc *model.Access) (bool, error) {
					return tc.hasAccess, nil
				},
			}

			h := handler.NewUserHandler(userSvc, roleSvc, false, nil)

			body, _ := json.Marshal(tc.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))

			// Inject auth context if present in test case
			if tc.authUserId != nil {
				ctx := context.WithValue(req.Context(), auth.AuthUserIdKey, tc.authUserId)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()
			h.Create(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Fatalf("status code mismatch: got %d, want %d (body: %s)", rec.Code, tc.expectedStatus, rec.Body.String())
			}
		})
	}
}
