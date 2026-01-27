package service

import (
	"fmt"
	"time"

	"github.com/darisadam/madabank-server/internal/domain/card"
	"github.com/darisadam/madabank-server/internal/pkg/crypto"
	"github.com/darisadam/madabank-server/internal/repository"
	"github.com/google/uuid"
)

type CardService interface {
	CreateCard(userID uuid.UUID, req *card.CreateCardRequest) (*card.CardResponse, error)
	GetUserCards(userID uuid.UUID, accountID uuid.UUID) ([]*card.CardResponse, error)
	GetCardDetails(userID uuid.UUID, cardID uuid.UUID, password string) (*card.CardDetailsResponse, error)
	UpdateCard(userID uuid.UUID, cardID uuid.UUID, req *card.UpdateCardRequest) (*card.CardResponse, error)
	BlockCard(userID uuid.UUID, cardID uuid.UUID) error
	DeleteCard(userID uuid.UUID, cardID uuid.UUID) error
}

type cardService struct {
	cardRepo    repository.CardRepository
	accountRepo repository.AccountRepository
	userRepo    repository.UserRepository
	encryptor   *crypto.Encryptor
}

func NewCardService(
	cardRepo repository.CardRepository,
	accountRepo repository.AccountRepository,
	userRepo repository.UserRepository,
	encryptor *crypto.Encryptor,
) CardService {
	return &cardService{
		cardRepo:    cardRepo,
		accountRepo: accountRepo,
		userRepo:    userRepo,
		encryptor:   encryptor,
	}
}

func (s *cardService) CreateCard(userID uuid.UUID, req *card.CreateCardRequest) (*card.CardResponse, error) {
	// Parse and verify account ownership
	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		return nil, fmt.Errorf("invalid account_id")
	}

	account, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		return nil, fmt.Errorf("account not found")
	}

	if account.UserID != userID {
		return nil, fmt.Errorf("unauthorized: account does not belong to user")
	}

	// Check one card per account limit
	existingCards, err := s.cardRepo.GetByAccountID(accountID)
	if err == nil && len(existingCards) >= 1 {
		return nil, fmt.Errorf("each account can only have one debit card")
	}

	// Generate card number and CVV
	cardNumber, err := s.cardRepo.GenerateCardNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate card number: %w", err)
	}

	cvv := s.cardRepo.GenerateCVV()

	// Encrypt sensitive data
	encryptedCardNumber, err := s.encryptor.Encrypt(cardNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt card number: %w", err)
	}

	encryptedCVV, err := s.encryptor.Encrypt(cvv)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt CVV: %w", err)
	}

	// Set expiry date (3 years from now)
	now := time.Now()
	expiryDate := now.AddDate(3, 0, 0)

	// Create card
	newCard := &card.Card{
		ID:                  uuid.New(),
		AccountID:           accountID,
		CardNumberEncrypted: encryptedCardNumber,
		CVVEncrypted:        encryptedCVV,
		CardHolderName:      req.CardHolderName,
		CardType:            card.CardType(req.CardType),
		ExpiryMonth:         int(expiryDate.Month()),
		ExpiryYear:          expiryDate.Year(),
		Status:              card.CardStatusActive,
		DailyLimit:          req.DailyLimit,
	}

	if err := s.cardRepo.Create(newCard); err != nil {
		return nil, fmt.Errorf("failed to create card: %w", err)
	}

	return s.toCardResponse(newCard, cardNumber), nil
}

func (s *cardService) GetUserCards(userID uuid.UUID, accountID uuid.UUID) ([]*card.CardResponse, error) {
	// Verify account ownership
	account, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		return nil, fmt.Errorf("account not found")
	}

	if account.UserID != userID {
		return nil, fmt.Errorf("unauthorized: account does not belong to user")
	}

	cards, err := s.cardRepo.GetByAccountID(accountID)
	if err != nil {
		return nil, err
	}

	responses := make([]*card.CardResponse, len(cards))
	for i, c := range cards {
		// Decrypt card number to mask it
		cardNumber, err := s.encryptor.Decrypt(c.CardNumberEncrypted)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt card number: %w", err)
		}
		responses[i] = s.toCardResponse(c, cardNumber)
	}

	return responses, nil
}

