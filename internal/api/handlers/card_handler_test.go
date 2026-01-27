package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/darisadam/madabank-server/internal/domain/card"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCardService is a mock implementation of service.CardService
type MockCardService struct {
	mock.Mock
}

func (m *MockCardService) CreateCard(userID uuid.UUID, req *card.CreateCardRequest) (*card.CardResponse, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*card.CardResponse), args.Error(1)
}

func (m *MockCardService) GetUserCards(userID uuid.UUID, accountID uuid.UUID) ([]*card.CardResponse, error) {
	args := m.Called(userID, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*card.CardResponse), args.Error(1)
}

func (m *MockCardService) GetCardDetails(userID uuid.UUID, cardID uuid.UUID, password string) (*card.CardDetailsResponse, error) {
	args := m.Called(userID, cardID, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*card.CardDetailsResponse), args.Error(1)
}

func (m *MockCardService) UpdateCard(userID uuid.UUID, cardID uuid.UUID, req *card.UpdateCardRequest) (*card.CardResponse, error) {
	args := m.Called(userID, cardID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*card.CardResponse), args.Error(1)
}

func (m *MockCardService) BlockCard(userID uuid.UUID, cardID uuid.UUID) error {
	args := m.Called(userID, cardID)
	return args.Error(0)
}

func (m *MockCardService) DeleteCard(userID uuid.UUID, cardID uuid.UUID) error {
	args := m.Called(userID, cardID)
	return args.Error(0)
}

func setupCardRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// ==================== CreateCard Tests ====================

func TestCardHandler_CreateCard_Success(t *testing.T) {
	mockService := new(MockCardService)
	handler := NewCardHandler(mockService)

	router := setupCardRouter()
	userID := uuid.New()
	accountID := uuid.New()

	router.POST("/cards", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.CreateCard(c)
	})

	cardResp := &card.CardResponse{
		ID:               uuid.New(),
		CardNumberMasked: "****1234",
		CardHolderName:   "John Doe",
		CardType:         card.CardTypeDebit,
		Status:           card.CardStatusActive,
	}

	mockService.On("CreateCard", userID, mock.AnythingOfType("*card.CreateCardRequest")).Return(cardResp, nil)

	reqBody := `{"account_id":"` + accountID.String() + `","card_holder_name":"John Doe","card_type":"debit","daily_limit":5000000}`
	req, _ := http.NewRequest("POST", "/cards", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestCardHandler_CreateCard_Unauthorized(t *testing.T) {
	mockService := new(MockCardService)
	handler := NewCardHandler(mockService)

	router := setupCardRouter()
	router.POST("/cards", handler.CreateCard) // No user_id

	reqBody := `{"account_id":"` + uuid.New().String() + `","card_holder_name":"John Doe"}`
	req, _ := http.NewRequest("POST", "/cards", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ==================== GetCards Tests ====================

func TestCardHandler_GetCards_Success(t *testing.T) {
	mockService := new(MockCardService)
	handler := NewCardHandler(mockService)

	router := setupCardRouter()
	userID := uuid.New()
	accountID := uuid.New()

	router.GET("/cards", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.GetCards(c)
	})

	cards := []*card.CardResponse{
		{ID: uuid.New(), CardNumberMasked: "****1234"},
	}

	mockService.On("GetUserCards", userID, accountID).Return(cards, nil)

	req, _ := http.NewRequest("GET", "/cards?account_id="+accountID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCardHandler_GetCards_MissingAccountID(t *testing.T) {
	mockService := new(MockCardService)
	handler := NewCardHandler(mockService)

	router := setupCardRouter()
	userID := uuid.New()

	router.GET("/cards", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.GetCards(c)
	})

	req, _ := http.NewRequest("GET", "/cards", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "account_id is required")
}

// ==================== GetCardDetails Tests ====================

func TestCardHandler_GetCardDetails_Success(t *testing.T) {
	mockService := new(MockCardService)
	handler := NewCardHandler(mockService)

	router := setupCardRouter()
	userID := uuid.New()
	cardID := uuid.New()

	router.POST("/cards/details", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.GetCardDetails(c)
	})

	detailsResp := &card.CardDetailsResponse{
		CardNumber:  "4111111111111111",
		CVV:         "123",
		ExpiryMonth: 12,
		ExpiryYear:  2028,
	}

	mockService.On("GetCardDetails", userID, cardID, "password123").Return(detailsResp, nil)

	reqBody := `{"card_id":"` + cardID.String() + `","password":"password123"}`
	req, _ := http.NewRequest("POST", "/cards/details", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestCardHandler_GetCardDetails_WrongPassword(t *testing.T) {
	mockService := new(MockCardService)
	handler := NewCardHandler(mockService)

	router := setupCardRouter()
	userID := uuid.New()
	cardID := uuid.New()

	router.POST("/cards/details", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.GetCardDetails(c)
	})

	mockService.On("GetCardDetails", userID, cardID, "wrongpassword").Return(nil, assert.AnError)

	reqBody := `{"card_id":"` + cardID.String() + `","password":"wrongpassword"}`
	req, _ := http.NewRequest("POST", "/cards/details", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}

// ==================== UpdateCard Tests ====================

func TestCardHandler_UpdateCard_Success(t *testing.T) {
	mockService := new(MockCardService)
	handler := NewCardHandler(mockService)

	router := setupCardRouter()
	userID := uuid.New()
	cardID := uuid.New()

	router.PATCH("/cards/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.UpdateCard(c)
	})

	cardResp := &card.CardResponse{
		ID:         cardID,
		DailyLimit: 10000000,
	}

	mockService.On("UpdateCard", userID, cardID, mock.AnythingOfType("*card.UpdateCardRequest")).Return(cardResp, nil)

	reqBody := `{"daily_limit":10000000}`
	req, _ := http.NewRequest("PATCH", "/cards/"+cardID.String(), bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// ==================== BlockCard Tests ====================

func TestCardHandler_BlockCard_Success(t *testing.T) {
	mockService := new(MockCardService)
	handler := NewCardHandler(mockService)

	router := setupCardRouter()
	userID := uuid.New()
	cardID := uuid.New()

	router.POST("/cards/:id/block", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.BlockCard(c)
	})

	mockService.On("BlockCard", userID, cardID).Return(nil)

	req, _ := http.NewRequest("POST", "/cards/"+cardID.String()+"/block", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// ==================== DeleteCard Tests ====================

func TestCardHandler_DeleteCard_Success(t *testing.T) {
	mockService := new(MockCardService)
	handler := NewCardHandler(mockService)

	router := setupCardRouter()
	userID := uuid.New()
	cardID := uuid.New()

	router.DELETE("/cards/:id", func(c *gin.Context) {
		c.Set("user_id", userID)
		handler.DeleteCard(c)
	})

	mockService.On("DeleteCard", userID, cardID).Return(nil)

	req, _ := http.NewRequest("DELETE", "/cards/"+cardID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}

func TestCardHandler_DeleteCard_Unauthorized(t *testing.T) {
	mockService := new(MockCardService)
	handler := NewCardHandler(mockService)

	router := setupCardRouter()
	cardID := uuid.New()

	router.DELETE("/cards/:id", handler.DeleteCard) // No user_id

	req, _ := http.NewRequest("DELETE", "/cards/"+cardID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
