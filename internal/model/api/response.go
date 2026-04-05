package api

type ShortenResp struct {
	Result string `json:"result"`
}

type ShortenBatchResp struct {
	CorrelationID string `json:"correlationId"`
	ShortURL      string `json:"shortUrl"`
}
