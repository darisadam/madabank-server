package service

import (
	"fmt"
	"testing"
	"time"

	domainAccount "github.com/darisadam/madabank-server/internal/domain/account"
	"github.com/darisadam/madabank-server/internal/domain/card"
	"github.com/darisadam/madabank-server/internal/domain/user"
	"github.com/darisadam/madabank-server/internal/pkg/crypto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCardRepository is a mock implementation of repository.CardRepository
type MockCardRepository struct {
	mock.Mock
}

func (m *MockCardRepository) Create(c *card.Card) error {
	args := m.Called(c)
	return args.Error(0)
}

func (m *MockCardRepository) GetByID(id uuid.UUID) (*card.Card, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*card.Card), args.Error(1)
}

func (m *MockCardRepository) GetByAccountID(accountID uuid.UUID) ([]*card.Card, error) {
	args := m.Called(accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*card.Card), args.Error(1)
}

func (m *MockCardRepository) Update(id uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return args.Error(0)
}

func (m *MockCardRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockCardRepository) GenerateCardNumber() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockCardRepository) GenerateCVV() string {
	args := m.Called()
	return args.String(0)
}

func setupCardServiceTest(t *testing.T) (*cardService, *MockCardRepository, *MockAccountRepository, *MockUserRepository) {
	cardRepo := new(MockCardRepository)
	accountRepo := new(MockAccountRepository)
	userRepo := new(MockUserRepository)

	encryptor, err := crypto.NewEncryptor("12345678901234567890123456789012") // 32 bytes
	assert.NoError(t, err)

	svc := NewCardService(cardRepo, accountRepo, userRepo, encryptor).(*cardService)
	return svc, cardRepo, accountRepo, userRepo
}

func TestCreateCard_Success(t *testing.T) {
	svc, cardRepo, accountRepo, _ := setupCardServiceTest(t)
	userID := uuid.New()
	accountID := uuid.New()

	req := &card.CreateCardRequest{
		AccountID:      accountID.String(),
		CardHolderName: "John Doe",
		CardType:       "debit",
		DailyLimit:     5000,
	}

	// Mock account ownership
	accountRepo.On("GetByID", accountID).Return(&domainAccount.Account{
		ID:     accountID,
		UserID: userID,
	}, nil)

	// Mock card generation
	cardRepo.On("GenerateCardNumber").Return("4111111111111111", nil)
	cardRepo.On("GenerateCVV").Return("123")
	cardRepo.On("Create", mock.AnythingOfType("*card.Card")).Return(nil)

	resp, err := svc.CreateCard(userID, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "John Doe", resp.CardHolderName)
	assert.Equal(t, card.CardStatusActive, resp.Status)
	assert.Contains(t, resp.CardNumberMasked, "****") // Should be masked
	cardRepo.AssertExpectations(t)
	accountRepo.AssertExpectations(t)
}

