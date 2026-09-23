package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/FTATM/software-dashboard-tool/internal/auth"
	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
)

type userService struct {
	txManager    model.TransactionManager
	prefixError  string
	userRepo     model.UserRepository
	jwtKey       []byte
	roleRepo     model.RoleRepository
	auditLogRepo model.AuditLogRepository
	roleCache    model.RoleCache
}

func NewUserService(txManager model.TransactionManager, userRepo model.UserRepository, jwtKey []byte, roleRepo model.RoleRepository, auditLogRepo model.AuditLogRepository, roleCache model.RoleCache) model.UserService {
	return &userService{
		txManager:    txManager,
		prefixError:  "userService",
		userRepo:     userRepo,
		jwtKey:       jwtKey,
		roleRepo:     roleRepo,
		auditLogRepo: auditLogRepo,
		roleCache:    roleCache,
	}
}

func (s *userService) CreateUser(ctx context.Context, createUser *model.CreateUser, authUserId int) (*model.User, error) {
	const fname = "CreateUser"
	hash, err := argon2id.CreateHash(createUser.Password, argon2id.DefaultParams)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	user := model.User{
		FirstName:    createUser.FirstName,
		LastName:     createUser.LastName,
		Username:     createUser.Username,
		Active:       createUser.Active,
		PasswordHash: hash,
		RoleId:       createUser.RoleId,
		Email:        createUser.Email,
		Tel:          createUser.Tel,
	}

	countValidate, err := s.userRepo.CountValidate(ctx, &user)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if countValidate > 0 || user.Username == "System" {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, model.ErrDuplicate)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	err = s.userRepo.Create(ctx, &user)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	newData, err := model.StructToDynamicJSON(user)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	audit := model.AuditLog{
		EntityType: "user",
		EntityId:   strconv.Itoa(user.UserId),
		MenuType:   "User",
		Action:     model.CreateAction,
		ChangedBy:  authUserId,
		OldData:    nil,
		NewData:    newData,
	}

	if err = s.auditLogRepo.Create(tx.Context(), []model.AuditLog{audit}); err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return &user, nil
}

func (s *userService) UpdateUser(ctx context.Context, updateUser *model.UpdateUser, authUserId int) (*model.User, error) {
	const fname = "UpdateUser"
	var err error
	var hash string
	if len(updateUser.Password) != 0 {
		hash, err = argon2id.CreateHash(updateUser.Password, argon2id.DefaultParams)
		if err != nil {
			return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
		}
	}

	user := model.User{
		UserId:        updateUser.UserId,
		FirstName:     updateUser.FirstName,
		LastName:      updateUser.LastName,
		Active:        updateUser.Active,
		PasswordHash:  hash,
		RoleId:        updateUser.RoleId,
		Email:         updateUser.Email,
		Tel:           updateUser.Tel,
		LineUserToken: updateUser.LineUserToken,
	}

	countValidate, err := s.userRepo.CountValidate(ctx, &user)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if countValidate > 0 || user.Username == "System" {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, model.ErrDuplicate)
	}

	oldUser, err := s.userRepo.GetById(ctx, user.UserId)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if oldUser.IsSame(user) && len(user.PasswordHash) == 0 {
		return &user, nil
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	err = s.userRepo.Update(ctx, &user)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	oldData, err := model.StructToDynamicJSON(oldUser)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	newData, err := model.StructToDynamicJSON(user)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	audit := model.AuditLog{
		EntityType: "user",
		EntityId:   strconv.Itoa(user.UserId),
		MenuType:   "User",
		Action:     model.UpdateAction,
		ChangedBy:  authUserId,
		OldData:    oldData,
		NewData:    newData,
	}

	if err = s.auditLogRepo.Create(tx.Context(), []model.AuditLog{audit}); err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return &user, nil
}

func (s *userService) GetPermissionMapByUserId(ctx context.Context, userId int) (map[string][]string, error) {
	const fname = "GetPermissionMapByUserId"

	// 1. Fetch the user to check active status AND get their RoleId
	user, err := s.userRepo.GetUserForPermissionById(ctx, userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		} else {
			return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
		}
	}

	// 2. ⚡ FAST PATH: Check the cache using roleId instead of userId
	if cached, ok := s.roleCache.Get(user.RoleId); ok {
		return cached, nil
	}

	// 3. SLOW PATH: Fetch permissions for this specific role from the database
	// We switch to your existing GetPermissionDescByRoleId method here
	rolePermissionDescs, err := s.roleRepo.GetPermissionDescByRoleId(ctx, user.RoleId)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	permissionMap := make(map[string][]string)
	for _, p := range rolePermissionDescs {
		permissionMap[p.MenuName] = append(permissionMap[p.MenuName], p.ActionName)
	}

	// 4. ⚡ SAVE TO CACHE: Store the result mapped to the roleId
	s.roleCache.Set(user.RoleId, permissionMap)

	return permissionMap, nil
}

