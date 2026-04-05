package model

import (
	"time"

	"github.com/google/uuid"
)

type ShortenURL struct {
	ID            uuid.UUID
	CorrelationID string
	OriginalURL   string
	ShortURL      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
