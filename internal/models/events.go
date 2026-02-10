package models

// generate:reset
type Event struct {
	Timestamp   int64  `json:"ts"`
	Action      string `json:"action"`
	UserID      string `json:"user_id"`
	OriginalURL string `json:"original_url"`
}
