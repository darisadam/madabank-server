package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/darisadam/madabank-server/internal/domain/account"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAccountService is a mock implementation of service.AccountService
type MockAccountService struct {
	mock.Mock
}

func (m *MockAccountService) CreateAccount(userID uuid.UUID, req *account.CreateAccountRequest) (*account.Account, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*account.Account), args.Error(1)
}

func (m *MockAccountService) GetAccount(accountID uuid.UUID, userID uuid.UUID) (*account.Account, error) {
	args := m.Called(accountID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*account.Account), args.Error(1)
}

func (m *MockAccountService) GetAccountByNumber(accountNumber string, userID uuid.UUID) (*account.Account, error) {
	args := m.Called(accountNumber, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*account.Account), args.Error(1)
}

func (m *MockAccountService) GetUserAccounts(userID uuid.UUID) ([]*account.Account, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*account.Account), args.Error(1)
}

func (m *MockAccountService) GetBalance(accountID uuid.UUID, userID uuid.UUID) (*account.BalanceResponse, error) {
	args := m.Called(accountID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*account.BalanceResponse), args.Error(1)
}

func (m *MockAccountService) UpdateAccount(accountID uuid.UUID, userID uuid.UUID, req *account.UpdateAccountRequest) (*account.Account, error) {
	args := m.Called(accountID, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*account.Account), args.Error(1)
}

func (m *MockAccountService) CloseAccount(accountID uuid.UUID, userID uuid.UUID) error {
	args := m.Called(accountID, userID)
	return args.Error(0)
}

func setupAccountRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// ==================== CreateAccount Tests ====================

