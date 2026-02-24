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
	ID          string `json:"id"`
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
	UserID      int64  `json:"user_id"`
	Deleted     bool   `json:"deleted"`
}

type Visit struct {
	LinkID    string    `json:"link_id"`
	IP        string    `json:"ip"`
	UA        string    `json:"ua"`
	Referer   string    `json:"referer"`
	VisitedAt time.Time `json:"visited_at"`
}

type VisitsByDate struct {
	Date  time.Time `json:"date"`
	Count int       `json:"count"`
}

var ErrOriginalURLExists = errors.New("original url already exists")
