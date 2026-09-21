package model

import (
	"context"
	"time"

	"github.com/FTATM/software-dashboard-tool/internal/auth"
)

type User struct {
	UserId        int        `json:"userId" db:"user_id"`
	FirstName     string     `json:"firstName" db:"first_name"`
	LastName      string     `json:"lastName" db:"last_name"`
	Username      string     `json:"username" db:"username"`
	PasswordHash  string     `json:"-" db:"password_hash"`
	Active        bool       `json:"active" db:"active"`
	RoleId        int        `json:"roleId" db:"role_id"`
	Email         string     `json:"email" db:"email"`
	Tel           string     `json:"tel" db:"tel"`
	LineUserToken *string    `json:"lineUserToken" db:"line_user_token"`
	CreatedAt     time.Time  `json:"-" db:"created_at"`
	DeletedAt     *time.Time `json:"-" db:"deleted_at"`
}

func (s *User) IsSame(req User) bool {
	if s == nil {
		return false
	}

	return s.UserId == req.UserId &&
		s.FirstName == req.FirstName &&
		s.LastName == req.LastName &&
		s.Username == req.Username &&
		s.RoleId == req.RoleId &&
		s.Email == req.Email &&
		s.Tel == req.Tel &&
		s.Active == req.Active
}

type UserDetail struct {
	UserId        int     `json:"userId"`
	FirstName     string  `json:"firstName"`
	LastName      string  `json:"lastName"`
	Username      string  `json:"username"`
	Active        bool    `json:"active"`
	RoleId        int     `json:"roleId"`
	Email         string  `json:"email"`
	Tel           string  `json:"tel"`
	LineUserToken *string `json:"lineUserToken"`
}

type LoginCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateUser struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Active    bool   `json:"active"`
	RoleId    int    `json:"roleId"`
	Email     string `json:"email"`
	Tel       string `json:"tel"`
}

type UpdateUser struct {
	UserId    int    `json:"userId"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Password  string `json:"password,omitempty"`
	Active    bool   `json:"active"`
	RoleId    int    `json:"roleId"`
	Email     string `json:"email"`
	Tel       string `json:"tel"`
}

type UserLinkLine struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	IdToken       string `json:"idToken"`
	LineUserToken string `json:"-"`
}

type UserRepository interface {
	GetById(ctx context.Context, userId int) (*User, error)
	Create(ctx context.Context, user *User) error
	GetByUsername(ctx context.Context, username string) (*User, error)
	UserCount(ctx context.Context) (int, error)
	GetAll(ctx context.Context, active bool) ([]User, error)
	Update(ctx context.Context, user *User) error
	CountValidate(ctx context.Context, user *User) (int, error)
	Delete(ctx context.Context, userId int) error
	GetActiveById(ctx context.Context, userId int) (bool, error)
	GetUserForPermissionById(ctx context.Context, userId int) (*User, error)
	UpdateLineUserToken(ctx context.Context, userId int, lineUserToken string) error
}

type UserService interface {
	LoginUserJwt(ctx context.Context, creds *LoginCredentials, issueTime, expTime time.Time, clientInfo *auth.ClientInfo) (*User, string, error)
	CreateUser(ctx context.Context, createUser *CreateUser, authUserId int) (*User, error)
	GetAllDetail(ctx context.Context, active bool) ([]UserDetail, error)
	UpdateUser(ctx context.Context, updateUser *UpdateUser, authUserId int) (*User, error)
	GetPermissionMapByUserId(ctx context.Context, userId int) (map[string][]string, error)
	DeleteUser(ctx context.Context, deleteUserId, authUserId int) error
	UserLinkLineUserToken(ctx context.Context, linkline UserLinkLine, authUserId int) error
}