func (s *cardService) GetCardDetails(userID uuid.UUID, cardID uuid.UUID, password string) (*card.CardDetailsResponse, error) {
	// Verify user password for additional security
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	if !crypto.CheckPassword(password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid password")
	}

	// Get card
	c, err := s.cardRepo.GetByID(cardID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	account, err := s.accountRepo.GetByID(c.AccountID)
	if err != nil {
		return nil, fmt.Errorf("account not found")
	}

	if account.UserID != userID {
		return nil, fmt.Errorf("unauthorized: card does not belong to user")
	}

	// Decrypt sensitive data
	cardNumber, err := s.encryptor.Decrypt(c.CardNumberEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt card number: %w", err)
	}

	cvv, err := s.encryptor.Decrypt(c.CVVEncrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt CVV: %w", err)
	}

	return &card.CardDetailsResponse{
		CardNumber:  cardNumber,
		CVV:         cvv,
		ExpiryMonth: c.ExpiryMonth,
		ExpiryYear:  c.ExpiryYear,
	}, nil
}

func (s *cardService) UpdateCard(userID uuid.UUID, cardID uuid.UUID, req *card.UpdateCardRequest) (*card.CardResponse, error) {
	// Get and verify ownership
	c, err := s.cardRepo.GetByID(cardID)
	if err != nil {
		return nil, err
	}

	account, err := s.accountRepo.GetByID(c.AccountID)
	if err != nil {
		return nil, fmt.Errorf("account not found")
	}

	if account.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}

	// Build updates
	updates := make(map[string]interface{})

	if req.Status != nil {
		updates["status"] = card.CardStatus(*req.Status)
	}

	if req.DailyLimit != nil {
		updates["daily_limit"] = *req.DailyLimit
	}

	if len(updates) == 0 {
		// Decrypt to return response
		cardNumber, _ := s.encryptor.Decrypt(c.CardNumberEncrypted)
		return s.toCardResponse(c, cardNumber), nil
	}

	if err := s.cardRepo.Update(cardID, updates); err != nil {
		return nil, err
	}

	// Get updated card
	updatedCard, err := s.cardRepo.GetByID(cardID)
	if err != nil {
		return nil, err
	}

	cardNumber, _ := s.encryptor.Decrypt(updatedCard.CardNumberEncrypted)
	return s.toCardResponse(updatedCard, cardNumber), nil
}

func (s *cardService) BlockCard(userID uuid.UUID, cardID uuid.UUID) error {
	req := &card.UpdateCardRequest{
		Status: stringPtr("blocked"),
	}
	_, err := s.UpdateCard(userID, cardID, req)
	return err
}

func (s *cardService) DeleteCard(userID uuid.UUID, cardID uuid.UUID) error {
	// Verify ownership
	c, err := s.cardRepo.GetByID(cardID)
	if err != nil {
		return err
	}

	account, err := s.accountRepo.GetByID(c.AccountID)
	if err != nil {
		return fmt.Errorf("account not found")
	}

	if account.UserID != userID {
		return fmt.Errorf("unauthorized")
	}

	return s.cardRepo.Delete(cardID)
}

func (s *cardService) toCardResponse(c *card.Card, cardNumber string) *card.CardResponse {
	return &card.CardResponse{
		ID:               c.ID,
		AccountID:        c.AccountID,
		CardNumberMasked: crypto.MaskCardNumber(cardNumber),
		CardHolderName:   c.CardHolderName,
		CardType:         c.CardType,
		ExpiryMonth:      c.ExpiryMonth,
		ExpiryYear:       c.ExpiryYear,
		Status:           c.Status,
		DailyLimit:       c.DailyLimit,
		CreatedAt:        c.CreatedAt,
	}
}

func stringPtr(s string) *string {
	return &s
}
