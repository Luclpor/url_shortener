package api

// CreateShortenReq is the request model for single and batch URL shortening.
type CreateShortenReq struct {
	// CorrelationID identifies the item in a batch shortening request.
	CorrelationID string `json:"correlation_id"`
	// OriginalURL contains the full URL in a batch shortening request.
	OriginalURL string `json:"original_url"`
	// URL contains the full URL in a single JSON shortening request.
	URL string `json:"url"`
}

// CreateShortURLBatchAPIModel wraps batch shortening request items.
type CreateShortURLBatchAPIModel struct {
	// URLs is the list of URLs to shorten.
	URLs []CreateShortenReq `json:"urls"`
}
