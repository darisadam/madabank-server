package handlers

import (
	"net/http"

	"github.com/darisadam/madabank-server/internal/domain/transaction"
	"github.com/darisadam/madabank-server/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TransactionHandler struct {
	transactionService service.TransactionService
}

func NewTransactionHandler(transactionService service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
	}
}

// Transfer godoc
// @Summary Transfer money between accounts
// @Description Transfer money from one account to another (must own source account)
// @Tags transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body transaction.TransferRequest true "Transfer details"
// @Success 201 {object} transaction.Transaction
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/transactions/transfer [post]
func (h *TransactionHandler) Transfer(c *gin.Context) {
	val, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := val.(uuid.UUID)

	var req transaction.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	txn, err := h.transactionService.Transfer(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, txn)
}

// Deposit godoc
// @Summary Deposit money to account
// @Description Deposit money to an account
// @Tags transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body transaction.DepositRequest true "Deposit details"
// @Success 201 {object} transaction.Transaction
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/transactions/deposit [post]
func (h *TransactionHandler) Deposit(c *gin.Context) {
	val, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := val.(uuid.UUID)

	var req transaction.DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	txn, err := h.transactionService.Deposit(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, txn)
}

// Withdraw godoc
// @Summary Withdraw money from account
// @Description Withdraw money from an account
// @Tags transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body transaction.WithdrawalRequest true "Withdrawal details"
// @Success 201 {object} transaction.Transaction
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/transactions/withdraw [post]
func (h *TransactionHandler) Withdraw(c *gin.Context) {
	val, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := val.(uuid.UUID)

	var req transaction.WithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	txn, err := h.transactionService.Withdrawal(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, txn)
}

// GetHistory godoc
// @Summary Get transaction history
// @Description Get transaction history for an account with optional filters
// @Tags transactions
// @Produce json
// @Security BearerAuth
// @Param account_id query string true "Account ID"
// @Param limit query int false "Limit" default(20)
// @Param offset query int false "Offset" default(0)
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param type query string false "Transaction type"
// @Success 200 {object} transaction.TransactionHistoryResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/transactions/history [get]
func (h *TransactionHandler) GetHistory(c *gin.Context) {
	val, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := val.(uuid.UUID)

	var req transaction.TransactionHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	history, err := h.transactionService.GetTransactionHistory(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

// GetTransaction godoc
// @Summary Get transaction details
// @Description Get details of a specific transaction
// @Tags transactions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Transaction ID"
// @Success 200 {object} transaction.Transaction
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/transactions/{id} [get]
func (h *TransactionHandler) GetTransaction(c *gin.Context) {
	val, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := val.(uuid.UUID)

	transactionIDStr := c.Param("id")
	transactionID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction ID"})
		return
	}

	txn, err := h.transactionService.GetTransaction(userID, transactionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, txn)
}
