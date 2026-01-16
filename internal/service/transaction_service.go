package service

import (
	"fmt"
	"time"

	"github.com/darisadam/madabank-server/internal/domain/audit"
	"github.com/darisadam/madabank-server/internal/domain/transaction"
	"github.com/darisadam/madabank-server/internal/pkg/logger"
	"github.com/darisadam/madabank-server/internal/pkg/metrics"
	"github.com/darisadam/madabank-server/internal/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TransactionService interface {
	Transfer(userID uuid.UUID, req *transaction.TransferRequest) (*transaction.Transaction, error)
	Deposit(userID uuid.UUID, req *transaction.DepositRequest) (*transaction.Transaction, error)
	Withdrawal(userID uuid.UUID, req *transaction.WithdrawalRequest) (*transaction.Transaction, error)
	GetTransactionHistory(userID uuid.UUID, req *transaction.TransactionHistoryRequest) (*transaction.TransactionHistoryResponse, error)
	GetTransaction(userID uuid.UUID, transactionID uuid.UUID) (*transaction.Transaction, error)
}

type transactionService struct {
	transactionRepo repository.TransactionRepository
	accountRepo     repository.AccountRepository
	auditRepo       repository.AuditRepository
}

func NewTransactionService(
	transactionRepo repository.TransactionRepository,
	accountRepo repository.AccountRepository,
	auditRepo repository.AuditRepository,
) TransactionService {
	return &transactionService{
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
		auditRepo:       auditRepo,
	}
}

func (s *transactionService) Transfer(userID uuid.UUID, req *transaction.TransferRequest) (*transaction.Transaction, error) {
	start := time.Now()

	// Parse UUIDs
	fromAccountID, err := uuid.Parse(req.FromAccountID)
	if err != nil {
		metrics.RecordTransactionError("transfer", "invalid_from_account")
		return nil, fmt.Errorf("invalid from_account_id")
	}

	toAccountID, err := uuid.Parse(req.ToAccountID)
	if err != nil {
		metrics.RecordTransactionError("transfer", "invalid_to_account")
		return nil, fmt.Errorf("invalid to_account_id")
	}

	// Validate accounts are different
	if fromAccountID == toAccountID {
		metrics.RecordTransactionError("transfer", "same_account")
		return nil, fmt.Errorf("cannot transfer to the same account")
	}

	// Check idempotency - prevent duplicate transfers
	existing, err := s.transactionRepo.GetByIdempotencyKey(req.IdempotencyKey)
	if err == nil {
		// Transaction already exists - record as successful (idempotency worked)
		duration := time.Since(start).Seconds()
		metrics.RecordTransaction("transfer", "completed", req.Amount, "USD", duration)
		return existing, nil
	}

	// Verify source account ownership
	fromAccount, err := s.accountRepo.GetByID(fromAccountID)
	if err != nil {
		metrics.RecordTransactionError("transfer", "source_not_found")
		return nil, fmt.Errorf("source account not found")
	}
	if fromAccount.UserID != userID {
		metrics.RecordTransactionError("transfer", "unauthorized")
		return nil, fmt.Errorf("unauthorized: source account does not belong to user")
	}

	// Verify destination account exists and is active
	toAccount, err := s.accountRepo.GetByID(toAccountID)
	if err != nil {
		metrics.RecordTransactionError("transfer", "destination_not_found")
		return nil, fmt.Errorf("destination account not found")
	}

	// Validate currency match
	if fromAccount.Currency != toAccount.Currency {
		metrics.RecordTransactionError("transfer", "currency_mismatch")
		return nil, fmt.Errorf("currency mismatch: source account is %s, destination is %s", fromAccount.Currency, toAccount.Currency)
	}

	// Create transaction object
	txn := &transaction.Transaction{
		ID:              uuid.New(),
		IdempotencyKey:  req.IdempotencyKey,
		FromAccountID:   &fromAccountID,
		ToAccountID:     &toAccountID,
		Amount:          req.Amount,
		TransactionType: transaction.TransactionTypeTransfer,
		Status:          transaction.TransactionStatusPending,
		Description:     req.Description,
		Metadata: map[string]interface{}{
			"initiated_by": userID.String(),
			"currency":     fromAccount.Currency,
		},
	}

	// Execute transfer with ACID guarantees
	err = s.transactionRepo.ExecuteTransfer(fromAccountID, toAccountID, req.Amount, txn)
	if err != nil {
		// Record failed transaction
		duration := time.Since(start).Seconds()
		metrics.RecordTransaction("transfer", "failed", req.Amount, fromAccount.Currency, duration)
		metrics.RecordTransactionError("transfer", "execution_failed")

		// Log failed transaction attempt
		if errAudit := s.auditRepo.Create(&audit.AuditLog{
			EventID:  uuid.New(),
			UserID:   &userID,
			Action:   "TRANSFER_FAILED",
			Resource: fmt.Sprintf("transaction:%s", txn.ID),
			Status:   "failed",
			Metadata: map[string]interface{}{
				"error":  err.Error(),
				"amount": req.Amount,
				"from":   req.FromAccountID,
				"to":     req.ToAccountID,
			},
		}); errAudit != nil {
			logger.Error("Failed to create audit log for failed transfer", zap.Error(errAudit))
		}
		return nil, err
	}

	// Record successful transaction
	duration := time.Since(start).Seconds()
	metrics.RecordTransaction("transfer", "completed", req.Amount, fromAccount.Currency, duration)

	// Log successful transaction
	if err := s.auditRepo.Create(&audit.AuditLog{
		EventID:  uuid.New(),
		UserID:   &userID,
		Action:   "TRANSFER_COMPLETED",
		Resource: fmt.Sprintf("transaction:%s", txn.ID),
		Status:   "success",
		Metadata: map[string]interface{}{
			"amount": req.Amount,
			"from":   req.FromAccountID,
			"to":     req.ToAccountID,
		},
	}); err != nil {
		logger.Error("Failed to create audit log for completed transfer", zap.Error(err))
	}

	// Retrieve the completed transaction
	return s.transactionRepo.GetByID(txn.ID)
}

