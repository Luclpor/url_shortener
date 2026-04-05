package api

type CreateShortenReq struct {
	CorrelationID string `json:"correlation_id"`
	URL           string `json:"original_url"`
}

type CreateShortURLBatchApiModel struct {
	URLs []CreateShortenReq `json:"urls"`
}
