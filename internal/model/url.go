package model

import "time"

type URL struct {
	ID          int64      `db:"id" json:"id"`
	Slug        string     `db:"slug" json:"slug"`
	OriginalURL string     `db:"original_url" json:"original_url"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	ExpiresAt   *time.Time `db:"expires_at" json:"expires_at,omitempty"`
	ClickCount  int        `db:"click_count" json:"click_count"`
}

type Click struct {
	ID        int64     `db:"id" json:"id"`
	URLID     int64     `db:"url_id" json:"url_id"`
	IPAddress string    `db:"ip_address" json:"ip_address"`
	UserAgent string    `db:"user_agent" json:"user_agent"`
	Referrer  string    `db:"referrer" json:"referrer"`
	ClickedAt time.Time `db:"clicked_at" json:"clicked_at"`
}

type ShortenRequest struct {
	URL        string `json:"url" binding:"required,url"`
	CustomSlug string `json:"custom_slug,omitempty"`
	ExpiresIn  string `json:"expires_in" binding:"required"`
}

type ShortenResponse struct {
	Slug        string     `json:"slug"`
	ShortURL    string     `json:"short_url"`
	OriginalURL string     `json:"original_url"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type StatsResponse struct {
	Slug        string  `json:"slug"`
	OriginalURL string  `json:"original_url"`
	ClickCount  int     `json:"click_count"`
	CreatedAt   time.Time `json:"created_at"`
	RecentClicks []Click `json:"recent_clicks,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