func (s *transactionService) Deposit(userID uuid.UUID, req *transaction.DepositRequest) (*transaction.Transaction, error) {
	start := time.Now()

	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		metrics.RecordTransactionError("deposit", "invalid_account")
		return nil, fmt.Errorf("invalid account_id")
	}

	// Check idempotency
	existing, err := s.transactionRepo.GetByIdempotencyKey(req.IdempotencyKey)
	if err == nil {
		duration := time.Since(start).Seconds()
		metrics.RecordTransaction("deposit", "completed", req.Amount, "USD", duration)
		return existing, nil
	}

	// Verify account ownership
	account, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		metrics.RecordTransactionError("deposit", "account_not_found")
		return nil, fmt.Errorf("account not found")
	}
	if account.UserID != userID {
		metrics.RecordTransactionError("deposit", "unauthorized")
		return nil, fmt.Errorf("unauthorized: account does not belong to user")
	}

	// Create transaction
	txn := &transaction.Transaction{
		ID:              uuid.New(),
		IdempotencyKey:  req.IdempotencyKey,
		ToAccountID:     &accountID,
		Amount:          req.Amount,
		TransactionType: transaction.TransactionTypeDeposit,
		Status:          transaction.TransactionStatusPending,
		Description:     req.Description,
		Metadata: map[string]interface{}{
			"initiated_by": userID.String(),
			"currency":     account.Currency,
		},
	}

	// Execute deposit
	err = s.transactionRepo.ExecuteDeposit(accountID, req.Amount, txn)
	if err != nil {
		duration := time.Since(start).Seconds()
		metrics.RecordTransaction("deposit", "failed", req.Amount, account.Currency, duration)
		metrics.RecordTransactionError("deposit", "execution_failed")

		if errAudit := s.auditRepo.Create(&audit.AuditLog{
			EventID:  uuid.New(),
			UserID:   &userID,
			Action:   "DEPOSIT_FAILED",
			Resource: fmt.Sprintf("transaction:%s", txn.ID),
			Status:   "failed",
			Metadata: map[string]interface{}{
				"error":  err.Error(),
				"amount": req.Amount,
			},
		}); errAudit != nil {
			logger.Error("Failed to create audit log for failed deposit", zap.Error(errAudit))
		}
		return nil, err
	}

	duration := time.Since(start).Seconds()
	metrics.RecordTransaction("deposit", "completed", req.Amount, account.Currency, duration)

	if err := s.auditRepo.Create(&audit.AuditLog{
		EventID:  uuid.New(),
		UserID:   &userID,
		Action:   "DEPOSIT_COMPLETED",
		Resource: fmt.Sprintf("transaction:%s", txn.ID),
		Status:   "success",
		Metadata: map[string]interface{}{
			"amount": req.Amount,
		},
	}); err != nil {
		logger.Error("Failed to create audit log for completed deposit", zap.Error(err))
	}

	return s.transactionRepo.GetByID(txn.ID)
}

