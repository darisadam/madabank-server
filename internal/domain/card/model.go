package card

import (
	"time"

	"github.com/google/uuid"
)

type CardType string
type CardStatus string

const (
	CardTypeDebit  CardType = "debit"
	CardTypeCredit CardType = "credit"

	CardStatusActive  CardStatus = "active"
	CardStatusBlocked CardStatus = "blocked"
	CardStatusExpired CardStatus = "expired"
)

type Card struct {
	ID                  uuid.UUID  `json:"id"`
	AccountID           uuid.UUID  `json:"account_id"`
	CardNumberEncrypted string     `json:"-"` // Never expose in JSON
	CVVEncrypted        string     `json:"-"` // Never expose in JSON
	CardHolderName      string     `json:"card_holder_name"`
	CardType            CardType   `json:"card_type"`
	ExpiryMonth         int        `json:"expiry_month"`
	ExpiryYear          int        `json:"expiry_year"`
	Status              CardStatus `json:"status"`
	DailyLimit          float64    `json:"daily_limit"`
	CreatedAt           time.Time  `json:"created_at"`
}

type CardResponse struct {
	ID               uuid.UUID  `json:"id"`
	AccountID        uuid.UUID  `json:"account_id"`
	CardNumberMasked string     `json:"card_number_masked"` // Only last 4 digits
	CardHolderName   string     `json:"card_holder_name"`
	CardType         CardType   `json:"card_type"`
	ExpiryMonth      int        `json:"expiry_month"`
	ExpiryYear       int        `json:"expiry_year"`
	Status           CardStatus `json:"status"`
	DailyLimit       float64    `json:"daily_limit"`
	CreatedAt        time.Time  `json:"created_at"`
}

type CreateCardRequest struct {
	AccountID      string  `json:"account_id" binding:"required,uuid"`
	CardHolderName string  `json:"card_holder_name" binding:"required,min=3,max=100"`
	CardType       string  `json:"card_type" binding:"required,oneof=debit credit"`
	DailyLimit     float64 `json:"daily_limit" binding:"required,gt=0"`
}

type UpdateCardRequest struct {
	Status     *string  `json:"status,omitempty" binding:"omitempty,oneof=active blocked"`
	DailyLimit *float64 `json:"daily_limit,omitempty" binding:"omitempty,gt=0"`
}

type CardDetailsRequest struct {
	CardID string `json:"card_id" binding:"required,uuid"`
	// Require additional verification for showing full card details
	Password string `json:"password" binding:"required"`
}

type CardDetailsResponse struct {
	CardNumber  string `json:"card_number"`
	CVV         string `json:"cvv"`
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
}
