package account

import (
	"time"

	"github.com/google/uuid"
)

type AccountType string
type AccountStatus string

const (
	AccountTypeChecking AccountType = "checking"
	AccountTypeSavings  AccountType = "savings"

	AccountStatusActive AccountStatus = "active"
	AccountStatusFrozen AccountStatus = "frozen"
	AccountStatusClosed AccountStatus = "closed"
)

type Account struct {
	ID            uuid.UUID     `json:"id"`
	UserID        uuid.UUID     `json:"user_id"`
	AccountNumber string        `json:"account_number"`
	AccountType   AccountType   `json:"account_type"`
	Balance       float64       `json:"balance"`
	Currency      string        `json:"currency"`
	InterestRate  float64       `json:"interest_rate"`
	Status        AccountStatus `json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type CreateAccountRequest struct {
	AccountType  string  `json:"account_type" binding:"required,oneof=checking savings"`
	Currency     string  `json:"currency" binding:"required,len=3"`
	InterestRate float64 `json:"interest_rate,omitempty"`
}

type AccountResponse struct {
	ID            uuid.UUID     `json:"id"`
	AccountNumber string        `json:"account_number"`
	AccountType   AccountType   `json:"account_type"`
	Balance       float64       `json:"balance"`
	Currency      string        `json:"currency"`
	InterestRate  float64       `json:"interest_rate"`
	Status        AccountStatus `json:"status"`
	CreatedAt     time.Time     `json:"created_at"`
}

type AccountListResponse struct {
	Accounts []AccountResponse `json:"accounts"`
	Total    int               `json:"total"`
}

type BalanceResponse struct {
	AccountID     uuid.UUID `json:"account_id"`
	AccountNumber string    `json:"account_number"`
	Balance       float64   `json:"balance"`
	Currency      string    `json:"currency"`
	AsOfDate      time.Time `json:"as_of_date"`
}

type UpdateAccountRequest struct {
	Status *string `json:"status,omitempty" binding:"omitempty,oneof=active frozen closed"`
}
