package repository

import (
	"database/sql"

	"github.com/galanjabal3/linkas/internal/model"
	"github.com/jmoiron/sqlx"
)

type URLRepository struct {
	db *sqlx.DB
}

func NewURLRepository(db *sqlx.DB) *URLRepository {
	return &URLRepository{db: db}
}

func (r *URLRepository) Create(url *model.URL) error {
	query := `
		INSERT INTO urls (slug, original_url, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, url.Slug, url.OriginalURL, url.ExpiresAt).Scan(&url.ID, &url.CreatedAt)
}

func (r *URLRepository) GetBySlug(slug string) (*model.URL, error) {
	var url model.URL
	query := `SELECT * FROM urls WHERE slug = $1 AND (expires_at IS NULL OR expires_at > NOW())`
	err := r.db.Get(&url, query, slug)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &url, nil
}

func (r *URLRepository) GetBySlugRaw(slug string) (*model.URL, error) {
	var url model.URL
	query := `SELECT * FROM urls WHERE slug = $1`
	err := r.db.Get(&url, query, slug)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &url, nil
}

func (r *URLRepository) IncrementClickCount(slug string) error {
	query := `UPDATE urls SET click_count = click_count + 1 WHERE slug = $1`
	_, err := r.db.Exec(query, slug)
	return err
}

func (r *URLRepository) RecordClick(click *model.Click) error {
	query := `
		INSERT INTO clicks (url_id, ip_address, user_agent, referrer)
		VALUES ($1, $2, $3, $4)
		RETURNING id, clicked_at
	`
	return r.db.QueryRow(query, click.URLID, click.IPAddress, click.UserAgent, click.Referrer).Scan(&click.ID, &click.ClickedAt)
}

func (r *URLRepository) GetRecentClicks(urlID int64, limit int) ([]model.Click, error) {
	var clicks []model.Click
	query := `
		SELECT * FROM clicks 
		WHERE url_id = $1 
		ORDER BY clicked_at DESC 
		LIMIT $2
	`
	err := r.db.Select(&clicks, query, urlID, limit)
	return clicks, err
}

func (r *URLRepository) Delete(slug string) error {
	query := `DELETE FROM urls WHERE slug = $1`
	result, err := r.db.Exec(query, slug)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *URLRepository) SlugExists(slug string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM urls WHERE slug = $1`
	err := r.db.Get(&count, query, slug)
	return count > 0, err
}
