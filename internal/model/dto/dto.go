package dto

import "github.com/google/uuid"

// URLDto carries URL data between the service layer and repositories.
type URLDto struct {
	// CorrelationID links a batch request item to a repository record.
	CorrelationID string `json:"correlation_id"`
	// OriginalURL is the full URL supplied by the client.
	OriginalURL string `json:"original_url"`
	// ShortURL is the generated short URL key.
	ShortURL string `json:"short_url"`
	// UserID identifies the owner of the URL.
	UserID uuid.UUID `json:"user_id"`
}
