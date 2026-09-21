package repo

import (
	"context"
	"fmt"

	"github.com/FTATM/software-dashboard-tool/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepo struct {
	pool        DBTX
	prefixError string
}

func NewUserRepository(pool *pgxpool.Pool) model.UserRepository {
	return &userRepo{pool: pool, prefixError: "userRepo"}
}

func (r *userRepo) db(ctx context.Context) DBTX {
	if tx := extractTx(ctx); tx != nil {
		return tx
	}
	return r.pool
}

func (r *userRepo) GetById(ctx context.Context, id int) (*model.User, error) {
	const fname = "GetById"
	user := &model.User{}
	query := `SELECT user_id, first_name, last_name, active, role_id, email, tel FROM "user" WHERE user_id = $1`
	err := r.db(ctx).QueryRow(ctx, query, id).Scan(
		&user.UserId,
		&user.FirstName,
		&user.LastName,
		&user.Active,
		&user.RoleId,
		&user.Email,
		&user.Tel,
	)

	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return user, nil
}

func (r *userRepo) UserCount(ctx context.Context) (int, error) {
	const fname = "UserCount"
	var count int
	query := `SELECT count(*) FROM "user"`
	err := r.db(ctx).QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return count, nil
}

func (r *userRepo) GetAll(ctx context.Context, active bool) ([]model.User, error) {
	const fname = "GetAll"
	query := `
		SELECT 
			user_id, 
			first_name, 
			last_name, 
			username, 
			active, 
			role_id,
			email,
			tel,
			line_user_token
		FROM "user" 
		WHERE ($1 = false) OR ($1 = true AND deleted_at IS NULL)
		`
	rows, err := r.db(ctx).Query(ctx, query, active)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	users, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[model.User])
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return users, nil
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	const fname = "GetByUsername"
	var user model.User
	query := `SELECT user_id, username, first_name, last_name, password_hash, active, role_id FROM "user" WHERE username = $1`
	err := r.db(ctx).QueryRow(ctx, query, username).Scan(
		&user.UserId,
		&user.Username,
		&user.FirstName,
		&user.LastName,
		&user.PasswordHash,
		&user.Active,
		&user.RoleId,
	)
	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return &user, nil
}

func (r *userRepo) Create(ctx context.Context, user *model.User) error {
	const fname = "Create"
	query := `
			INSERT INTO "user" (first_name, last_name, username, password_hash, active, role_id, email, tel)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			RETURNING user_id
		`
	err := r.db(ctx).QueryRow(ctx, query, user.FirstName, user.LastName, user.Username, user.PasswordHash, user.Active, user.RoleId, user.Email, user.Tel).Scan(&user.UserId)

	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return nil

}

func (r *userRepo) Update(ctx context.Context, user *model.User) error {
	const fname = "Update"
	query := `
			UPDATE "user"
			SET 
				first_name = $1, 
				last_name = $2,
				active = $3,
				password_hash = COALESCE(NULLIF($4, ''), password_hash),
				role_id = $6,
				email = $7,
				tel = $8
			WHERE user_id = $5
		`
	result, err := r.db(ctx).Exec(ctx, query, user.FirstName, user.LastName, user.Active, user.PasswordHash, user.UserId, user.RoleId, user.Email, user.Tel)

	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, pgx.ErrNoRows)
	}

	return nil

}

func (r *userRepo) CountValidate(ctx context.Context, user *model.User) (int, error) {
	const fname = "CountValidate"
	var count int
	query := `
		SELECT count(*) 
		FROM "user" 
		WHERE 
			first_name = $1 AND 
			last_name = $2 AND 
			user_id != $3 AND
			active = true
	`
	err := r.db(ctx).QueryRow(ctx, query, user.FirstName, user.LastName, user.UserId).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	return count, nil
}

func (r *userRepo) Delete(ctx context.Context, userId int) error {
	const fname = "Delete"
	query := `
			UPDATE "user"
			SET 
				active = false,
				deleted_at = now()
			WHERE user_id = $1
		`
	result, err := r.db(ctx).Exec(ctx, query, userId)

	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, pgx.ErrNoRows)
	}

	return nil

}

func (r *userRepo) GetActiveById(ctx context.Context, userId int) (bool, error) {
	const fname = "GetActiveById"
	var userActive bool
	query := `SELECT active FROM "user" WHERE user_id = $1`
	err := r.db(ctx).QueryRow(ctx, query, userId).Scan(
		&userActive,
	)

	if err != nil {
		return false, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return userActive, nil
}

func (r *userRepo) GetUserForPermissionById(ctx context.Context, userId int) (*model.User, error) {
	const fname = "GetUserForPermissionById"
	var user model.User
	query := `SELECT user_id, role_id FROM "user" WHERE user_id = $1 AND active = true`
	err := r.db(ctx).QueryRow(ctx, query, userId).Scan(
		&user.UserId, &user.RoleId,
	)

	if err != nil {
		return nil, fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}
	return &user, nil
}

func (r *userRepo) UpdateLineUserToken(ctx context.Context, userId int, lineUserToken string) error {
	const fname = "Delete"
	query := `
			UPDATE "user"
			SET 
				line_user_token = $2
			WHERE user_id = $1
		`
	result, err := r.db(ctx).Exec(ctx, query, userId, lineUserToken)

	if err != nil {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, err)
	}

	if result.RowsAffected() != 1 {
		return fmt.Errorf("[%s]>[%s]: %w", r.prefixError, fname, pgx.ErrNoRows)
	}

	return nil

}
