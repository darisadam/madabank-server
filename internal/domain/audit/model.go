package audit

import (
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID           int64                  `json:"id"`
	EventID      uuid.UUID              `json:"event_id"`
	Timestamp    time.Time              `json:"timestamp"`
	UserID       *uuid.UUID             `json:"user_id,omitempty"`
	Action       string                 `json:"action"`
	Resource     string                 `json:"resource,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	Status       string                 `json:"status"`
	RequestBody  map[string]interface{} `json:"request_body,omitempty"`
	ResponseBody map[string]interface{} `json:"response_body,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

type CreateAuditLogRequest struct {
	UserID       *uuid.UUID
	Action       string
	Resource     string
	IPAddress    string
	UserAgent    string
	Status       string
	RequestBody  map[string]interface{}
	ResponseBody map[string]interface{}
	Metadata     map[string]interface{}
}
