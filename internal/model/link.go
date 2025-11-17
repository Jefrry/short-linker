package model

type LinkPayload struct {
	URL string `json:"url"`
}

type LinkResponse struct {
	ShortURL string `json:"result"`
}