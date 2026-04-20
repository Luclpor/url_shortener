package model

import (
	"time"

	"github.com/google/uuid"
)

type ShortenURL struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	CorrelationID *string
	OriginalURL   string `json:"original_url"`
	ShortURL      string `json:"short_url"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
