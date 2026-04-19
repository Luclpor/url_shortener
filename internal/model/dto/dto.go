package dto

import "github.com/google/uuid"

type URLDto struct {
	CorrelationID string    `json:"correlation_id"`
	OriginalURL   string    `json:"original_url"`
	ShortURL      string    `json:"short_url"`
	UserID        uuid.UUID `json:"user_id"`
}
