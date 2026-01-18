package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/darisadam/madabank-server/internal/domain/transaction"
	"github.com/google/uuid"
)

type TransactionRepository interface {
	Create(tx *transaction.Transaction) error
	GetByID(id uuid.UUID) (*transaction.Transaction, error)
	GetByIdempotencyKey(key string) (*transaction.Transaction, error)
	GetByAccountID(accountID uuid.UUID, limit, offset int) ([]*transaction.Transaction, error)
	GetByAccountIDWithFilters(accountID uuid.UUID, filters map[string]interface{}, limit, offset int) ([]*transaction.Transaction, error)
	UpdateStatus(id uuid.UUID, status transaction.TransactionStatus) error

	// ACID operations - these run in a database transaction
	ExecuteTransfer(fromAccountID, toAccountID uuid.UUID, amount float64, txn *transaction.Transaction) error
	ExecuteDeposit(accountID uuid.UUID, amount float64, txn *transaction.Transaction) error
	ExecuteWithdrawal(accountID uuid.UUID, amount float64, txn *transaction.Transaction) error
}

type transactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

func (r *transactionRepository) Create(txn *transaction.Transaction) error {
	// Serialize metadata to JSON
	metadataJSON, err := json.Marshal(txn.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO transactions (id, idempotency_key, from_account_id, to_account_id, 
		                         amount, transaction_type, status, description, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at
	`

	err = r.db.QueryRow(
		query,
		txn.ID,
		txn.IdempotencyKey,
		txn.FromAccountID,
		txn.ToAccountID,
		txn.Amount,
		txn.TransactionType,
		txn.Status,
		txn.Description,
		metadataJSON,
	).Scan(&txn.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

func (r *transactionRepository) GetByID(id uuid.UUID) (*transaction.Transaction, error) {
	query := `
		SELECT id, idempotency_key, from_account_id, to_account_id, amount, 
		       transaction_type, status, description, metadata, created_at, completed_at
		FROM transactions
		WHERE id = $1
	`

	txn := &transaction.Transaction{}
	var metadataJSON []byte

	err := r.db.QueryRow(query, id).Scan(
		&txn.ID,
		&txn.IdempotencyKey,
		&txn.FromAccountID,
		&txn.ToAccountID,
		&txn.Amount,
		&txn.TransactionType,
		&txn.Status,
		&txn.Description,
		&metadataJSON,
		&txn.CreatedAt,
		&txn.CompletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("transaction not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	// Unmarshal metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &txn.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return txn, nil
}

func (r *transactionRepository) GetByIdempotencyKey(key string) (*transaction.Transaction, error) {
	query := `
		SELECT id, idempotency_key, from_account_id, to_account_id, amount,
		       transaction_type, status, description, metadata, created_at, completed_at
		FROM transactions
		WHERE idempotency_key = $1
	`

	txn := &transaction.Transaction{}
	var metadataJSON []byte

	err := r.db.QueryRow(query, key).Scan(
		&txn.ID,
		&txn.IdempotencyKey,
		&txn.FromAccountID,
		&txn.ToAccountID,
		&txn.Amount,
		&txn.TransactionType,
		&txn.Status,
		&txn.Description,
		&metadataJSON,
		&txn.CreatedAt,
		&txn.CompletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("transaction not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &txn.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return txn, nil
}

func (r *transactionRepository) GetByAccountID(accountID uuid.UUID, limit, offset int) ([]*transaction.Transaction, error) {
	query := `
		SELECT id, idempotency_key, from_account_id, to_account_id, amount,
		       transaction_type, status, description, metadata, created_at, completed_at
		FROM transactions
		WHERE (from_account_id = $1 OR to_account_id = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, accountID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return r.scanTransactions(rows)
}

func (r *transactionRepository) GetByAccountIDWithFilters(accountID uuid.UUID, filters map[string]interface{}, limit, offset int) ([]*transaction.Transaction, error) {
	query := `
		SELECT id, idempotency_key, from_account_id, to_account_id, amount,
		       transaction_type, status, description, metadata, created_at, completed_at
		FROM transactions
		WHERE (from_account_id = $1 OR to_account_id = $1)
	`
	args := []interface{}{accountID}
	argPos := 2

	// Add filters
	if startDate, ok := filters["start_date"].(time.Time); ok {
		query += fmt.Sprintf(" AND created_at >= $%d", argPos)
		args = append(args, startDate)
		argPos++
	}

	if endDate, ok := filters["end_date"].(time.Time); ok {
		query += fmt.Sprintf(" AND created_at <= $%d", argPos)
		args = append(args, endDate)
		argPos++
	}

	if txnType, ok := filters["type"].(string); ok {
		query += fmt.Sprintf(" AND transaction_type = $%d", argPos)
		args = append(args, txnType)
		argPos++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPos, argPos+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return r.scanTransactions(rows)
}

func (r *transactionRepository) UpdateStatus(id uuid.UUID, status transaction.TransactionStatus) error {
	query := `
		UPDATE transactions 
		SET status = $1, completed_at = CASE WHEN $1 = 'completed' THEN CURRENT_TIMESTAMP ELSE completed_at END
		WHERE id = $2
	`

	result, err := r.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("transaction not found")
	}

	return nil
}

// ExecuteTransfer performs a transfer with ACID guarantees using database transaction
func (r *transactionRepository) ExecuteTransfer(fromAccountID, toAccountID uuid.UUID, amount float64, txn *transaction.Transaction) error {
	// Start database transaction
	dbTx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = dbTx.Rollback() // Rollback if not committed
	}()

	// Lock both accounts for update (prevents race conditions)
	var fromBalance, toBalance float64
	var fromStatus, toStatus string

	// Lock source account
	query := `SELECT balance, status FROM accounts WHERE id = $1 AND status = 'active' FOR UPDATE`
	err = dbTx.QueryRow(query, fromAccountID).Scan(&fromBalance, &fromStatus)
	if err != nil {
		return fmt.Errorf("failed to lock source account: %w", err)
	}

	// Lock destination account
	err = dbTx.QueryRow(query, toAccountID).Scan(&toBalance, &toStatus)
	if err != nil {
		return fmt.Errorf("failed to lock destination account: %w", err)
	}

	// Validate sufficient balance
	if fromBalance < amount {
		return fmt.Errorf("insufficient balance: have %.2f, need %.2f", fromBalance, amount)
	}

	// Debit source account
	_, err = dbTx.Exec(`UPDATE accounts SET balance = balance - $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, amount, fromAccountID)
	if err != nil {
		return fmt.Errorf("failed to debit source account: %w", err)
	}

	// Credit destination account
	_, err = dbTx.Exec(`UPDATE accounts SET balance = balance + $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, amount, toAccountID)
	if err != nil {
		return fmt.Errorf("failed to credit destination account: %w", err)
	}

	// Insert transaction record
	metadataJSON, _ := json.Marshal(txn.Metadata)
	_, err = dbTx.Exec(`
		INSERT INTO transactions (id, idempotency_key, from_account_id, to_account_id, 
		                         amount, transaction_type, status, description, metadata, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, CURRENT_TIMESTAMP)
	`, txn.ID, txn.IdempotencyKey, fromAccountID, toAccountID, amount, txn.TransactionType, transaction.TransactionStatusCompleted, txn.Description, metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	// Commit transaction - ACID guarantee
	if err := dbTx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ExecuteDeposit performs a deposit with ACID guarantees
func (r *transactionRepository) ExecuteDeposit(accountID uuid.UUID, amount float64, txn *transaction.Transaction) error {
	dbTx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = dbTx.Rollback()
	}()

	// Lock account
	var status string
	err = dbTx.QueryRow(`SELECT status FROM accounts WHERE id = $1 AND status = 'active' FOR UPDATE`, accountID).Scan(&status)
	if err != nil {
		return fmt.Errorf("failed to lock account: %w", err)
	}

	// Credit account
	_, err = dbTx.Exec(`UPDATE accounts SET balance = balance + $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, amount, accountID)
	if err != nil {
		return fmt.Errorf("failed to credit account: %w", err)
	}

	// Insert transaction
	metadataJSON, _ := json.Marshal(txn.Metadata)
	_, err = dbTx.Exec(`
		INSERT INTO transactions (id, idempotency_key, to_account_id, amount, transaction_type, status, description, metadata, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP)
	`, txn.ID, txn.IdempotencyKey, accountID, amount, txn.TransactionType, transaction.TransactionStatusCompleted, txn.Description, metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	if err := dbTx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ExecuteWithdrawal performs a withdrawal with ACID guarantees
func (r *transactionRepository) ExecuteWithdrawal(accountID uuid.UUID, amount float64, txn *transaction.Transaction) error {
	dbTx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = dbTx.Rollback()
	}()

	// Lock account and check balance
	var balance float64
	var status string
	err = dbTx.QueryRow(`SELECT balance, status FROM accounts WHERE id = $1 AND status = 'active' FOR UPDATE`, accountID).Scan(&balance, &status)
	if err != nil {
		return fmt.Errorf("failed to lock account: %w", err)
	}

	if balance < amount {
		return fmt.Errorf("insufficient balance: have %.2f, need %.2f", balance, amount)
	}

	// Debit account
	_, err = dbTx.Exec(`UPDATE accounts SET balance = balance - $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`, amount, accountID)
	if err != nil {
		return fmt.Errorf("failed to debit account: %w", err)
	}

	// Insert transaction
	metadataJSON, _ := json.Marshal(txn.Metadata)
	_, err = dbTx.Exec(`
		INSERT INTO transactions (id, idempotency_key, from_account_id, amount, transaction_type, status, description, metadata, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP)
	`, txn.ID, txn.IdempotencyKey, accountID, amount, txn.TransactionType, transaction.TransactionStatusCompleted, txn.Description, metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	if err := dbTx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *transactionRepository) scanTransactions(rows *sql.Rows) ([]*transaction.Transaction, error) {
	transactions := []*transaction.Transaction{}

	for rows.Next() {
		txn := &transaction.Transaction{}
		var metadataJSON []byte

		err := rows.Scan(
			&txn.ID,
			&txn.IdempotencyKey,
			&txn.FromAccountID,
			&txn.ToAccountID,
			&txn.Amount,
			&txn.TransactionType,
			&txn.Status,
			&txn.Description,
			&metadataJSON,
			&txn.CreatedAt,
			&txn.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &txn.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		transactions = append(transactions, txn)
	}

	return transactions, nil
}
