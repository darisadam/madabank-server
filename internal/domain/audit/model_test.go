package audit

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuditLog_Structure(t *testing.T) {
	userID := uuid.New()
	eventID := uuid.New()

	log := AuditLog{
		ID:        1,
		EventID:   eventID,
		Timestamp: time.Now(),
		UserID:    &userID,
		Action:    "LOGIN",
		Resource:  "user:session",
		IPAddress: "192.168.1.1",
		UserAgent: "Mozilla/5.0",
		Status:    "success",
		Metadata: map[string]interface{}{
			"browser": "Chrome",
		},
	}

	assert.Equal(t, int64(1), log.ID)
	assert.Equal(t, eventID, log.EventID)
	assert.NotNil(t, log.UserID)
	assert.Equal(t, "LOGIN", log.Action)
	assert.Equal(t, "user:session", log.Resource)
	assert.Equal(t, "success", log.Status)
	assert.Contains(t, log.Metadata, "browser")
}

func TestAuditLog_OptionalFields(t *testing.T) {
	log := AuditLog{
		ID:      2,
		EventID: uuid.New(),
		Action:  "SYSTEM_EVENT",
		Status:  "info",
	}

	// Optional fields should be nil/empty
	assert.Nil(t, log.UserID)
	assert.Empty(t, log.Resource)
	assert.Empty(t, log.IPAddress)
	assert.Empty(t, log.UserAgent)
	assert.Nil(t, log.RequestBody)
	assert.Nil(t, log.ResponseBody)
	assert.Nil(t, log.Metadata)
}

func TestCreateAuditLogRequest_Structure(t *testing.T) {
	userID := uuid.New()
	req := CreateAuditLogRequest{
		UserID:    &userID,
		Action:    "DEPOSIT",
		Resource:  "transaction:123",
		IPAddress: "10.0.0.1",
		UserAgent: "MadaBank-iOS/1.0",
		Status:    "success",
		RequestBody: map[string]interface{}{
			"amount": 500.00,
		},
		Metadata: map[string]interface{}{
			"source": "mobile",
		},
	}

	assert.NotNil(t, req.UserID)
	assert.Equal(t, "DEPOSIT", req.Action)
	assert.Equal(t, "transaction:123", req.Resource)
	assert.Equal(t, "success", req.Status)
	assert.Contains(t, req.RequestBody, "amount")
	assert.Contains(t, req.Metadata, "source")
}

func TestAuditLog_RequestResponseBodies(t *testing.T) {
	log := AuditLog{
		ID:      3,
		EventID: uuid.New(),
		Action:  "API_CALL",
		RequestBody: map[string]interface{}{
			"method":   "POST",
			"endpoint": "/api/v1/accounts",
		},
		ResponseBody: map[string]interface{}{
			"status_code": 201,
			"body":        "created",
		},
		Status: "success",
	}

	assert.NotNil(t, log.RequestBody)
	assert.Equal(t, "POST", log.RequestBody["method"])
	assert.NotNil(t, log.ResponseBody)
	assert.Equal(t, 201, log.ResponseBody["status_code"])
}