func (s *transactionService) Withdrawal(userID uuid.UUID, req *transaction.WithdrawalRequest) (*transaction.Transaction, error) {
	start := time.Now()

	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		metrics.RecordTransactionError("withdrawal", "invalid_account")
		return nil, fmt.Errorf("invalid account_id")
	}

	// Check idempotency
	existing, err := s.transactionRepo.GetByIdempotencyKey(req.IdempotencyKey)
	if err == nil {
		duration := time.Since(start).Seconds()
		metrics.RecordTransaction("withdrawal", "completed", req.Amount, "USD", duration)
		return existing, nil
	}

	// Verify account ownership
	account, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		metrics.RecordTransactionError("withdrawal", "account_not_found")
		return nil, fmt.Errorf("account not found")
	}
	if account.UserID != userID {
		metrics.RecordTransactionError("withdrawal", "unauthorized")
		return nil, fmt.Errorf("unauthorized: account does not belong to user")
	}

	// Create transaction
	txn := &transaction.Transaction{
		ID:              uuid.New(),
		IdempotencyKey:  req.IdempotencyKey,
		FromAccountID:   &accountID,
		Amount:          req.Amount,
		TransactionType: transaction.TransactionTypeWithdrawal,
		Status:          transaction.TransactionStatusPending,
		Description:     req.Description,
		Metadata: map[string]interface{}{
			"initiated_by": userID.String(),
			"currency":     account.Currency,
		},
	}

	// Execute withdrawal
	err = s.transactionRepo.ExecuteWithdrawal(accountID, req.Amount, txn)
	if err != nil {
		duration := time.Since(start).Seconds()
		metrics.RecordTransaction("withdrawal", "failed", req.Amount, account.Currency, duration)
		metrics.RecordTransactionError("withdrawal", "execution_failed")

		if errAudit := s.auditRepo.Create(&audit.AuditLog{
			EventID:  uuid.New(),
			UserID:   &userID,
			Action:   "WITHDRAWAL_FAILED",
			Resource: fmt.Sprintf("transaction:%s", txn.ID),
			Status:   "failed",
			Metadata: map[string]interface{}{
				"error":  err.Error(),
				"amount": req.Amount,
			},
		}); errAudit != nil {
			logger.Error("Failed to create audit log for failed withdrawal", zap.Error(errAudit))
		}
		return nil, err
	}

	duration := time.Since(start).Seconds()
	metrics.RecordTransaction("withdrawal", "completed", req.Amount, account.Currency, duration)

	if err := s.auditRepo.Create(&audit.AuditLog{
		EventID:  uuid.New(),
		UserID:   &userID,
		Action:   "WITHDRAWAL_COMPLETED",
		Resource: fmt.Sprintf("transaction:%s", txn.ID),
		Status:   "success",
		Metadata: map[string]interface{}{
			"amount": req.Amount,
		},
	}); err != nil {
		logger.Error("Failed to create audit log for completed withdrawal", zap.Error(err))
	}

	return s.transactionRepo.GetByID(txn.ID)
}

func (s *transactionService) GetTransactionHistory(userID uuid.UUID, req *transaction.TransactionHistoryRequest) (*transaction.TransactionHistoryResponse, error) {
	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		return nil, fmt.Errorf("invalid account_id")
	}

	// Verify account ownership
	account, err := s.accountRepo.GetByID(accountID)
	if err != nil {
		return nil, fmt.Errorf("account not found")
	}
	if account.UserID != userID {
		return nil, fmt.Errorf("unauthorized: account does not belong to user")
	}

	// Set defaults
	limit := req.Limit
	if limit == 0 {
		limit = 20
	}
	offset := req.Offset

	if limit > 100 {
		limit = 100
	}

	// Build filters
	filters := make(map[string]interface{})
	if req.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", req.StartDate)
		if err == nil {
			filters["start_date"] = startDate
		}
	}
	if req.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", req.EndDate)
		if err == nil {
			filters["end_date"] = endDate
		}
	}
	if req.TxnType != "" {
		filters["type"] = req.TxnType
	}

	// Get transactions
	var transactions []*transaction.Transaction
	if len(filters) > 0 {
		transactions, err = s.transactionRepo.GetByAccountIDWithFilters(accountID, filters, limit, offset)
	} else {
		transactions, err = s.transactionRepo.GetByAccountID(accountID, limit, offset)
	}
	if err != nil {
		return nil, err
	}

	// Convert to response format
	txnResponses := make([]transaction.TransactionResponse, len(transactions))
	for i, txn := range transactions {
		txnResponses[i] = transaction.TransactionResponse{
			ID:              txn.ID,
			FromAccountID:   txn.FromAccountID,
			ToAccountID:     txn.ToAccountID,
			Amount:          txn.Amount,
			TransactionType: txn.TransactionType,
			Status:          txn.Status,
			Description:     txn.Description,
			CreatedAt:       txn.CreatedAt,
			CompletedAt:     txn.CompletedAt,
		}
	}

	return &transaction.TransactionHistoryResponse{
		Transactions: txnResponses,
		Total:        len(txnResponses),
		Limit:        limit,
		Offset:       offset,
	}, nil
}

func (s *transactionService) GetTransaction(userID uuid.UUID, transactionID uuid.UUID) (*transaction.Transaction, error) {
	txn, err := s.transactionRepo.GetByID(transactionID)
	if err != nil {
		return nil, err
	}

	// Verify user has access to this transaction
	var hasAccess bool
	if txn.FromAccountID != nil {
		fromAccount, err := s.accountRepo.GetByID(*txn.FromAccountID)
		if err == nil && fromAccount.UserID == userID {
			hasAccess = true
		}
	}
	if txn.ToAccountID != nil {
		toAccount, err := s.accountRepo.GetByID(*txn.ToAccountID)
		if err == nil && toAccount.UserID == userID {
			hasAccess = true
		}
	}

	if !hasAccess {
		return nil, fmt.Errorf("unauthorized: transaction does not belong to user")
	}

	return txn, nil
}
