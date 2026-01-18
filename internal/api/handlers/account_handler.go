package handlers

import (
	"net/http"

	"github.com/darisadam/madabank-server/internal/domain/account"
	"github.com/darisadam/madabank-server/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AccountHandler struct {
	accountService service.AccountService
}

func NewAccountHandler(accountService service.AccountService) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
	}
}

// CreateAccount godoc
// @Summary Create new account
// @Description Create a new checking or savings account
// @Tags accounts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body account.CreateAccountRequest true "Account details"
// @Success 201 {object} account.Account
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/accounts [post]
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	// userID is a uuid.UUID
	val, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := val.(uuid.UUID)

	var req account.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newAccount, err := h.accountService.CreateAccount(userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newAccount)
}

// GetAccounts godoc
// @Summary Get user accounts
// @Description Get all accounts belonging to the authenticated user
// @Tags accounts
// @Produce json
// @Security BearerAuth
// @Success 200 {object} account.AccountListResponse
// @Failure 401 {object} map[string]string
// @Router /api/v1/accounts [get]
func (h *AccountHandler) GetAccounts(c *gin.Context) {
	val, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := val.(uuid.UUID)

	accounts, err := h.accountService.GetUserAccounts(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response format
	accountResponses := make([]account.AccountResponse, len(accounts))
	for i, acc := range accounts {
		accountResponses[i] = account.AccountResponse{
			ID:            acc.ID,
			AccountNumber: acc.AccountNumber,
			AccountType:   acc.AccountType,
			Balance:       acc.Balance,
			Currency:      acc.Currency,
			InterestRate:  acc.InterestRate,
			Status:        acc.Status,
			CreatedAt:     acc.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, account.AccountListResponse{
		Accounts: accountResponses,
		Total:    len(accountResponses),
	})
}

// GetAccount godoc
// @Summary Get account details
// @Description Get details of a specific account
// @Tags accounts
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Success 200 {object} account.Account
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/accounts/{id} [get]
func (h *AccountHandler) GetAccount(c *gin.Context) {
	val, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := val.(uuid.UUID)

	accountIDStr := c.Param("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	acc, err := h.accountService.GetAccount(accountID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, acc)
}

// GetBalance godoc
// @Summary Get account balance
// @Description Get current balance of a specific account
// @Tags accounts
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Success 200 {object} account.BalanceResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/accounts/{id}/balance [get]
func (h *AccountHandler) GetBalance(c *gin.Context) {
	val, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := val.(uuid.UUID)

	accountIDStr := c.Param("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	balance, err := h.accountService.GetBalance(accountID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, balance)
}

// UpdateAccount godoc
// @Summary Update account
// @Description Update account status (freeze, activate, close)
// @Tags accounts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Param request body account.UpdateAccountRequest true "Update details"
// @Success 200 {object} account.Account
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/accounts/{id} [patch]
func (h *AccountHandler) UpdateAccount(c *gin.Context) {
	val, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := val.(uuid.UUID)

	accountIDStr := c.Param("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	var req account.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.accountService.UpdateAccount(accountID, userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// CloseAccount godoc
// @Summary Close account
// @Description Close an account (balance must be zero)
// @Tags accounts
// @Produce json
// @Security BearerAuth
// @Param id path string true "Account ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/accounts/{id} [delete]
func (h *AccountHandler) CloseAccount(c *gin.Context) {
	val, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := val.(uuid.UUID)

	accountIDStr := c.Param("id")
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account ID"})
		return
	}

	if err := h.accountService.CloseAccount(accountID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
