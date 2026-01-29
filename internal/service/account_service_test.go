package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/darisadam/madabank-server/internal/domain/account"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAccountRepository is a mock implementation of repository.AccountRepository
type MockAccountRepository struct {
	mock.Mock
}

func (m *MockAccountRepository) Create(a *account.Account) error {
	args := m.Called(a)
	return args.Error(0)
}

func (m *MockAccountRepository) GetByID(id uuid.UUID) (*account.Account, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*account.Account), args.Error(1)
}

func (m *MockAccountRepository) GetByAccountNumber(accountNumber string) (*account.Account, error) {
	args := m.Called(accountNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*account.Account), args.Error(1)
}

func (m *MockAccountRepository) GetByUserID(userID uuid.UUID) ([]*account.Account, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*account.Account), args.Error(1)
}

func (m *MockAccountRepository) Update(id uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return args.Error(0)
}

func (m *MockAccountRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAccountRepository) GenerateAccountNumber() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockAccountRepository) UpdateBalance(id uuid.UUID, amount float64) error {
	args := m.Called(id, amount)
	return args.Error(0)
}

func setupAccountServiceTest(t *testing.T) (*accountService, *MockAccountRepository) {
	mockRepo := new(MockAccountRepository)
	svc := NewAccountService(mockRepo).(*accountService)
	return svc, mockRepo
}

func TestCreateAccount_Checking_Success(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	userID := uuid.New()

	req := &account.CreateAccountRequest{
		AccountType: "checking",
		Currency:    "USD",
	}

	// Mock: user has no existing accounts (allows creation)
	mockRepo.On("GetByUserID", userID).Return([]*account.Account{}, nil)
	mockRepo.On("GenerateAccountNumber").Return("1234567890", nil)
	mockRepo.On("Create", mock.AnythingOfType("*account.Account")).Return(nil)

	acc, err := svc.CreateAccount(userID, req)
	assert.NoError(t, err)
	assert.NotNil(t, acc)
	assert.Equal(t, account.AccountTypeChecking, acc.AccountType)
	assert.Equal(t, "USD", acc.Currency)
	assert.Equal(t, account.AccountStatusActive, acc.Status)
	assert.Equal(t, float64(0), acc.Balance)
	mockRepo.AssertExpectations(t)
}

func TestCreateAccount_Savings_WithDefaultInterest(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	userID := uuid.New()

	req := &account.CreateAccountRequest{
		AccountType:  "savings",
		Currency:     "USD",
		InterestRate: 0, // Should default to 3.25%
	}

	// Mock: user has no existing accounts (allows creation)
	mockRepo.On("GetByUserID", userID).Return([]*account.Account{}, nil)
	mockRepo.On("GenerateAccountNumber").Return("9876543210", nil)
	mockRepo.On("Create", mock.AnythingOfType("*account.Account")).Return(nil)

	acc, err := svc.CreateAccount(userID, req)
	assert.NoError(t, err)
	assert.Equal(t, account.AccountTypeSavings, acc.AccountType)
	assert.Equal(t, 0.0325, acc.InterestRate) // Default rate
	mockRepo.AssertExpectations(t)
}

func TestCreateAccount_InvalidType(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	userID := uuid.New()

	req := &account.CreateAccountRequest{
		AccountType: "invalid",
		Currency:    "USD",
	}

	// Mock: user has no existing accounts
	mockRepo.On("GetByUserID", userID).Return([]*account.Account{}, nil)

	acc, err := svc.CreateAccount(userID, req)
	assert.Error(t, err)
	assert.Nil(t, acc)
	assert.Contains(t, err.Error(), "invalid account type")
}

func TestCreateAccount_MaxAccountsReached(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	userID := uuid.New()

	req := &account.CreateAccountRequest{
		AccountType: "checking",
		Currency:    "IDR",
	}

	// Mock: user already has 3 accounts (max reached)
	existingAccounts := []*account.Account{
		{ID: uuid.New(), UserID: userID},
		{ID: uuid.New(), UserID: userID},
		{ID: uuid.New(), UserID: userID},
	}
	mockRepo.On("GetByUserID", userID).Return(existingAccounts, nil)

	acc, err := svc.CreateAccount(userID, req)
	assert.Error(t, err)
	assert.Nil(t, acc)
	assert.Contains(t, err.Error(), "maximum of 3 accounts")
}

