package api

type CreateShortenReq struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
	URL           string `json:"url"`
}

type CreateShortURLBatchAPIModel struct {
	URLs []CreateShortenReq `json:"urls"`
}