func TestAccountHandler_CreateAccount_Success(t *testing.T) {
	mockService := new(MockAccountService)
	handler := NewAccountHandler(mockService)

	router := setupAccountRouter()
	userID := uuid.New()

	router.POST("/accounts", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.CreateAccount(c)
	})

	expectedAccount := &account.Account{
		ID:            uuid.New(),
		UserID:        userID,
		AccountNumber: "1234567890",
		AccountType:   account.AccountTypeChecking,
		Balance:       0,
		Currency:      "IDR",
		Status:        account.AccountStatusActive,
	}

	mockService.On("CreateAccount", userID, mock.AnythingOfType("*account.CreateAccountRequest")).Return(expectedAccount, nil)

	reqBody := `{"account_type":"checking","currency":"IDR"}`
	req, _ := http.NewRequest("POST", "/accounts", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestAccountHandler_CreateAccount_Unauthorized(t *testing.T) {
	mockService := new(MockAccountService)
	handler := NewAccountHandler(mockService)

	router := setupAccountRouter()
	router.POST("/accounts", handler.CreateAccount) // No user_id

	reqBody := `{"account_type":"checking","currency":"IDR"}`
	req, _ := http.NewRequest("POST", "/accounts", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAccountHandler_CreateAccount_MaxReached(t *testing.T) {
	mockService := new(MockAccountService)
	handler := NewAccountHandler(mockService)

	router := setupAccountRouter()
	userID := uuid.New()

	router.POST("/accounts", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.CreateAccount(c)
	})

	mockService.On("CreateAccount", userID, mock.AnythingOfType("*account.CreateAccountRequest")).
		Return(nil, assert.AnError)

	reqBody := `{"account_type":"checking","currency":"IDR"}`
	req, _ := http.NewRequest("POST", "/accounts", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

// ==================== GetAccounts Tests ====================

func TestAccountHandler_GetAccounts_Success(t *testing.T) {
	mockService := new(MockAccountService)
	handler := NewAccountHandler(mockService)

	router := setupAccountRouter()
	userID := uuid.New()

	router.GET("/accounts", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.GetAccounts(c)
	})

	accounts := []*account.Account{
		{ID: uuid.New(), AccountNumber: "1111111111", Balance: 1000},
		{ID: uuid.New(), AccountNumber: "2222222222", Balance: 2000},
	}

	mockService.On("GetUserAccounts", userID).Return(accounts, nil)

	req, _ := http.NewRequest("GET", "/accounts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response account.AccountListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Accounts, 2)
	mockService.AssertExpectations(t)
}

// ==================== GetAccountByID Tests ====================

func TestAccountHandler_GetAccountByID_Success(t *testing.T) {
	mockService := new(MockAccountService)
	handler := NewAccountHandler(mockService)

	router := setupAccountRouter()
	userID := uuid.New()
	accountID := uuid.New()

	router.GET("/accounts/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.GetAccount(c)
	})

	expectedAccount := &account.Account{
		ID:            accountID,
		UserID:        userID,
		AccountNumber: "1234567890",
		Balance:       5000,
	}

	mockService.On("GetAccount", accountID, userID).Return(expectedAccount, nil)

	req, _ := http.NewRequest("GET", "/accounts/"+accountID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestAccountHandler_GetAccountByID_NotFound(t *testing.T) {
	mockService := new(MockAccountService)
	handler := NewAccountHandler(mockService)

	router := setupAccountRouter()
	userID := uuid.New()
	accountID := uuid.New()

	router.GET("/accounts/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.GetAccount(c)
	})

	mockService.On("GetAccount", accountID, userID).Return(nil, assert.AnError)

	req, _ := http.NewRequest("GET", "/accounts/"+accountID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// ==================== GetBalance Tests ====================

func TestAccountHandler_GetBalance_Success(t *testing.T) {
	mockService := new(MockAccountService)
	handler := NewAccountHandler(mockService)

	router := setupAccountRouter()
	userID := uuid.New()
	accountID := uuid.New()

	router.GET("/accounts/:id/balance", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.GetBalance(c)
	})

	balanceResponse := &account.BalanceResponse{
		AccountID: accountID,
		Balance:   10000,
		Currency:  "IDR",
	}

	mockService.On("GetBalance", accountID, userID).Return(balanceResponse, nil)

	req, _ := http.NewRequest("GET", "/accounts/"+accountID.String()+"/balance", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// ==================== UpdateAccount Tests ====================

func TestAccountHandler_UpdateAccount_Success(t *testing.T) {
	mockService := new(MockAccountService)
	handler := NewAccountHandler(mockService)

	router := setupAccountRouter()
	userID := uuid.New()
	accountID := uuid.New()

	router.PUT("/accounts/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.UpdateAccount(c)
	})

	updatedAccount := &account.Account{
		ID:     accountID,
		Status: account.AccountStatusFrozen,
	}

	mockService.On("UpdateAccount", accountID, userID, mock.AnythingOfType("*account.UpdateAccountRequest")).Return(updatedAccount, nil)

	reqBody := `{"status":"frozen"}`
	req, _ := http.NewRequest("PUT", "/accounts/"+accountID.String(), bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// ==================== CloseAccount Tests ====================

func TestAccountHandler_CloseAccount_Success(t *testing.T) {
	mockService := new(MockAccountService)
	handler := NewAccountHandler(mockService)

	router := setupAccountRouter()
	userID := uuid.New()
	accountID := uuid.New()

	router.DELETE("/accounts/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.CloseAccount(c)
	})

	mockService.On("CloseAccount", accountID, userID).Return(nil)

	req, _ := http.NewRequest("DELETE", "/accounts/"+accountID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}

func TestAccountHandler_CloseAccount_NonZeroBalance(t *testing.T) {
	mockService := new(MockAccountService)
	handler := NewAccountHandler(mockService)

	router := setupAccountRouter()
	userID := uuid.New()
	accountID := uuid.New()

	router.DELETE("/accounts/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.CloseAccount(c)
	})

	mockService.On("CloseAccount", accountID, userID).Return(assert.AnError)

	req, _ := http.NewRequest("DELETE", "/accounts/"+accountID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}
