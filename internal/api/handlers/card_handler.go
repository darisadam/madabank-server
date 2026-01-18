package handlers

import (
	"net/http"

	"github.com/darisadam/madabank-server/internal/domain/card"
	"github.com/darisadam/madabank-server/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CardHandler struct {
	cardService service.CardService
}

func NewCardHandler(cardService service.CardService) *CardHandler {
	return &CardHandler{
		cardService: cardService,
	}
}

// CreateCard godoc
// @Summary Create new card
// @Description Create a new debit or credit card for an account
// @Tags cards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body card.CreateCardRequest true "Card details"
// @Success 201 {object} card.CardResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/cards [post]
func (h *CardHandler) CreateCard(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req card.CreateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newCard, err := h.cardService.CreateCard(userID.(uuid.UUID), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newCard)
}

// GetCards godoc
// @Summary Get user cards
// @Description Get all cards for a specific account
// @Tags cards
// @Produce json
// @Security BearerAuth
// @Param account_id query string true "Account ID"
// @Success 200 {array} card.CardResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/cards [get]
func (h *CardHandler) GetCards(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	accountIDStr := c.Query("account_id")
	if accountIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account_id is required"})
		return
	}

	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account_id"})
		return
	}

	cards, err := h.cardService.GetUserCards(userID.(uuid.UUID), accountID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cards": cards})
}

// GetCardDetails godoc
// @Summary Get full card details
// @Description Get decrypted card details (requires password verification)
// @Tags cards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body card.CardDetailsRequest true "Card ID and password"
// @Success 200 {object} card.CardDetailsResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/cards/details [post]
func (h *CardHandler) GetCardDetails(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req card.CardDetailsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cardID, err := uuid.Parse(req.CardID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card_id"})
		return
	}

	details, err := h.cardService.GetCardDetails(userID.(uuid.UUID), cardID, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, details)
}

// UpdateCard godoc
// @Summary Update card
// @Description Update card status or daily limit
// @Tags cards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Card ID"
// @Param request body card.UpdateCardRequest true "Update details"
// @Success 200 {object} card.CardResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/cards/{id} [patch]
func (h *CardHandler) UpdateCard(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	cardIDStr := c.Param("id")
	cardID, err := uuid.Parse(cardIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card ID"})
		return
	}

	var req card.UpdateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.cardService.UpdateCard(userID.(uuid.UUID), cardID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// BlockCard godoc
// @Summary Block card
// @Description Block a card (set status to blocked)
// @Tags cards
// @Produce json
// @Security BearerAuth
// @Param id path string true "Card ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/cards/{id}/block [post]
func (h *CardHandler) BlockCard(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	cardIDStr := c.Param("id")
	cardID, err := uuid.Parse(cardIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card ID"})
		return
	}

	if err := h.cardService.BlockCard(userID.(uuid.UUID), cardID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Card blocked successfully"})
}

// DeleteCard godoc
// @Summary Delete card
// @Description Delete a card (soft delete)
// @Tags cards
// @Produce json
// @Security BearerAuth
// @Param id path string true "Card ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/cards/{id} [delete]
func (h *CardHandler) DeleteCard(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	cardIDStr := c.Param("id")
	cardID, err := uuid.Parse(cardIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid card ID"})
		return
	}

	if err := h.cardService.DeleteCard(userID.(uuid.UUID), cardID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
