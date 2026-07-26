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

// StatsResp is the response returned by the internal service statistics endpoint.
type StatsResp struct {
	// URLs is the total number of shortened URL records in the service.
	URLs int `json:"urls"`
	// Users is the total number of users in the service.
	Users int `json:"users"`
}