func TestGetAccount_Success(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	userID := uuid.New()
	accountID := uuid.New()

	expectedAccount := &account.Account{
		ID:     accountID,
		UserID: userID,
	}

	mockRepo.On("GetByID", accountID).Return(expectedAccount, nil)

	acc, err := svc.GetAccount(accountID, userID)
	assert.NoError(t, err)
	assert.Equal(t, accountID, acc.ID)
	mockRepo.AssertExpectations(t)
}

func TestGetAccount_UnauthorizedAccess(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	ownerID := uuid.New()
	requestorID := uuid.New() // Different user
	accountID := uuid.New()

	existingAccount := &account.Account{
		ID:     accountID,
		UserID: ownerID, // Belongs to different user
	}

	mockRepo.On("GetByID", accountID).Return(existingAccount, nil)

	acc, err := svc.GetAccount(accountID, requestorID)
	assert.Error(t, err)
	assert.Nil(t, acc)
	assert.Contains(t, err.Error(), "unauthorized access")
	mockRepo.AssertExpectations(t)
}

func TestGetAccount_NotFound(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	accountID := uuid.New()
	userID := uuid.New()

	mockRepo.On("GetByID", accountID).Return(nil, fmt.Errorf("account not found"))

	acc, err := svc.GetAccount(accountID, userID)
	assert.Error(t, err)
	assert.Nil(t, acc)
	mockRepo.AssertExpectations(t)
}

