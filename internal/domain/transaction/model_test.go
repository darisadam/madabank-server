package transaction

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTransactionType_Constants(t *testing.T) {
	assert.Equal(t, TransactionType("transfer"), TransactionTypeTransfer)
	assert.Equal(t, TransactionType("deposit"), TransactionTypeDeposit)
	assert.Equal(t, TransactionType("withdrawal"), TransactionTypeWithdrawal)
	assert.Equal(t, TransactionType("interest"), TransactionTypeInterest)
	assert.Equal(t, TransactionType("fee"), TransactionTypeFee)
}

func TestTransactionStatus_Constants(t *testing.T) {
	assert.Equal(t, TransactionStatus("pending"), TransactionStatusPending)
	assert.Equal(t, TransactionStatus("completed"), TransactionStatusCompleted)
	assert.Equal(t, TransactionStatus("failed"), TransactionStatusFailed)
	assert.Equal(t, TransactionStatus("reversed"), TransactionStatusReversed)
}

func TestTransaction_Structure(t *testing.T) {
	txnID := uuid.New()
	fromID := uuid.New()
	toID := uuid.New()

	txn := Transaction{
		ID:              txnID,
		IdempotencyKey:  "unique-key-123",
		FromAccountID:   &fromID,
		ToAccountID:     &toID,
		Amount:          250.50,
		TransactionType: TransactionTypeTransfer,
		Status:          TransactionStatusPending,
		Description:     "Test transfer",
		Metadata: map[string]interface{}{
			"note": "unit test",
		},
		CreatedAt: time.Now(),
	}

	assert.Equal(t, txnID, txn.ID)
	assert.Equal(t, "unique-key-123", txn.IdempotencyKey)
	assert.Equal(t, &fromID, txn.FromAccountID)
	assert.Equal(t, &toID, txn.ToAccountID)
	assert.Equal(t, 250.50, txn.Amount)
	assert.Equal(t, TransactionTypeTransfer, txn.TransactionType)
	assert.Equal(t, TransactionStatusPending, txn.Status)
}

func TestTransaction_OptionalAccountIDs(t *testing.T) {
	// Deposit only has ToAccountID
	toID := uuid.New()
	deposit := Transaction{
		ID:              uuid.New(),
		ToAccountID:     &toID,
		Amount:          100.00,
		TransactionType: TransactionTypeDeposit,
	}

	assert.Nil(t, deposit.FromAccountID)
	assert.NotNil(t, deposit.ToAccountID)

	// Withdrawal only has FromAccountID
	fromID := uuid.New()
	withdrawal := Transaction{
		ID:              uuid.New(),
		FromAccountID:   &fromID,
		Amount:          50.00,
		TransactionType: TransactionTypeWithdrawal,
	}

	assert.NotNil(t, withdrawal.FromAccountID)
	assert.Nil(t, withdrawal.ToAccountID)
}

func TestTransferRequest_Structure(t *testing.T) {
	req := TransferRequest{
		FromAccountID:  uuid.New().String(),
		ToAccountID:    uuid.New().String(),
		Amount:         500.00,
		Description:    "Payment for services",
		IdempotencyKey: "transfer-key-456",
	}

	assert.NotEmpty(t, req.FromAccountID)
	assert.NotEmpty(t, req.ToAccountID)
	assert.Equal(t, 500.00, req.Amount)
	assert.Equal(t, "transfer-key-456", req.IdempotencyKey)
}

func TestDepositRequest_Structure(t *testing.T) {
	req := DepositRequest{
		AccountID:      uuid.New().String(),
		Amount:         1000.00,
		Description:    "Salary",
		IdempotencyKey: "deposit-key-789",
	}

	assert.NotEmpty(t, req.AccountID)
	assert.Equal(t, 1000.00, req.Amount)
}

func TestWithdrawalRequest_Structure(t *testing.T) {
	req := WithdrawalRequest{
		AccountID:      uuid.New().String(),
		Amount:         200.00,
		Description:    "ATM",
		IdempotencyKey: "withdrawal-key-101",
	}

	assert.NotEmpty(t, req.AccountID)
	assert.Equal(t, 200.00, req.Amount)
}

func TestTransactionHistoryRequest_Defaults(t *testing.T) {
	req := TransactionHistoryRequest{
		AccountID: uuid.New().String(),
	}

	// Limit and Offset should default to zero
	assert.Equal(t, 0, req.Limit)
	assert.Equal(t, 0, req.Offset)
	assert.Empty(t, req.StartDate)
	assert.Empty(t, req.EndDate)
	assert.Empty(t, req.TxnType)
}

func TestTransactionHistoryRequest_WithFilters(t *testing.T) {
	req := TransactionHistoryRequest{
		AccountID: uuid.New().String(),
		Limit:     50,
		Offset:    10,
		StartDate: "2025-01-01",
		EndDate:   "2025-12-31",
		TxnType:   "transfer",
	}

	assert.Equal(t, 50, req.Limit)
	assert.Equal(t, 10, req.Offset)
	assert.Equal(t, "2025-01-01", req.StartDate)
	assert.Equal(t, "2025-12-31", req.EndDate)
	assert.Equal(t, "transfer", req.TxnType)
}

func TestTransactionHistoryResponse_Structure(t *testing.T) {
	resp := TransactionHistoryResponse{
		Transactions: []TransactionResponse{
			{ID: uuid.New(), Amount: 100},
			{ID: uuid.New(), Amount: 200},
		},
		Total:  2,
		Limit:  20,
		Offset: 0,
	}

	assert.Len(t, resp.Transactions, 2)
	assert.Equal(t, 2, resp.Total)
	assert.Equal(t, 20, resp.Limit)
}

func TestQRResolutionResponse_Structure(t *testing.T) {
	accountID := uuid.New()
	resp := QRResolutionResponse{
		AccountID: accountID,
		OwnerName: "John Doe",
		Currency:  "USD",
	}

	assert.Equal(t, accountID, resp.AccountID)
	assert.Equal(t, "John Doe", resp.OwnerName)
	assert.Equal(t, "USD", resp.Currency)
}

func TestQRResolutionRequest_Structure(t *testing.T) {
	req := QRResolutionRequest{
		QRCode: "madabank:account:12345678-1234-1234-1234-123456789012",
	}

	assert.Contains(t, req.QRCode, "madabank:account:")
}
