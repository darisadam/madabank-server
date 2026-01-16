package repository

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/darisadam/madabank-server/internal/domain/card"
	"github.com/google/uuid"
)

type CardRepository interface {
	Create(card *card.Card) error
	GetByID(id uuid.UUID) (*card.Card, error)
	GetByAccountID(accountID uuid.UUID) ([]*card.Card, error)
	Update(id uuid.UUID, updates map[string]interface{}) error
	Delete(id uuid.UUID) error
	GenerateCardNumber() (string, error)
	GenerateCVV() string
}

type cardRepository struct {
	db *sql.DB
}

func NewCardRepository(db *sql.DB) CardRepository {
	return &cardRepository{db: db}
}

func (r *cardRepository) Create(c *card.Card) error {
	query := `
		INSERT INTO cards (id, account_id, card_number_encrypted, cvv_encrypted, 
		                   card_holder_name, card_type, expiry_month, expiry_year, 
		                   status, daily_limit)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at
	`

	err := r.db.QueryRow(
		query,
		c.ID,
		c.AccountID,
		c.CardNumberEncrypted,
		c.CVVEncrypted,
		c.CardHolderName,
		c.CardType,
		c.ExpiryMonth,
		c.ExpiryYear,
		c.Status,
		c.DailyLimit,
	).Scan(&c.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create card: %w", err)
	}

	return nil
}

func (r *cardRepository) GetByID(id uuid.UUID) (*card.Card, error) {
	query := `
		SELECT id, account_id, card_number_encrypted, cvv_encrypted, card_holder_name,
		       card_type, expiry_month, expiry_year, status, daily_limit, created_at
		FROM cards
		WHERE id = $1 AND status != 'expired'
	`

	c := &card.Card{}
	err := r.db.QueryRow(query, id).Scan(
		&c.ID,
		&c.AccountID,
		&c.CardNumberEncrypted,
		&c.CVVEncrypted,
		&c.CardHolderName,
		&c.CardType,
		&c.ExpiryMonth,
		&c.ExpiryYear,
		&c.Status,
		&c.DailyLimit,
		&c.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("card not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get card: %w", err)
	}

	return c, nil
}

func (r *cardRepository) GetByAccountID(accountID uuid.UUID) ([]*card.Card, error) {
	query := `
		SELECT id, account_id, card_number_encrypted, cvv_encrypted, card_holder_name,
		       card_type, expiry_month, expiry_year, status, daily_limit, created_at
		FROM cards
		WHERE account_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to list cards: %w", err)
	}
	defer rows.Close()

	cards := []*card.Card{}
	for rows.Next() {
		c := &card.Card{}
		err := rows.Scan(
			&c.ID,
			&c.AccountID,
			&c.CardNumberEncrypted,
			&c.CVVEncrypted,
			&c.CardHolderName,
			&c.CardType,
			&c.ExpiryMonth,
			&c.ExpiryYear,
			&c.Status,
			&c.DailyLimit,
			&c.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card: %w", err)
		}
		cards = append(cards, c)
	}

	return cards, nil
}

func (r *cardRepository) Update(id uuid.UUID, updates map[string]interface{}) error {
	query := "UPDATE cards SET "
	args := []interface{}{}
	argPos := 1

	for key, value := range updates {
		if argPos > 1 {
			query += ", "
		}
		query += fmt.Sprintf("%s = $%d", key, argPos)
		args = append(args, value)
		argPos++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argPos)
	args = append(args, id)

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update card: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("card not found")
	}

	return nil
}

func (r *cardRepository) Delete(id uuid.UUID) error {
	// Soft delete by setting status to expired
	query := `UPDATE cards SET status = 'expired' WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete card: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("card not found")
	}

	return nil
}

func (r *cardRepository) GenerateCardNumber() (string, error) {
	// Generate a valid 16-digit card number using Luhn algorithm
	// Format: 4XXX XXXX XXXX XXXX (starts with 4 for Visa simulation)

	rand.Seed(time.Now().UnixNano())

	const maxAttempts = 10
	for i := 0; i < maxAttempts; i++ {
		// Generate 15 random digits
		cardNumber := "4"
		for j := 0; j < 14; j++ {
			cardNumber += fmt.Sprintf("%d", rand.Intn(10))
		}

		// Calculate Luhn check digit
		checkDigit := r.calculateLuhnCheckDigit(cardNumber)
		cardNumber += fmt.Sprintf("%d", checkDigit)

		// Check if card number already exists
		var exists bool
		query := `SELECT EXISTS(SELECT 1 FROM cards WHERE card_number_encrypted = $1)`
		// Note: In production, you'd encrypt the card number before checking
		err := r.db.QueryRow(query, cardNumber).Scan(&exists)
		if err != nil {
			return "", fmt.Errorf("failed to check card number: %w", err)
		}

		if !exists {
			return cardNumber, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique card number after %d attempts", maxAttempts)
}

func (r *cardRepository) calculateLuhnCheckDigit(cardNumber string) int {
	var sum int

	for i := 0; i < len(cardNumber); i++ {
		digit := int(cardNumber[i] - '0')

		// Double every second digit from right to left
		if (len(cardNumber)-i)%2 == 0 {
			digit *= 2
			if digit > 9 {
				digit = digit%10 + digit/10
			}
		}

		sum += digit
	}

	return (10 - (sum % 10)) % 10
}

func (r *cardRepository) GenerateCVV() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%03d", rand.Intn(1000))
}
