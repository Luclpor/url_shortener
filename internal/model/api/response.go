package api

// ShortenResp is the response returned by the single URL shortening endpoint.
type ShortenResp struct {
	// Result contains the absolute shortened URL.
	Result string `json:"result"`
}

// ShortenBatchResp is one response item returned by the batch shortening endpoint.
type ShortenBatchResp struct {
	// CorrelationID mirrors the request item identifier.
	CorrelationID string `json:"correlation_id"`
	// ShortURL contains the absolute shortened URL.
	ShortURL string `json:"short_url"`
}
