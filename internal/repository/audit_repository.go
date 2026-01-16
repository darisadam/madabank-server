package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/darisadam/madabank-server/internal/domain/audit"
)

type AuditRepository interface {
	Create(log *audit.AuditLog) error
}

type auditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Create(log *audit.AuditLog) error {
	requestJSON, _ := json.Marshal(log.RequestBody)
	responseJSON, _ := json.Marshal(log.ResponseBody)
	metadataJSON, _ := json.Marshal(log.Metadata)

	query := `
		INSERT INTO audit_logs (event_id, user_id, action, resource, ip_address, 
		                       user_agent, status, request_body, response_body, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, timestamp
	`

	err := r.db.QueryRow(
		query,
		log.EventID,
		log.UserID,
		log.Action,
		log.Resource,
		log.IPAddress,
		log.UserAgent,
		log.Status,
		requestJSON,
		responseJSON,
		metadataJSON,
	).Scan(&log.ID, &log.Timestamp)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}
