package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/darisadam/madabank-server/internal/domain/user"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(user *user.User) error
	GetByID(id uuid.UUID) (*user.User, error)
	GetByEmail(email string) (*user.User, error)
	GetByPhone(phone string) (*user.User, error)
	Update(id uuid.UUID, updates map[string]interface{}) error
	Delete(id uuid.UUID) error
	List(limit, offset int) ([]*user.User, error)

	// Refresh Token methods
	SaveRefreshToken(userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	GetRefreshToken(tokenHash string) (uuid.UUID, time.Time, error)
	RevokeRefreshToken(tokenHash string) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(u *user.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, phone, date_of_birth, kyc_status, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		u.ID,
		u.Email,
		u.PasswordHash,
		u.FirstName,
		u.LastName,
		u.Phone,
		u.DateOfBirth,
		u.KYCStatus,
		u.IsActive,
	).Scan(&u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *userRepository) GetByID(id uuid.UUID) (*user.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone, date_of_birth, 
		       kyc_status, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	u := &user.User{}
	err := r.db.QueryRow(query, id).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.FirstName,
		&u.LastName,
		&u.Phone,
		&u.DateOfBirth,
		&u.KYCStatus,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return u, nil
}

func (r *userRepository) GetByEmail(email string) (*user.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone, date_of_birth,
		       kyc_status, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	u := &user.User{}
	err := r.db.QueryRow(query, email).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.FirstName,
		&u.LastName,
		&u.Phone,
		&u.DateOfBirth,
		&u.KYCStatus,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return u, nil
}

func (r *userRepository) GetByPhone(phone string) (*user.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone, date_of_birth,
		       kyc_status, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE phone = $1 AND deleted_at IS NULL
	`

	u := &user.User{}
	err := r.db.QueryRow(query, phone).Scan(
		&u.ID,
		&u.Email,
		&u.PasswordHash,
		&u.FirstName,
		&u.LastName,
		&u.Phone,
		&u.DateOfBirth,
		&u.KYCStatus,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return u, nil
}

func (r *userRepository) Update(id uuid.UUID, updates map[string]interface{}) error {
	// Build dynamic UPDATE query
	query := "UPDATE users SET "
	args := []interface{}{}
	argPos := 1

	for key, value := range updates {
		if argPos > 1 {
			query += ", "
		}
		query += fmt.Sprintf("%s = $%d", key, argPos)
		args = append(args, value)
		argPos++
	}

	query += fmt.Sprintf(", updated_at = CURRENT_TIMESTAMP WHERE id = $%d AND deleted_at IS NULL", argPos)
	args = append(args, id)

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepository) Delete(id uuid.UUID) error {
	// Soft delete
	query := `UPDATE users SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`

	result, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepository) List(limit, offset int) ([]*user.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, phone, date_of_birth,
		       kyc_status, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer func() { _ = rows.Close() }()

	users := []*user.User{}
	for rows.Next() {
		u := &user.User{}
		err := rows.Scan(
			&u.ID,
			&u.Email,
			&u.PasswordHash,
			&u.FirstName,
			&u.LastName,
			&u.Phone,
			&u.DateOfBirth,
			&u.KYCStatus,
			&u.IsActive,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, u)
	}

	return users, nil
}

func (r *userRepository) SaveRefreshToken(userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	query := `INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(query, uuid.New(), userID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}
	return nil
}

func (r *userRepository) GetRefreshToken(tokenHash string) (uuid.UUID, time.Time, error) {
	var userID uuid.UUID
	var expiresAt time.Time

	query := `SELECT user_id, expires_at FROM refresh_tokens WHERE token_hash = $1 AND revoked = false`
	err := r.db.QueryRow(query, tokenHash).Scan(&userID, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, time.Time{}, fmt.Errorf("invalid or revoked token")
		}
		return uuid.Nil, time.Time{}, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return userID, expiresAt, nil
}

func (r *userRepository) RevokeRefreshToken(tokenHash string) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE token_hash = $1`
	_, err := r.db.Exec(query, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}
	return nil
}
