package service

import (
	"fmt"
	"testing"
	"time"

	domainAccount "github.com/darisadam/madabank-server/internal/domain/account"
	"github.com/darisadam/madabank-server/internal/domain/audit"
	"github.com/darisadam/madabank-server/internal/domain/transaction"
	"github.com/darisadam/madabank-server/internal/domain/user"
	"github.com/darisadam/madabank-server/internal/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTransactionRepository is a mock implementation
type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) Create(t *transaction.Transaction) error {
	args := m.Called(t)
	return args.Error(0)
}

func (m *MockTransactionRepository) GetByID(id uuid.UUID) (*transaction.Transaction, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*transaction.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetByAccountID(accountID uuid.UUID, limit, offset int) ([]*transaction.Transaction, error) {
	args := m.Called(accountID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*transaction.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetByAccountIDWithFilters(accountID uuid.UUID, filters map[string]interface{}, limit, offset int) ([]*transaction.Transaction, error) {
	args := m.Called(accountID, filters, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*transaction.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) GetByIdempotencyKey(key string) (*transaction.Transaction, error) {
	args := m.Called(key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*transaction.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) ExecuteTransfer(from, to uuid.UUID, amount float64, txn *transaction.Transaction) error {
	args := m.Called(from, to, amount, txn)
	return args.Error(0)
}

func (m *MockTransactionRepository) ExecuteDeposit(accountID uuid.UUID, amount float64, txn *transaction.Transaction) error {
	args := m.Called(accountID, amount, txn)
	return args.Error(0)
}

func (m *MockTransactionRepository) ExecuteWithdrawal(accountID uuid.UUID, amount float64, txn *transaction.Transaction) error {
	args := m.Called(accountID, amount, txn)
	return args.Error(0)
}

func (m *MockTransactionRepository) UpdateStatus(id uuid.UUID, status transaction.TransactionStatus) error {
	args := m.Called(id, status)
	return args.Error(0)
}

// MockAuditRepository is a mock implementation
type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) Create(log *audit.AuditLog) error {
	args := m.Called(log)
	return args.Error(0)
}

func (m *MockAuditRepository) GetByUserID(userID uuid.UUID, limit, offset int) ([]*audit.AuditLog, error) {
	args := m.Called(userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*audit.AuditLog), args.Error(1)
}

func setupTransactionServiceTest(t *testing.T) (*transactionService, *MockTransactionRepository, *MockAccountRepository, *MockAuditRepository, *MockUserRepository) {
	logger.Init("test")
	txnRepo := new(MockTransactionRepository)
	accountRepo := new(MockAccountRepository)
	auditRepo := new(MockAuditRepository)
	userRepo := new(MockUserRepository)

	svc := NewTransactionService(txnRepo, accountRepo, auditRepo, userRepo).(*transactionService)
	return svc, txnRepo, accountRepo, auditRepo, userRepo
}

// ==================== Transfer Tests ====================

func TestTransfer_Success(t *testing.T) {
	svc, txnRepo, accountRepo, auditRepo, _ := setupTransactionServiceTest(t)
	userID := uuid.New()
	fromAccountID := uuid.New()
	toAccountID := uuid.New()

	req := &transaction.TransferRequest{
		FromAccountID:  fromAccountID.String(),
		ToAccountID:    toAccountID.String(),
		Amount:         100.00,
		Description:    "Test transfer",
		IdempotencyKey: "test-key-123",
	}

	// Mock idempotency check (not found)
	txnRepo.On("GetByIdempotencyKey", req.IdempotencyKey).Return(nil, fmt.Errorf("not found"))

	// Mock source account
	accountRepo.On("GetByID", fromAccountID).Return(&domainAccount.Account{
		ID:       fromAccountID,
		UserID:   userID,
		Currency: "USD",
		Balance:  500.00,
	}, nil)

	// Mock destination account
	accountRepo.On("GetByID", toAccountID).Return(&domainAccount.Account{
		ID:       toAccountID,
		UserID:   uuid.New(), // Different user
		Currency: "USD",
	}, nil)

	// Mock execute transfer
	txnRepo.On("ExecuteTransfer", fromAccountID, toAccountID, 100.00, mock.AnythingOfType("*transaction.Transaction")).Return(nil)

	// Mock audit log
	auditRepo.On("Create", mock.AnythingOfType("*audit.AuditLog")).Return(nil)

	// Mock get result
	completedTxn := &transaction.Transaction{
		ID:              uuid.New(),
		FromAccountID:   &fromAccountID,
		ToAccountID:     &toAccountID,
		Amount:          100.00,
		TransactionType: transaction.TransactionTypeTransfer,
		Status:          transaction.TransactionStatusCompleted,
	}
	txnRepo.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(completedTxn, nil)

	result, err := svc.Transfer(userID, req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 100.00, result.Amount)
	txnRepo.AssertExpectations(t)
}

func TestTransfer_Idempotency(t *testing.T) {
	svc, txnRepo, _, _, _ := setupTransactionServiceTest(t)
	userID := uuid.New()
	fromAccountID := uuid.New()
	toAccountID := uuid.New()

	existingTxn := &transaction.Transaction{
		ID:              uuid.New(),
		FromAccountID:   &fromAccountID,
		ToAccountID:     &toAccountID,
		Amount:          100.00,
		TransactionType: transaction.TransactionTypeTransfer,
		Status:          transaction.TransactionStatusCompleted,
	}

	req := &transaction.TransferRequest{
		FromAccountID:  fromAccountID.String(),
		ToAccountID:    toAccountID.String(),
		Amount:         100.00,
		IdempotencyKey: "duplicate-key",
	}

	// Mock idempotency check returns existing transaction
	txnRepo.On("GetByIdempotencyKey", req.IdempotencyKey).Return(existingTxn, nil)

	result, err := svc.Transfer(userID, req)
	assert.NoError(t, err)
	assert.Equal(t, existingTxn.ID, result.ID)
	// ExecuteTransfer should NOT be called
	txnRepo.AssertNotCalled(t, "ExecuteTransfer", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestTransfer_SameAccount(t *testing.T) {
	svc, _, _, _, _ := setupTransactionServiceTest(t)
	userID := uuid.New()
	accountID := uuid.New()

	req := &transaction.TransferRequest{
		FromAccountID: accountID.String(),
		ToAccountID:   accountID.String(), // Same account
		Amount:        100.00,
	}

	result, err := svc.Transfer(userID, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cannot transfer to the same account")
}

func TestTransfer_InvalidFromAccountID(t *testing.T) {
	svc, _, _, _, _ := setupTransactionServiceTest(t)
	userID := uuid.New()

	req := &transaction.TransferRequest{
		FromAccountID: "invalid",
		ToAccountID:   uuid.New().String(),
		Amount:        100.00,
	}

	result, err := svc.Transfer(userID, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid from_account_id")
}

func TestTransfer_Unauthorized(t *testing.T) {
	svc, txnRepo, accountRepo, _, _ := setupTransactionServiceTest(t)
	userID := uuid.New()
	otherUserID := uuid.New()
	fromAccountID := uuid.New()
	toAccountID := uuid.New()

	req := &transaction.TransferRequest{
		FromAccountID:  fromAccountID.String(),
		ToAccountID:    toAccountID.String(),
		Amount:         100.00,
		IdempotencyKey: "key",
	}

	txnRepo.On("GetByIdempotencyKey", req.IdempotencyKey).Return(nil, fmt.Errorf("not found"))

	// Source account belongs to different user
	accountRepo.On("GetByID", fromAccountID).Return(&domainAccount.Account{
		ID:     fromAccountID,
		UserID: otherUserID,
	}, nil)

	result, err := svc.Transfer(userID, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unauthorized")
}

func TestTransfer_CurrencyMismatch(t *testing.T) {
	svc, txnRepo, accountRepo, _, _ := setupTransactionServiceTest(t)
	userID := uuid.New()
	fromAccountID := uuid.New()
	toAccountID := uuid.New()

	req := &transaction.TransferRequest{
		FromAccountID:  fromAccountID.String(),
		ToAccountID:    toAccountID.String(),
		Amount:         100.00,
		IdempotencyKey: "key",
	}

	txnRepo.On("GetByIdempotencyKey", req.IdempotencyKey).Return(nil, fmt.Errorf("not found"))

	accountRepo.On("GetByID", fromAccountID).Return(&domainAccount.Account{
		ID:       fromAccountID,
		UserID:   userID,
		Currency: "USD",
	}, nil)

	accountRepo.On("GetByID", toAccountID).Return(&domainAccount.Account{
		ID:       toAccountID,
		UserID:   uuid.New(),
		Currency: "EUR", // Different currency
	}, nil)

	result, err := svc.Transfer(userID, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "currency mismatch")
}

// ==================== Deposit Tests ====================

func TestDeposit_Success(t *testing.T) {
	svc, txnRepo, accountRepo, auditRepo, _ := setupTransactionServiceTest(t)
	userID := uuid.New()
	accountID := uuid.New()

	req := &transaction.DepositRequest{
		AccountID:      accountID.String(),
		Amount:         500.00,
		Description:    "Salary deposit",
		IdempotencyKey: "deposit-key",
	}

	txnRepo.On("GetByIdempotencyKey", req.IdempotencyKey).Return(nil, fmt.Errorf("not found"))

	accountRepo.On("GetByID", accountID).Return(&domainAccount.Account{
		ID:       accountID,
		UserID:   userID,
		Currency: "USD",
	}, nil)

	txnRepo.On("ExecuteDeposit", accountID, 500.00, mock.AnythingOfType("*transaction.Transaction")).Return(nil)
	auditRepo.On("Create", mock.AnythingOfType("*audit.AuditLog")).Return(nil)

	completedTxn := &transaction.Transaction{
		ID:              uuid.New(),
		ToAccountID:     &accountID,
		Amount:          500.00,
		TransactionType: transaction.TransactionTypeDeposit,
		Status:          transaction.TransactionStatusCompleted,
	}
	txnRepo.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(completedTxn, nil)

	result, err := svc.Deposit(userID, req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 500.00, result.Amount)
}

func TestDeposit_Unauthorized(t *testing.T) {
	svc, txnRepo, accountRepo, _, _ := setupTransactionServiceTest(t)
	userID := uuid.New()
	otherUserID := uuid.New()
	accountID := uuid.New()

	req := &transaction.DepositRequest{
		AccountID:      accountID.String(),
		Amount:         500.00,
		IdempotencyKey: "key",
	}

	txnRepo.On("GetByIdempotencyKey", req.IdempotencyKey).Return(nil, fmt.Errorf("not found"))

	accountRepo.On("GetByID", accountID).Return(&domainAccount.Account{
		ID:     accountID,
		UserID: otherUserID, // Different user
	}, nil)

	result, err := svc.Deposit(userID, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unauthorized")
}

// ==================== Withdrawal Tests ====================

func TestWithdrawal_Success(t *testing.T) {
	svc, txnRepo, accountRepo, auditRepo, _ := setupTransactionServiceTest(t)
	userID := uuid.New()
	accountID := uuid.New()

	req := &transaction.WithdrawalRequest{
		AccountID:      accountID.String(),
		Amount:         200.00,
		Description:    "ATM withdrawal",
		IdempotencyKey: "withdraw-key",
	}

	txnRepo.On("GetByIdempotencyKey", req.IdempotencyKey).Return(nil, fmt.Errorf("not found"))

	accountRepo.On("GetByID", accountID).Return(&domainAccount.Account{
		ID:       accountID,
		UserID:   userID,
		Currency: "USD",
		Balance:  1000.00,
	}, nil)

	txnRepo.On("ExecuteWithdrawal", accountID, 200.00, mock.AnythingOfType("*transaction.Transaction")).Return(nil)
	auditRepo.On("Create", mock.AnythingOfType("*audit.AuditLog")).Return(nil)

	completedTxn := &transaction.Transaction{
		ID:              uuid.New(),
		FromAccountID:   &accountID,
		Amount:          200.00,
		TransactionType: transaction.TransactionTypeWithdrawal,
		Status:          transaction.TransactionStatusCompleted,
	}
	txnRepo.On("GetByID", mock.AnythingOfType("uuid.UUID")).Return(completedTxn, nil)

	result, err := svc.Withdrawal(userID, req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 200.00, result.Amount)
}

func TestWithdrawal_InsufficientFunds(t *testing.T) {
	svc, txnRepo, accountRepo, auditRepo, _ := setupTransactionServiceTest(t)
	userID := uuid.New()
	accountID := uuid.New()

	req := &transaction.WithdrawalRequest{
		AccountID:      accountID.String(),
		Amount:         2000.00, // More than balance
		IdempotencyKey: "key",
	}

	txnRepo.On("GetByIdempotencyKey", req.IdempotencyKey).Return(nil, fmt.Errorf("not found"))

	accountRepo.On("GetByID", accountID).Return(&domainAccount.Account{
		ID:       accountID,
		UserID:   userID,
		Currency: "USD",
		Balance:  500.00,
	}, nil)

	// ExecuteWithdrawal returns an error
	txnRepo.On("ExecuteWithdrawal", accountID, 2000.00, mock.AnythingOfType("*transaction.Transaction")).Return(fmt.Errorf("insufficient funds"))
	auditRepo.On("Create", mock.AnythingOfType("*audit.AuditLog")).Return(nil)

	result, err := svc.Withdrawal(userID, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "insufficient funds")
}

// ==================== GetTransactionHistory Tests ====================

func TestGetTransactionHistory_Success(t *testing.T) {
	svc, txnRepo, accountRepo, _, _ := setupTransactionServiceTest(t)
	userID := uuid.New()
	accountID := uuid.New()

	req := &transaction.TransactionHistoryRequest{
		AccountID: accountID.String(),
		Limit:     10,
		Offset:    0,
	}

	accountRepo.On("GetByID", accountID).Return(&domainAccount.Account{
		ID:     accountID,
		UserID: userID,
	}, nil)

	transactions := []*transaction.Transaction{
		{ID: uuid.New(), Amount: 100.00, TransactionType: transaction.TransactionTypeTransfer, CreatedAt: time.Now()},
		{ID: uuid.New(), Amount: 50.00, TransactionType: transaction.TransactionTypeDeposit, CreatedAt: time.Now()},
	}

	txnRepo.On("GetByAccountID", accountID, 10, 0).Return(transactions, nil)

	result, err := svc.GetTransactionHistory(userID, req)
	assert.NoError(t, err)
	assert.Len(t, result.Transactions, 2)
	assert.Equal(t, 10, result.Limit)
}

func TestGetTransactionHistory_WithFilters(t *testing.T) {
	svc, txnRepo, accountRepo, _, _ := setupTransactionServiceTest(t)
	userID := uuid.New()
	accountID := uuid.New()

	req := &transaction.TransactionHistoryRequest{
		AccountID: accountID.String(),
		TxnType:   "transfer",
		StartDate: "2025-01-01",
		EndDate:   "2025-12-31",
		Limit:     20,
	}

	accountRepo.On("GetByID", accountID).Return(&domainAccount.Account{
		ID:     accountID,
		UserID: userID,
	}, nil)

	transactions := []*transaction.Transaction{
		{ID: uuid.New(), Amount: 100.00, TransactionType: transaction.TransactionTypeTransfer},
	}

	txnRepo.On("GetByAccountIDWithFilters", accountID, mock.AnythingOfType("map[string]interface {}"), 20, 0).Return(transactions, nil)

	result, err := svc.GetTransactionHistory(userID, req)
	assert.NoError(t, err)
	assert.Len(t, result.Transactions, 1)
}

func TestGetTransactionHistory_LimitCap(t *testing.T) {
	svc, txnRepo, accountRepo, _, _ := setupTransactionServiceTest(t)
	userID := uuid.New()
	accountID := uuid.New()

	req := &transaction.TransactionHistoryRequest{
		AccountID: accountID.String(),
		Limit:     500, // Exceeds cap of 100
	}

	accountRepo.On("GetByID", accountID).Return(&domainAccount.Account{
		ID:     accountID,
		UserID: userID,
	}, nil)

	txnRepo.On("GetByAccountID", accountID, 100, 0).Return([]*transaction.Transaction{}, nil) // Capped at 100

	result, err := svc.GetTransactionHistory(userID, req)
	assert.NoError(t, err)
	assert.Equal(t, 100, result.Limit) // Should be capped
}

// ==================== GetTransaction Tests ====================

func TestGetTransaction_Success_FromAccount(t *testing.T) {
	svc, txnRepo, accountRepo, _, _ := setupTransactionServiceTest(t)
	userID := uuid.New()
	transactionID := uuid.New()
	fromAccountID := uuid.New()

	existingTxn := &transaction.Transaction{
		ID:            transactionID,
		FromAccountID: &fromAccountID,
		Amount:        100.00,
	}

	txnRepo.On("GetByID", transactionID).Return(existingTxn, nil)

	accountRepo.On("GetByID", fromAccountID).Return(&domainAccount.Account{
		ID:     fromAccountID,
		UserID: userID, // User owns this account
	}, nil)

	result, err := svc.GetTransaction(userID, transactionID)
	assert.NoError(t, err)
	assert.Equal(t, transactionID, result.ID)
}

func TestGetTransaction_Unauthorized(t *testing.T) {
	svc, txnRepo, accountRepo, _, _ := setupTransactionServiceTest(t)
	userID := uuid.New()
	otherUserID := uuid.New()
	transactionID := uuid.New()
	fromAccountID := uuid.New()
	toAccountID := uuid.New()

	existingTxn := &transaction.Transaction{
		ID:            transactionID,
		FromAccountID: &fromAccountID,
		ToAccountID:   &toAccountID,
		Amount:        100.00,
	}

	txnRepo.On("GetByID", transactionID).Return(existingTxn, nil)

	// Neither account belongs to user
	accountRepo.On("GetByID", fromAccountID).Return(&domainAccount.Account{
		ID:     fromAccountID,
		UserID: otherUserID,
	}, nil)
	accountRepo.On("GetByID", toAccountID).Return(&domainAccount.Account{
		ID:     toAccountID,
		UserID: otherUserID,
	}, nil)

	result, err := svc.GetTransaction(userID, transactionID)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unauthorized")
}

// ==================== ResolveQR Tests ====================

func TestResolveQR_Success(t *testing.T) {
	svc, _, accountRepo, _, userRepo := setupTransactionServiceTest(t)
	accountID := uuid.New()
	ownerID := uuid.New()
	qrCode := fmt.Sprintf("madabank:account:%s", accountID.String())

	accountRepo.On("GetByID", accountID).Return(&domainAccount.Account{
		ID:       accountID,
		UserID:   ownerID,
		Currency: "USD",
	}, nil)

	userRepo.On("GetByID", ownerID).Return(&user.User{
		ID:        ownerID,
		FirstName: "John",
		LastName:  "Doe",
	}, nil)

	result, err := svc.ResolveQR(qrCode)
	assert.NoError(t, err)
	assert.Equal(t, accountID, result.AccountID)
	assert.Equal(t, "John Doe", result.OwnerName)
	assert.Equal(t, "USD", result.Currency)
}

func TestResolveQR_InvalidFormat(t *testing.T) {
	svc, _, _, _, _ := setupTransactionServiceTest(t)

	result, err := svc.ResolveQR("invalid-qr-code")
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid QR code format")
}

func TestResolveQR_AccountNotFound(t *testing.T) {
	svc, _, accountRepo, _, _ := setupTransactionServiceTest(t)
	accountID := uuid.New()
	qrCode := fmt.Sprintf("madabank:account:%s", accountID.String())

	accountRepo.On("GetByID", accountID).Return(nil, fmt.Errorf("not found"))

	result, err := svc.ResolveQR(qrCode)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "account not found")
}
