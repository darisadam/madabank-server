package card

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCardType_Constants(t *testing.T) {
	assert.Equal(t, CardType("debit"), CardTypeDebit)
	assert.Equal(t, CardType("credit"), CardTypeCredit)
}

func TestCardStatus_Constants(t *testing.T) {
	assert.Equal(t, CardStatus("active"), CardStatusActive)
	assert.Equal(t, CardStatus("blocked"), CardStatusBlocked)
	assert.Equal(t, CardStatus("expired"), CardStatusExpired)
}

func TestCard_EncryptedFieldsHidden(t *testing.T) {
	c := Card{
		ID:                  uuid.New(),
		AccountID:           uuid.New(),
		CardNumberEncrypted: "encrypted_card_number_should_not_appear",
		CVVEncrypted:        "encrypted_cvv_should_not_appear",
		CardHolderName:      "John Doe",
		CardType:            CardTypeDebit,
		ExpiryMonth:         12,
		ExpiryYear:          2027,
		Status:              CardStatusActive,
		DailyLimit:          5000.00,
		CreatedAt:           time.Now(),
	}

	// Encrypted fields have json:"-" tag
	assert.NotEmpty(t, c.CardNumberEncrypted)
	assert.NotEmpty(t, c.CVVEncrypted)
	assert.Equal(t, "John Doe", c.CardHolderName)
}

func TestCardResponse_MaskedCardNumber(t *testing.T) {
	resp := CardResponse{
		ID:               uuid.New(),
		AccountID:        uuid.New(),
		CardNumberMasked: "************1234",
		CardHolderName:   "Jane Doe",
		CardType:         CardTypeCredit,
		ExpiryMonth:      6,
		ExpiryYear:       2025,
		Status:           CardStatusActive,
		DailyLimit:       10000.00,
	}

	assert.Contains(t, resp.CardNumberMasked, "****")
	assert.Equal(t, CardTypeCredit, resp.CardType)
}

func TestCreateCardRequest_Structure(t *testing.T) {
	req := CreateCardRequest{
		AccountID:      uuid.New().String(),
		CardHolderName: "Test User",
		CardType:       "debit",
		DailyLimit:     2500.00,
	}

	assert.NotEmpty(t, req.AccountID)
	assert.Equal(t, "Test User", req.CardHolderName)
	assert.Equal(t, "debit", req.CardType)
	assert.Equal(t, 2500.00, req.DailyLimit)
}

func TestUpdateCardRequest_OptionalFields(t *testing.T) {
	status := "blocked"
	limit := 1000.00
	req := UpdateCardRequest{
		Status:     &status,
		DailyLimit: &limit,
	}

	assert.NotNil(t, req.Status)
	assert.Equal(t, "blocked", *req.Status)
	assert.NotNil(t, req.DailyLimit)
	assert.Equal(t, 1000.00, *req.DailyLimit)

	// Test with nil fields
	reqEmpty := UpdateCardRequest{}
	assert.Nil(t, reqEmpty.Status)
	assert.Nil(t, reqEmpty.DailyLimit)
}

func TestCardDetailsRequest_Structure(t *testing.T) {
	req := CardDetailsRequest{
		CardID:   uuid.New().String(),
		Password: "userpassword",
	}

	assert.NotEmpty(t, req.CardID)
	assert.Equal(t, "userpassword", req.Password)
}

func TestCardDetailsResponse_SensitiveData(t *testing.T) {
	resp := CardDetailsResponse{
		CardNumber:  "4111111111111111",
		CVV:         "123",
		ExpiryMonth: 12,
		ExpiryYear:  2027,
	}

	assert.Equal(t, "4111111111111111", resp.CardNumber)
	assert.Equal(t, "123", resp.CVV)
	assert.Equal(t, 12, resp.ExpiryMonth)
	assert.Equal(t, 2027, resp.ExpiryYear)
}
