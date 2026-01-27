package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/darisadam/madabank-server/internal/domain/transaction"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTransactionService is a mock implementation of service.TransactionService
type MockTransactionService struct {
	mock.Mock
}

func (m *MockTransactionService) Transfer(userID uuid.UUID, req *transaction.TransferRequest) (*transaction.Transaction, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*transaction.Transaction), args.Error(1)
}

func (m *MockTransactionService) Deposit(userID uuid.UUID, req *transaction.DepositRequest) (*transaction.Transaction, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*transaction.Transaction), args.Error(1)
}

func (m *MockTransactionService) Withdrawal(userID uuid.UUID, req *transaction.WithdrawalRequest) (*transaction.Transaction, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*transaction.Transaction), args.Error(1)
}

func (m *MockTransactionService) GetTransactionHistory(userID uuid.UUID, req *transaction.TransactionHistoryRequest) (*transaction.TransactionHistoryResponse, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*transaction.TransactionHistoryResponse), args.Error(1)
}

func (m *MockTransactionService) GetTransaction(userID uuid.UUID, transactionID uuid.UUID) (*transaction.Transaction, error) {
	args := m.Called(userID, transactionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*transaction.Transaction), args.Error(1)
}

func (m *MockTransactionService) ResolveQR(qrCode string) (*transaction.QRResolutionResponse, error) {
	args := m.Called(qrCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*transaction.QRResolutionResponse), args.Error(1)
}

func setupTransactionRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// ==================== Transfer Tests ====================

