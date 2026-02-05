package model

import (
	"errors"
	"time"
)

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
	UserID      int64
	Deleted     bool
}

type Visit struct {
	LinkID    string
	IP        string
	UA        string
	Referer   string
	VisitedAt time.Time
}

var ErrOriginalURLExists = errors.New("original url already exists")
