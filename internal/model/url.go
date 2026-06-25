package model

import (
	"time"

	"github.com/google/uuid"
)

// ShortenURL describes a stored mapping between a short URL and its original URL.
type ShortenURL struct {
	// ID is the storage identifier of the URL record.
	ID uuid.UUID
	// UserID identifies the owner of the URL.
	UserID uuid.UUID
	// CorrelationID links a batch request item to its response item.
	CorrelationID *string
	// OriginalURL is the full URL supplied by the client.
	OriginalURL string `json:"original_url"`
	// ShortURL is the generated short URL key.
	ShortURL string `json:"short_url"`
	// IsDeleted marks URLs that should return HTTP 410.
	IsDeleted bool
	// CreatedAt is the creation timestamp from persistent storage.
	CreatedAt time.Time
	// UpdatedAt is the last update timestamp from persistent storage.
	UpdatedAt time.Time
}
