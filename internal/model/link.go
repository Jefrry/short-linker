package model

import "errors"

type LinkPayload struct {
	URL string `json:"url"`
}

type LinkBatchPayload struct {
	CorrelationID string `json:"correlation_id"`
	URL           string `json:"original_url"`
}

type LinkResponse struct {
	ShortURL string `json:"result"`
}

type LinkBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type LinkItem struct {
	ID          string
	OriginalURL string
}

var ErrOriginalURLExists = errors.New("original url already exists")