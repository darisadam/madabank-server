package repository

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/darisadam/madabank-server/internal/domain/account"
	"github.com/google/uuid"
)

type AccountRepository interface {
	Create(acc *account.Account) error
	GetByID(id uuid.UUID) (*account.Account, error)
	GetByAccountNumber(accountNumber string) (*account.Account, error)
	GetByUserID(userID uuid.UUID) ([]*account.Account, error)
	Update(id uuid.UUID, updates map[string]interface{}) error
	UpdateBalance(id uuid.UUID, newBalance float64) error
	Delete(id uuid.UUID) error
	GenerateAccountNumber() (string, error)
}

type accountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) Create(acc *account.Account) error {
	query := `
		INSERT INTO accounts (id, user_id, account_number, account_type, balance, currency, interest_rate, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		acc.ID,
		acc.UserID,
		acc.AccountNumber,
		acc.AccountType,
		acc.Balance,
		acc.Currency,
		acc.InterestRate,
		acc.Status,
	).Scan(&acc.CreatedAt, &acc.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	return nil
}

func (r *accountRepository) GetByID(id uuid.UUID) (*account.Account, error) {
	query := `
		SELECT id, user_id, account_number, account_type, balance, currency, 
		       interest_rate, status, created_at, updated_at
		FROM accounts
		WHERE id = $1 AND status != 'closed'
	`

	acc := &account.Account{}
	err := r.db.QueryRow(query, id).Scan(
		&acc.ID,
		&acc.UserID,
		&acc.AccountNumber,
		&acc.AccountType,
		&acc.Balance,
		&acc.Currency,
		&acc.InterestRate,
		&acc.Status,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("account not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return acc, nil
}

func (r *accountRepository) GetByAccountNumber(accountNumber string) (*account.Account, error) {
	query := `
		SELECT id, user_id, account_number, account_type, balance, currency,
		       interest_rate, status, created_at, updated_at
		FROM accounts
		WHERE account_number = $1 AND status != 'closed'
	`

	acc := &account.Account{}
	err := r.db.QueryRow(query, accountNumber).Scan(
		&acc.ID,
		&acc.UserID,
		&acc.AccountNumber,
		&acc.AccountType,
		&acc.Balance,
		&acc.Currency,
		&acc.InterestRate,
		&acc.Status,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("account not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return acc, nil
}

func (r *accountRepository) GetByUserID(userID uuid.UUID) ([]*account.Account, error) {
	query := `
		SELECT id, user_id, account_number, account_type, balance, currency,
		       interest_rate, status, created_at, updated_at
		FROM accounts
		WHERE user_id = $1 AND status != 'closed'
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list accounts: %w", err)
	}
	defer rows.Close()

	accounts := []*account.Account{}
	for rows.Next() {
		acc := &account.Account{}
		err := rows.Scan(
			&acc.ID,
			&acc.UserID,
			&acc.AccountNumber,
			&acc.AccountType,
			&acc.Balance,
			&acc.Currency,
			&acc.InterestRate,
			&acc.Status,
			&acc.CreatedAt,
			&acc.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, acc)
	}

	return accounts, nil
}

func (r *accountRepository) Update(id uuid.UUID, updates map[string]interface{}) error {
	query := "UPDATE accounts SET "
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

	query += fmt.Sprintf(", updated_at = CURRENT_TIMESTAMP WHERE id = $%d AND status != 'closed'", argPos)
	args = append(args, id)

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("account not found or already closed")
	}

	return nil
}

func (r *accountRepository) UpdateBalance(id uuid.UUID, newBalance float64) error {
	query := `
		UPDATE accounts 
		SET balance = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2 AND status = 'active'
	`

	result, err := r.db.Exec(query, newBalance, id)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("account not found or not active")
	}

	return nil
}

func (r *accountRepository) Delete(id uuid.UUID) error {
	// Soft delete by setting status to closed
	query := `UPDATE accounts SET status = 'closed', updated_at = CURRENT_TIMESTAMP WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to close account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("account not found")
	}

	return nil
}

func (r *accountRepository) GenerateAccountNumber() (string, error) {
	// Generate unique account number (format: MDAXXXXXXXXXX)
	// In production, you'd want more sophisticated logic

	const maxAttempts = 10
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < maxAttempts; i++ {
		// Generate 12-digit account number
		accountNumber := fmt.Sprintf("MDA%010d", rand.Int63n(10000000000))

		// Check if it already exists
		var exists bool
		query := `SELECT EXISTS(SELECT 1 FROM accounts WHERE account_number = $1)`
		err := r.db.QueryRow(query, accountNumber).Scan(&exists)
		if err != nil {
			return "", fmt.Errorf("failed to check account number: %w", err)
		}

		if !exists {
			return accountNumber, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique account number after %d attempts", maxAttempts)
}
