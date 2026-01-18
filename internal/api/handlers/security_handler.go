package handlers

import (
	"net/http"

	"github.com/darisadam/madabank-server/internal/service"
	"github.com/gin-gonic/gin"
)

type SecurityHandler struct {
	securityService service.SecurityService
}

func NewSecurityHandler(securityService service.SecurityService) *SecurityHandler {
	return &SecurityHandler{
		securityService: securityService,
	}
}

// GetPublicKey godoc
// @Summary Get RSA Public Key
// @Description Get the RSA public key for frontend E2EE encryption
// @Tags security
// @Produce text/plain
// @Success 200 {string} string "PEM encoded public key"
// @Router /api/v1/security/public-key [get]
func (h *SecurityHandler) GetPublicKey(c *gin.Context) {
	pem := h.securityService.GetPublicKeyPEM()
	if pem == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve public key"})
		return
	}

	c.String(http.StatusOK, pem)
}