func TestGetUserAccounts_Success(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	userID := uuid.New()

	accounts := []*account.Account{
		{ID: uuid.New(), UserID: userID},
		{ID: uuid.New(), UserID: userID},
	}

	mockRepo.On("GetByUserID", userID).Return(accounts, nil)

	result, err := svc.GetUserAccounts(userID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	mockRepo.AssertExpectations(t)
}

func TestGetBalance_Success(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	userID := uuid.New()
	accountID := uuid.New()

	expectedAccount := &account.Account{
		ID:            accountID,
		UserID:        userID,
		AccountNumber: "1234567890",
		Balance:       1000.50,
		Currency:      "USD",
		UpdatedAt:     time.Now(),
	}

	mockRepo.On("GetByID", accountID).Return(expectedAccount, nil)

	balance, err := svc.GetBalance(accountID, userID)
	assert.NoError(t, err)
	assert.Equal(t, 1000.50, balance.Balance)
	assert.Equal(t, "USD", balance.Currency)
	mockRepo.AssertExpectations(t)
}

func TestCloseAccount_Success(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	userID := uuid.New()
	accountID := uuid.New()

	existingAccount := &account.Account{
		ID:      accountID,
		UserID:  userID,
		Balance: 0, // Zero balance
	}

	mockRepo.On("GetByID", accountID).Return(existingAccount, nil)
	mockRepo.On("Delete", accountID).Return(nil)

	err := svc.CloseAccount(accountID, userID)
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestCloseAccount_NonZeroBalance(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	userID := uuid.New()
	accountID := uuid.New()

	existingAccount := &account.Account{
		ID:       accountID,
		UserID:   userID,
		Balance:  100.00, // Non-zero
		Currency: "USD",
	}

	mockRepo.On("GetByID", accountID).Return(existingAccount, nil)

	err := svc.CloseAccount(accountID, userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot close account with non-zero balance")
	mockRepo.AssertNotCalled(t, "Delete", mock.Anything)
}

func TestUpdateAccount_StatusTransition_Success(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	userID := uuid.New()
	accountID := uuid.New()

	existingAccount := &account.Account{
		ID:     accountID,
		UserID: userID,
		Status: account.AccountStatusActive,
	}

	newStatus := "frozen"
	req := &account.UpdateAccountRequest{Status: &newStatus}

	mockRepo.On("GetByID", accountID).Return(existingAccount, nil).Once()
	mockRepo.On("Update", accountID, mock.AnythingOfType("map[string]interface {}")).Return(nil)
	// After update, GetAccount is called again
	updatedAccount := &account.Account{
		ID:     accountID,
		UserID: userID,
		Status: account.AccountStatusFrozen,
	}
	mockRepo.On("GetByID", accountID).Return(updatedAccount, nil).Once()

	acc, err := svc.UpdateAccount(accountID, userID, req)
	assert.NoError(t, err)
	assert.Equal(t, account.AccountStatusFrozen, acc.Status)
	mockRepo.AssertExpectations(t)
}

func TestUpdateAccount_InvalidStatusTransition(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	userID := uuid.New()
	accountID := uuid.New()

	existingAccount := &account.Account{
		ID:     accountID,
		UserID: userID,
		Status: account.AccountStatusClosed, // Cannot transition from closed
	}

	newStatus := "active"
	req := &account.UpdateAccountRequest{Status: &newStatus}

	mockRepo.On("GetByID", accountID).Return(existingAccount, nil)

	acc, err := svc.UpdateAccount(accountID, userID, req)
	assert.Error(t, err)
	assert.Nil(t, acc)
	assert.Contains(t, err.Error(), "cannot transition from")
}

func TestValidateStatusTransition_AllCases(t *testing.T) {
	svc, _ := setupAccountServiceTest(t)

	tests := []struct {
		name    string
		current account.AccountStatus
		new     account.AccountStatus
		wantErr bool
	}{
		{"active->frozen", account.AccountStatusActive, account.AccountStatusFrozen, false},
		{"active->closed", account.AccountStatusActive, account.AccountStatusClosed, false},
		{"frozen->active", account.AccountStatusFrozen, account.AccountStatusActive, false},
		{"frozen->closed", account.AccountStatusFrozen, account.AccountStatusClosed, false},
		{"closed->active", account.AccountStatusClosed, account.AccountStatusActive, true},
		{"closed->frozen", account.AccountStatusClosed, account.AccountStatusFrozen, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.validateStatusTransition(tt.current, tt.new)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ==================== GetAccountByNumber Tests ====================

func TestGetAccountByNumber_Success(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	userID := uuid.New()
	accountNumber := "1234567890"

	expectedAccount := &account.Account{
		ID:            uuid.New(),
		UserID:        userID,
		AccountNumber: accountNumber,
	}

	mockRepo.On("GetByAccountNumber", accountNumber).Return(expectedAccount, nil)

	acc, err := svc.GetAccountByNumber(accountNumber, userID)
	assert.NoError(t, err)
	assert.Equal(t, accountNumber, acc.AccountNumber)
	mockRepo.AssertExpectations(t)
}

func TestGetAccountByNumber_NotFound(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	userID := uuid.New()
	accountNumber := "nonexistent"

	mockRepo.On("GetByAccountNumber", accountNumber).Return(nil, fmt.Errorf("account not found"))

	acc, err := svc.GetAccountByNumber(accountNumber, userID)
	assert.Error(t, err)
	assert.Nil(t, acc)
	mockRepo.AssertExpectations(t)
}

func TestGetAccountByNumber_Unauthorized(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	ownerID := uuid.New()
	requestorID := uuid.New() // Different user
	accountNumber := "1234567890"

	existingAccount := &account.Account{
		ID:            uuid.New(),
		UserID:        ownerID, // Belongs to different user
		AccountNumber: accountNumber,
	}

	mockRepo.On("GetByAccountNumber", accountNumber).Return(existingAccount, nil)

	acc, err := svc.GetAccountByNumber(accountNumber, requestorID)
	assert.Error(t, err)
	assert.Nil(t, acc)
	assert.Contains(t, err.Error(), "unauthorized access")
	mockRepo.AssertExpectations(t)
}

// ==================== GetBalance Additional Tests ====================

func TestGetBalance_Unauthorized(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	ownerID := uuid.New()
	requestorID := uuid.New()
	accountID := uuid.New()

	existingAccount := &account.Account{
		ID:     accountID,
		UserID: ownerID,
	}

	mockRepo.On("GetByID", accountID).Return(existingAccount, nil)

	balance, err := svc.GetBalance(accountID, requestorID)
	assert.Error(t, err)
	assert.Nil(t, balance)
	assert.Contains(t, err.Error(), "unauthorized access")
}

// ==================== UpdateAccount Additional Tests ====================

func TestUpdateAccount_UpdateStatusToFrozen(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	userID := uuid.New()
	accountID := uuid.New()

	existingAccount := &account.Account{
		ID:     accountID,
		UserID: userID,
		Status: account.AccountStatusActive,
	}

	status := "frozen"
	req := &account.UpdateAccountRequest{Status: &status}

	mockRepo.On("GetByID", accountID).Return(existingAccount, nil).Once()
	mockRepo.On("Update", accountID, mock.AnythingOfType("map[string]interface {}")).Return(nil)
	// After update, GetAccount is called again
	updatedAccount := &account.Account{
		ID:     accountID,
		UserID: userID,
		Status: account.AccountStatusFrozen,
	}
	mockRepo.On("GetByID", accountID).Return(updatedAccount, nil).Once()

	acc, err := svc.UpdateAccount(accountID, userID, req)
	assert.NoError(t, err)
	assert.Equal(t, account.AccountStatusFrozen, acc.Status)
	mockRepo.AssertExpectations(t)
}

func TestUpdateAccount_Unauthorized(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	ownerID := uuid.New()
	requestorID := uuid.New()
	accountID := uuid.New()

	existingAccount := &account.Account{
		ID:     accountID,
		UserID: ownerID,
	}

	status := "frozen"
	req := &account.UpdateAccountRequest{Status: &status}

	mockRepo.On("GetByID", accountID).Return(existingAccount, nil)

	acc, err := svc.UpdateAccount(accountID, requestorID, req)
	assert.Error(t, err)
	assert.Nil(t, acc)
	assert.Contains(t, err.Error(), "unauthorized access")
}

func TestUpdateAccount_UpdateFails(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	userID := uuid.New()
	accountID := uuid.New()

	existingAccount := &account.Account{
		ID:     accountID,
		UserID: userID,
		Status: account.AccountStatusActive,
	}

	status := "frozen"
	req := &account.UpdateAccountRequest{Status: &status}

	mockRepo.On("GetByID", accountID).Return(existingAccount, nil).Once()
	mockRepo.On("Update", accountID, mock.AnythingOfType("map[string]interface {}")).Return(fmt.Errorf("database error"))

	acc, err := svc.UpdateAccount(accountID, userID, req)
	assert.Error(t, err)
	assert.Nil(t, acc)
}

// ==================== CloseAccount Additional Tests ====================

func TestCloseAccount_Unauthorized(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	ownerID := uuid.New()
	requestorID := uuid.New()
	accountID := uuid.New()

	existingAccount := &account.Account{
		ID:      accountID,
		UserID:  ownerID,
		Balance: 0,
	}

	mockRepo.On("GetByID", accountID).Return(existingAccount, nil)

	err := svc.CloseAccount(accountID, requestorID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized access")
	mockRepo.AssertNotCalled(t, "Delete", mock.Anything)
}

func TestCloseAccount_NotFound(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	userID := uuid.New()
	accountID := uuid.New()

	mockRepo.On("GetByID", accountID).Return(nil, fmt.Errorf("account not found"))

	err := svc.CloseAccount(accountID, userID)
	assert.Error(t, err)
	mockRepo.AssertNotCalled(t, "Delete", mock.Anything)
}

// ==================== CreateAccount Additional Tests ====================

func TestCreateAccount_GenerateAccountNumberFails(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	userID := uuid.New()

	req := &account.CreateAccountRequest{
		AccountType: "checking",
		Currency:    "IDR",
	}

	mockRepo.On("GetByUserID", userID).Return([]*account.Account{}, nil)
	mockRepo.On("GenerateAccountNumber").Return("", fmt.Errorf("failed to generate"))

	acc, err := svc.CreateAccount(userID, req)
	assert.Error(t, err)
	assert.Nil(t, acc)
}

func TestCreateAccount_CreateFails(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	userID := uuid.New()

	req := &account.CreateAccountRequest{
		AccountType: "checking",
		Currency:    "IDR",
	}

	mockRepo.On("GetByUserID", userID).Return([]*account.Account{}, nil)
	mockRepo.On("GenerateAccountNumber").Return("1234567890", nil)
	mockRepo.On("Create", mock.AnythingOfType("*account.Account")).Return(fmt.Errorf("database error"))

	acc, err := svc.CreateAccount(userID, req)
	assert.Error(t, err)
	assert.Nil(t, acc)
}

func TestGetUserAccounts_Error(t *testing.T) {
	svc, mockRepo := setupAccountServiceTest(t)
	userID := uuid.New()

	mockRepo.On("GetByUserID", userID).Return(nil, fmt.Errorf("database error"))

	accounts, err := svc.GetUserAccounts(userID)
	assert.Error(t, err)
	assert.Nil(t, accounts)
}