func (s *userService) LoginUserJwt(ctx context.Context, creds *model.LoginCredentials, issueTime, expTime time.Time, clientInfo *auth.ClientInfo) (*model.User, string, error) {
	const fname = "LoginUserJwt"
	var tokenString string

	user, match, err := s.validateUserLogin(ctx, creds)
	if err != nil {
		return nil, tokenString, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if user == nil || !match {
		return nil, tokenString, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, model.ErrInvalidLogin)
	}

	claims := &auth.Claim{
		UserId:    user.UserId,
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		ExpiresAt: jwt.NewNumericDate(expTime),
		IssuedAt:  jwt.NewNumericDate(issueTime),
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Create the JWT string
	tokenString, err = token.SignedString(s.jwtKey)
	if err != nil {
		return nil, tokenString, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if clientInfo != nil {
		tx, err := s.txManager.Begin(ctx)
		if err != nil {
			return nil, tokenString, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
		}
		defer tx.Rollback(ctx)

		newData, err := model.StructToDynamicJSON(map[string]any{
			"ip":        clientInfo.IP,
			"userAgent": clientInfo.UserAgent,
		})
		if err != nil {
			return nil, tokenString, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
		}

		audit := model.AuditLog{
			EntityType: "user",
			EntityId:   strconv.Itoa(user.UserId),
			MenuType:   "Login",
			Action:     model.QueryAction,
			ChangedBy:  user.UserId,
			OldData:    nil,
			NewData:    newData,
		}

		if err = s.auditLogRepo.Create(tx.Context(), []model.AuditLog{audit}); err != nil {
			return nil, tokenString, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, tokenString, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
		}
	}

	return user, tokenString, nil
}

func (s *userService) validateUserLogin(ctx context.Context, creds *model.LoginCredentials) (*model.User, bool, error) {
	const fname = "validateLogin"

	user, err := s.userRepo.GetByUsername(ctx, creds.Username)
	if err != nil {
		return nil, false, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if !user.Active {
		return nil, false, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, model.ErrNotActive)
	}

	match, err := argon2id.ComparePasswordAndHash(creds.Password, user.PasswordHash)
	if err != nil {
		return nil, false, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return user, match, nil
}

func (s *userService) GetAllDetail(ctx context.Context, active bool) ([]model.UserDetail, error) {
	const fname = "GetAllDetail"
	users, err := s.userRepo.GetAll(ctx, active)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	userDetails := make([]model.UserDetail, 0, len(users))

	for _, u := range users {
		var truncatedToken *string
		if u.LineUserToken != nil {
			t := (*u.LineUserToken)[:min(len(*u.LineUserToken), 10)]
			truncatedToken = &t
		}

		detail := model.UserDetail{
			UserId:        u.UserId,
			FirstName:     u.FirstName,
			LastName:      u.LastName,
			Username:      u.Username,
			Active:        u.Active,
			RoleId:        u.RoleId,
			Email:         u.Email,
			Tel:           u.Tel,
			LineUserToken: truncatedToken,
		}

		userDetails = append(userDetails, detail)
	}
	return userDetails, nil
}

func (s *userService) DeleteUser(ctx context.Context, deleteUserId, authUserId int) error {
	const fname = "DeleteUser"
	var err error

	oldUser, err := s.userRepo.GetById(ctx, deleteUserId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	err = s.userRepo.Delete(ctx, deleteUserId)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	oldData, err := model.StructToDynamicJSON(oldUser)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	audit := model.AuditLog{
		EntityType: "user",
		EntityId:   strconv.Itoa(deleteUserId),
		MenuType:   "User",
		Action:     model.DeleteAction,
		ChangedBy:  authUserId,
		OldData:    oldData,
		NewData:    nil,
	}

	if err = s.auditLogRepo.Create(tx.Context(), []model.AuditLog{audit}); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}

func (s *userService) UserLinkLineUserToken(ctx context.Context, linkLine model.UserLinkLine, authUserId int) error {
	const fname = "UserLinkLineUserToken"
	var err error

	if len(linkLine.LineUserToken) == 0 || len(linkLine.Username) == 0 || len(linkLine.Password) == 0 {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, model.ErrInvalidLogin)
	}

	user, match, err := s.validateUserLogin(ctx, &model.LoginCredentials{Username: linkLine.Username, Password: linkLine.Password})
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if user == nil || !match {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, model.ErrInvalidLogin)
	}

	tx, err := s.txManager.Begin(ctx)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	defer tx.Rollback(ctx)

	err = s.userRepo.UpdateLineUserToken(ctx, user.UserId, linkLine.LineUserToken)
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	newData, err := model.StructToDynamicJSON(map[string]any{"lineUserToken": linkLine.LineUserToken})
	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}
	audit := model.AuditLog{
		EntityType: "user",
		EntityId:   strconv.Itoa(user.UserId),
		MenuType:   "User",
		Action:     model.UpdateAction,
		ChangedBy:  authUserId,
		OldData:    nil,
		NewData:    newData,
	}

	if err = s.auditLogRepo.Create(tx.Context(), []model.AuditLog{audit}); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", s.prefixError, fname, err)
	}

	return nil
}
