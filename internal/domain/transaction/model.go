package transaction

import (
	"time"

	"github.com/google/uuid"
)

type TransactionType string
type TransactionStatus string

const (
	TransactionTypeTransfer   TransactionType = "transfer"
	TransactionTypeDeposit    TransactionType = "deposit"
	TransactionTypeWithdrawal TransactionType = "withdrawal"
	TransactionTypeInterest   TransactionType = "interest"
	TransactionTypeFee        TransactionType = "fee"

	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusReversed  TransactionStatus = "reversed"
)

type Transaction struct {
	ID              uuid.UUID              `json:"id"`
	IdempotencyKey  string                 `json:"idempotency_key"`
	FromAccountID   *uuid.UUID             `json:"from_account_id,omitempty"`
	ToAccountID     *uuid.UUID             `json:"to_account_id,omitempty"`
	Amount          float64                `json:"amount"`
	TransactionType TransactionType        `json:"transaction_type"`
	Status          TransactionStatus      `json:"status"`
	Description     string                 `json:"description,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	CompletedAt     *time.Time             `json:"completed_at,omitempty"`
}

type TransferRequest struct {
	FromAccountID  string  `json:"from_account_id" binding:"required,uuid"`
	ToAccountID    string  `json:"to_account_id" binding:"required,uuid"`
	Amount         float64 `json:"amount" binding:"required,gt=0"`
	Description    string  `json:"description,omitempty"`
	IdempotencyKey string  `json:"idempotency_key" binding:"required"`
}

type DepositRequest struct {
	AccountID      string  `json:"account_id" binding:"required,uuid"`
	Amount         float64 `json:"amount" binding:"required,gt=0"`
	Description    string  `json:"description,omitempty"`
	IdempotencyKey string  `json:"idempotency_key" binding:"required"`
}

type WithdrawalRequest struct {
	AccountID      string  `json:"account_id" binding:"required,uuid"`
	Amount         float64 `json:"amount" binding:"required,gt=0"`
	Description    string  `json:"description,omitempty"`
	IdempotencyKey string  `json:"idempotency_key" binding:"required"`
}

type TransactionResponse struct {
	ID              uuid.UUID         `json:"id"`
	FromAccountID   *uuid.UUID        `json:"from_account_id,omitempty"`
	ToAccountID     *uuid.UUID        `json:"to_account_id,omitempty"`
	Amount          float64           `json:"amount"`
	TransactionType TransactionType   `json:"transaction_type"`
	Status          TransactionStatus `json:"status"`
	Description     string            `json:"description,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	CompletedAt     *time.Time        `json:"completed_at,omitempty"`
}

type TransactionHistoryRequest struct {
	AccountID string `form:"account_id" binding:"required,uuid"`
	Limit     int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset    int    `form:"offset" binding:"omitempty,min=0"`
	StartDate string `form:"start_date,omitempty"`
	EndDate   string `form:"end_date,omitempty"`
	TxnType   string `form:"type,omitempty"`
}

type TransactionHistoryResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	Total        int                   `json:"total"`
	Limit        int                   `json:"limit"`
	Offset       int                   `json:"offset"`
}

type QRResolutionResponse struct {
	AccountID uuid.UUID `json:"account_id"`
	OwnerName string    `json:"owner_name"`
	Currency  string    `json:"currency"`
}

type QRResolutionRequest struct {
	QRCode string `json:"qr_code" binding:"required"`
}