func TestCreateCard_InvalidAccountID(t *testing.T) {
	svc, _, _, _ := setupCardServiceTest(t)
	userID := uuid.New()

	req := &card.CreateCardRequest{
		AccountID: "invalid-uuid",
	}

	resp, err := svc.CreateCard(userID, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid account_id")
}

func TestCreateCard_AccountNotFound(t *testing.T) {
	svc, _, accountRepo, _ := setupCardServiceTest(t)
	userID := uuid.New()
	accountID := uuid.New()

	req := &card.CreateCardRequest{AccountID: accountID.String()}

	accountRepo.On("GetByID", accountID).Return(nil, fmt.Errorf("not found"))

	resp, err := svc.CreateCard(userID, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "account not found")
}

func TestCreateCard_Unauthorized(t *testing.T) {
	svc, _, accountRepo, _ := setupCardServiceTest(t)
	userID := uuid.New()
	otherUserID := uuid.New()
	accountID := uuid.New()

	req := &card.CreateCardRequest{AccountID: accountID.String()}

	// Account belongs to another user
	accountRepo.On("GetByID", accountID).Return(&domainAccount.Account{
		ID:     accountID,
		UserID: otherUserID,
	}, nil)

	resp, err := svc.CreateCard(userID, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "unauthorized")
}

func TestGetUserCards_Success(t *testing.T) {
	svc, cardRepo, accountRepo, _ := setupCardServiceTest(t)
	userID := uuid.New()
	accountID := uuid.New()

	// Mock account ownership
	accountRepo.On("GetByID", accountID).Return(&domainAccount.Account{
		ID:     accountID,
		UserID: userID,
	}, nil)

	// Encrypt a card number for the mock
	encryptedNumber, _ := svc.encryptor.Encrypt("4111111111111111")

	cards := []*card.Card{
		{
			ID:                  uuid.New(),
			AccountID:           accountID,
			CardNumberEncrypted: encryptedNumber,
			CardHolderName:      "John Doe",
			CardType:            card.CardTypeDebit,
			Status:              card.CardStatusActive,
			CreatedAt:           time.Now(),
		},
	}

	cardRepo.On("GetByAccountID", accountID).Return(cards, nil)

	result, err := svc.GetUserCards(userID, accountID)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Contains(t, result[0].CardNumberMasked, "****")
	cardRepo.AssertExpectations(t)
	accountRepo.AssertExpectations(t)
}

func TestGetCardDetails_Success(t *testing.T) {
	svc, cardRepo, accountRepo, userRepo := setupCardServiceTest(t)
	userID := uuid.New()
	cardID := uuid.New()
	accountID := uuid.New()
	password := "password123"
	passwordHash, _ := crypto.HashPassword(password)

	// Mock user verification
	userRepo.On("GetByID", userID).Return(&user.User{
		ID:           userID,
		PasswordHash: passwordHash,
	}, nil)

	// Encrypt card data
	encryptedNumber, _ := svc.encryptor.Encrypt("4111111111111111")
	encryptedCVV, _ := svc.encryptor.Encrypt("123")

	// Mock card retrieval
	cardRepo.On("GetByID", cardID).Return(&card.Card{
		ID:                  cardID,
		AccountID:           accountID,
		CardNumberEncrypted: encryptedNumber,
		CVVEncrypted:        encryptedCVV,
		ExpiryMonth:         12,
		ExpiryYear:          2027,
	}, nil)

	// Mock account ownership
	accountRepo.On("GetByID", accountID).Return(&domainAccount.Account{
		ID:     accountID,
		UserID: userID,
	}, nil)

	details, err := svc.GetCardDetails(userID, cardID, password)
	assert.NoError(t, err)
	assert.Equal(t, "4111111111111111", details.CardNumber)
	assert.Equal(t, "123", details.CVV)
	assert.Equal(t, 12, details.ExpiryMonth)
	assert.Equal(t, 2027, details.ExpiryYear)
}

func TestGetCardDetails_InvalidPassword(t *testing.T) {
	svc, _, _, userRepo := setupCardServiceTest(t)
	userID := uuid.New()
	cardID := uuid.New()
	correctHash, _ := crypto.HashPassword("correct")

	userRepo.On("GetByID", userID).Return(&user.User{
		ID:           userID,
		PasswordHash: correctHash,
	}, nil)

	details, err := svc.GetCardDetails(userID, cardID, "wrong")
	assert.Error(t, err)
	assert.Nil(t, details)
	assert.Contains(t, err.Error(), "invalid password")
}

func TestBlockCard_Success(t *testing.T) {
	svc, cardRepo, accountRepo, _ := setupCardServiceTest(t)
	userID := uuid.New()
	cardID := uuid.New()
	accountID := uuid.New()

	encryptedNumber, _ := svc.encryptor.Encrypt("4111111111111111")

	cardRepo.On("GetByID", cardID).Return(&card.Card{
		ID:                  cardID,
		AccountID:           accountID,
		CardNumberEncrypted: encryptedNumber,
		Status:              card.CardStatusActive,
	}, nil).Twice()

	accountRepo.On("GetByID", accountID).Return(&domainAccount.Account{
		ID:     accountID,
		UserID: userID,
	}, nil)

	cardRepo.On("Update", cardID, mock.AnythingOfType("map[string]interface {}")).Return(nil)

	err := svc.BlockCard(userID, cardID)
	assert.NoError(t, err)
	cardRepo.AssertExpectations(t)
}

func TestDeleteCard_Success(t *testing.T) {
	svc, cardRepo, accountRepo, _ := setupCardServiceTest(t)
	userID := uuid.New()
	cardID := uuid.New()
	accountID := uuid.New()

	cardRepo.On("GetByID", cardID).Return(&card.Card{
		ID:        cardID,
		AccountID: accountID,
	}, nil)

	accountRepo.On("GetByID", accountID).Return(&domainAccount.Account{
		ID:     accountID,
		UserID: userID,
	}, nil)

	cardRepo.On("Delete", cardID).Return(nil)

	err := svc.DeleteCard(userID, cardID)
	assert.NoError(t, err)
	cardRepo.AssertExpectations(t)
}

func TestDeleteCard_Unauthorized(t *testing.T) {
	svc, cardRepo, accountRepo, _ := setupCardServiceTest(t)
	userID := uuid.New()
	otherUserID := uuid.New()
	cardID := uuid.New()
	accountID := uuid.New()

	cardRepo.On("GetByID", cardID).Return(&card.Card{
		ID:        cardID,
		AccountID: accountID,
	}, nil)

	// Account belongs to another user
	accountRepo.On("GetByID", accountID).Return(&domainAccount.Account{
		ID:     accountID,
		UserID: otherUserID,
	}, nil)

	err := svc.DeleteCard(userID, cardID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
	cardRepo.AssertNotCalled(t, "Delete", mock.Anything)
}
