package api

type ShortenResp struct {
	Result string `json:"result"`
}

type ShortenBatchResp struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
