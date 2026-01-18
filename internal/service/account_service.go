package service

import (
	"fmt"
	"time"

	"github.com/darisadam/madabank-server/internal/domain/account"
	"github.com/darisadam/madabank-server/internal/repository"
	"github.com/google/uuid"
)

type AccountService interface {
	CreateAccount(userID uuid.UUID, req *account.CreateAccountRequest) (*account.Account, error)
	GetAccount(accountID uuid.UUID, userID uuid.UUID) (*account.Account, error)
	GetAccountByNumber(accountNumber string, userID uuid.UUID) (*account.Account, error)
	GetUserAccounts(userID uuid.UUID) ([]*account.Account, error)
	GetBalance(accountID uuid.UUID, userID uuid.UUID) (*account.BalanceResponse, error)
	UpdateAccount(accountID uuid.UUID, userID uuid.UUID, req *account.UpdateAccountRequest) (*account.Account, error)
	CloseAccount(accountID uuid.UUID, userID uuid.UUID) error
}

type accountService struct {
	accountRepo repository.AccountRepository
}

func NewAccountService(accountRepo repository.AccountRepository) AccountService {
	return &accountService{
		accountRepo: accountRepo,
	}
}

func (s *accountService) CreateAccount(userID uuid.UUID, req *account.CreateAccountRequest) (*account.Account, error) {
	// Validate account type
	var accountType account.AccountType
	switch req.AccountType {
	case "checking":
		accountType = account.AccountTypeChecking
	case "savings":
		accountType = account.AccountTypeSavings
	default:
		return nil, fmt.Errorf("invalid account type")
	}

	// Set default interest rate for savings accounts
	interestRate := req.InterestRate
	if accountType == account.AccountTypeSavings && interestRate == 0 {
		interestRate = 0.0325 // 3.25% default
	}

	// Generate unique account number
	accountNumber, err := s.accountRepo.GenerateAccountNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to generate account number: %w", err)
	}

	// Create account
	newAccount := &account.Account{
		ID:            uuid.New(),
		UserID:        userID,
		AccountNumber: accountNumber,
		AccountType:   accountType,
		Balance:       0.00,
		Currency:      req.Currency,
		InterestRate:  interestRate,
		Status:        account.AccountStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.accountRepo.Create(newAccount); err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return newAccount, nil
}

func (s *accountService) GetAccount(accountID uuid.UUID, userID uuid.UUID) (*account.Account, error) {
	acc, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		return nil, err
	}

	// Verify account belongs to user
	if acc.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to account")
	}

	return acc, nil
}

func (s *accountService) GetAccountByNumber(accountNumber string, userID uuid.UUID) (*account.Account, error) {
	acc, err := s.accountRepo.GetByAccountNumber(accountNumber)
	if err != nil {
		return nil, err
	}

	// Verify account belongs to user
	if acc.UserID != userID {
		return nil, fmt.Errorf("unauthorized access to account")
	}

	return acc, nil
}

func (s *accountService) GetUserAccounts(userID uuid.UUID) ([]*account.Account, error) {
	return s.accountRepo.GetByUserID(userID)
}

func (s *accountService) GetBalance(accountID uuid.UUID, userID uuid.UUID) (*account.BalanceResponse, error) {
	acc, err := s.GetAccount(accountID, userID)
	if err != nil {
		return nil, err
	}

	return &account.BalanceResponse{
		AccountID:     acc.ID,
		AccountNumber: acc.AccountNumber,
		Balance:       acc.Balance,
		Currency:      acc.Currency,
		AsOfDate:      acc.UpdatedAt,
	}, nil
}

func (s *accountService) UpdateAccount(accountID uuid.UUID, userID uuid.UUID, req *account.UpdateAccountRequest) (*account.Account, error) {
	// Verify ownership
	acc, err := s.GetAccount(accountID, userID)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]interface{})

	if req.Status != nil {
		// Validate status transition
		newStatus := account.AccountStatus(*req.Status)
		if err := s.validateStatusTransition(acc.Status, newStatus); err != nil {
			return nil, err
		}
		updates["status"] = newStatus
	}

	if len(updates) == 0 {
		return acc, nil
	}

	if err := s.accountRepo.Update(accountID, updates); err != nil {
		return nil, err
	}

	return s.GetAccount(accountID, userID)
}

func (s *accountService) CloseAccount(accountID uuid.UUID, userID uuid.UUID) error {
	// Verify ownership
	acc, err := s.GetAccount(accountID, userID)
	if err != nil {
		return err
	}

	// Check if balance is zero
	if acc.Balance > 0 {
		return fmt.Errorf("cannot close account with non-zero balance. Current balance: %.2f %s", acc.Balance, acc.Currency)
	}

	return s.accountRepo.Delete(accountID)
}

func (s *accountService) validateStatusTransition(current, new account.AccountStatus) error {
	// Define valid status transitions
	validTransitions := map[account.AccountStatus][]account.AccountStatus{
		account.AccountStatusActive: {account.AccountStatusFrozen, account.AccountStatusClosed},
		account.AccountStatusFrozen: {account.AccountStatusActive, account.AccountStatusClosed},
		account.AccountStatusClosed: {}, // Closed accounts cannot transition
	}

	allowed, exists := validTransitions[current]
	if !exists {
		return fmt.Errorf("invalid current status")
	}

	for _, status := range allowed {
		if status == new {
			return nil
		}
	}

	return fmt.Errorf("cannot transition from %s to %s", current, new)
}
