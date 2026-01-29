package account

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAccountType_Constants(t *testing.T) {
	assert.Equal(t, AccountType("checking"), AccountTypeChecking)
	assert.Equal(t, AccountType("savings"), AccountTypeSavings)
}

func TestAccountStatus_Constants(t *testing.T) {
	assert.Equal(t, AccountStatus("active"), AccountStatusActive)
	assert.Equal(t, AccountStatus("frozen"), AccountStatusFrozen)
	assert.Equal(t, AccountStatus("closed"), AccountStatusClosed)
}

func TestAccount_Structure(t *testing.T) {
	userID := uuid.New()
	accountID := uuid.New()

	account := Account{
		ID:            accountID,
		UserID:        userID,
		AccountNumber: "1234567890",
		AccountType:   AccountTypeChecking,
		Balance:       1000.50,
		Currency:      "USD",
		InterestRate:  0.0,
		Status:        AccountStatusActive,
	}

	assert.Equal(t, accountID, account.ID)
	assert.Equal(t, userID, account.UserID)
	assert.Equal(t, "1234567890", account.AccountNumber)
	assert.Equal(t, AccountTypeChecking, account.AccountType)
	assert.Equal(t, 1000.50, account.Balance)
	assert.Equal(t, AccountStatusActive, account.Status)
}

func TestCreateAccountRequest_Fields(t *testing.T) {
	req := CreateAccountRequest{
		AccountType:  "savings",
		Currency:     "EUR",
		InterestRate: 0.05,
	}

	assert.Equal(t, "savings", req.AccountType)
	assert.Equal(t, "EUR", req.Currency)
	assert.Equal(t, 0.05, req.InterestRate)
}

func TestUpdateAccountRequest_OptionalStatus(t *testing.T) {
	status := "frozen"
	req := UpdateAccountRequest{
		Status: &status,
	}

	assert.NotNil(t, req.Status)
	assert.Equal(t, "frozen", *req.Status)

	// Test nil status
	reqEmpty := UpdateAccountRequest{}
	assert.Nil(t, reqEmpty.Status)
}

func TestBalanceResponse_Structure(t *testing.T) {
	accountID := uuid.New()
	resp := BalanceResponse{
		AccountID:     accountID,
		AccountNumber: "1234567890",
		Balance:       500.25,
		Currency:      "USD",
	}

	assert.Equal(t, accountID, resp.AccountID)
	assert.Equal(t, 500.25, resp.Balance)
}