func TestTransactionHandler_Transfer_Success(t *testing.T) {
	mockService := new(MockTransactionService)
	handler := NewTransactionHandler(mockService)

	router := setupTransactionRouter()
	userID := uuid.New()

	router.POST("/transfer", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.Transfer(c)
	})

	txn := &transaction.Transaction{
		ID:              uuid.New(),
		TransactionType: transaction.TransactionTypeTransfer,
		Amount:          100000,
		Status:          transaction.TransactionStatusCompleted,
	}

	mockService.On("Transfer", userID, mock.AnythingOfType("*transaction.TransferRequest")).Return(txn, nil)

	reqBody := `{"from_account_id":"` + uuid.New().String() + `","to_account_id":"` + uuid.New().String() + `","amount":100000,"idempotency_key":"test-key"}`
	req, _ := http.NewRequest("POST", "/transfer", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestTransactionHandler_Transfer_Error(t *testing.T) {
	mockService := new(MockTransactionService)
	handler := NewTransactionHandler(mockService)

	router := setupTransactionRouter()
	userID := uuid.New()

	router.POST("/transfer", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.Transfer(c)
	})

	mockService.On("Transfer", userID, mock.AnythingOfType("*transaction.TransferRequest")).Return(nil, assert.AnError)

	reqBody := `{"from_account_id":"` + uuid.New().String() + `","to_account_id":"` + uuid.New().String() + `","amount":100000000,"idempotency_key":"test-key"}`
	req, _ := http.NewRequest("POST", "/transfer", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestTransactionHandler_Transfer_Unauthorized(t *testing.T) {
	mockService := new(MockTransactionService)
	handler := NewTransactionHandler(mockService)

	router := setupTransactionRouter()
	router.POST("/transfer", handler.Transfer) // No user_id

	reqBody := `{"from_account_id":"` + uuid.New().String() + `","to_account_id":"` + uuid.New().String() + `","amount":100000,"idempotency_key":"test-key"}`
	req, _ := http.NewRequest("POST", "/transfer", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ==================== Deposit Tests ====================

func TestTransactionHandler_Deposit_Success(t *testing.T) {
	mockService := new(MockTransactionService)
	handler := NewTransactionHandler(mockService)

	router := setupTransactionRouter()
	userID := uuid.New()

	router.POST("/deposit", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.Deposit(c)
	})

	txn := &transaction.Transaction{
		ID:              uuid.New(),
		TransactionType: transaction.TransactionTypeDeposit,
		Amount:          500000,
		Status:          transaction.TransactionStatusCompleted,
	}

	mockService.On("Deposit", userID, mock.AnythingOfType("*transaction.DepositRequest")).Return(txn, nil)

	reqBody := `{"account_id":"` + uuid.New().String() + `","amount":500000,"idempotency_key":"test-key"}`
	req, _ := http.NewRequest("POST", "/deposit", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

// ==================== Withdraw Tests ====================

func TestTransactionHandler_Withdraw_Success(t *testing.T) {
	mockService := new(MockTransactionService)
	handler := NewTransactionHandler(mockService)

	router := setupTransactionRouter()
	userID := uuid.New()

	router.POST("/withdraw", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.Withdraw(c)
	})

	txn := &transaction.Transaction{
		ID:              uuid.New(),
		TransactionType: transaction.TransactionTypeWithdrawal,
		Amount:          200000,
		Status:          transaction.TransactionStatusCompleted,
	}

	mockService.On("Withdrawal", userID, mock.AnythingOfType("*transaction.WithdrawalRequest")).Return(txn, nil)

	reqBody := `{"account_id":"` + uuid.New().String() + `","amount":200000,"idempotency_key":"test-key"}`
	req, _ := http.NewRequest("POST", "/withdraw", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestTransactionHandler_Withdraw_Error(t *testing.T) {
	mockService := new(MockTransactionService)
	handler := NewTransactionHandler(mockService)

	router := setupTransactionRouter()
	userID := uuid.New()

	router.POST("/withdraw", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.Withdraw(c)
	})

	mockService.On("Withdrawal", userID, mock.AnythingOfType("*transaction.WithdrawalRequest")).Return(nil, assert.AnError)

	reqBody := `{"account_id":"` + uuid.New().String() + `","amount":200000,"idempotency_key":"test-key"}`
	req, _ := http.NewRequest("POST", "/withdraw", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

// ==================== GetHistory Tests ====================

func TestTransactionHandler_GetHistory_Success(t *testing.T) {
	mockService := new(MockTransactionService)
	handler := NewTransactionHandler(mockService)

	router := setupTransactionRouter()
	userID := uuid.New()
	accountID := uuid.New()

	router.GET("/transactions/history", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.GetHistory(c)
	})

	historyResp := &transaction.TransactionHistoryResponse{
		Transactions: []transaction.TransactionResponse{
			{ID: uuid.New(), Amount: 100000},
			{ID: uuid.New(), Amount: 200000},
		},
		Total:  2,
		Limit:  10,
		Offset: 0,
	}

	mockService.On("GetTransactionHistory", userID, mock.AnythingOfType("*transaction.TransactionHistoryRequest")).Return(historyResp, nil)

	req, _ := http.NewRequest("GET", "/transactions/history?account_id="+accountID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// ==================== GetTransaction Tests ====================

func TestTransactionHandler_GetTransaction_Success(t *testing.T) {
	mockService := new(MockTransactionService)
	handler := NewTransactionHandler(mockService)

	router := setupTransactionRouter()
	userID := uuid.New()
	transactionID := uuid.New()

	router.GET("/transactions/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.GetTransaction(c)
	})

	txn := &transaction.Transaction{
		ID:              transactionID,
		TransactionType: transaction.TransactionTypeTransfer,
		Amount:          100000,
	}

	mockService.On("GetTransaction", userID, transactionID).Return(txn, nil)

	req, _ := http.NewRequest("GET", "/transactions/"+transactionID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestTransactionHandler_GetTransaction_NotFound(t *testing.T) {
	mockService := new(MockTransactionService)
	handler := NewTransactionHandler(mockService)

	router := setupTransactionRouter()
	userID := uuid.New()
	transactionID := uuid.New()

	router.GET("/transactions/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.GetTransaction(c)
	})

	mockService.On("GetTransaction", userID, transactionID).Return(nil, assert.AnError)

	req, _ := http.NewRequest("GET", "/transactions/"+transactionID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// ==================== ResolveQR Tests ====================

func TestTransactionHandler_ResolveQR_Success(t *testing.T) {
	mockService := new(MockTransactionService)
	handler := NewTransactionHandler(mockService)

	router := setupTransactionRouter()
	userID := uuid.New()

	router.POST("/qr/resolve", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.ResolveQR(c)
	})

	qrResp := &transaction.QRResolutionResponse{
		AccountID: uuid.New(),
		OwnerName: "John Doe",
		Currency:  "IDR",
	}

	mockService.On("ResolveQR", "some-qr-code").Return(qrResp, nil)

	reqBody := `{"qr_code":"some-qr-code"}`
	req, _ := http.NewRequest("POST", "/qr/resolve", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}
