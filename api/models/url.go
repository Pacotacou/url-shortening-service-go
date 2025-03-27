package models

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type URL struct {
	ID          int       `json:"id"`
	OriginalURL string    `json:"original_url"`
	Shortcode   string    `json:"shortcode"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	AccessCount int       `json:"access_count"`
}

type URLRepository struct {
	db *sql.DB
}

func NewURLRepository(db *sql.DB) *URLRepository {
	return &URLRepository{db: db}
}

func (r *URLRepository) DeleteURL(shortcode string, ctx context.Context) error {
	query := `DELETE FROM shortened_urls WHERE shortcode = $1`
	result, err := r.db.ExecContext(ctx, query, shortcode)
	if err != nil {
		return fmt.Errorf("error deleting url: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("shortcode %s not found", shortcode)
	}
	return nil
}

func (r *URLRepository) UpdateURL(shortcode string, newURL string, ctx context.Context) (URL, error) {
	query := `UPDATE shortened_urls SET original_url = $1,
	updated_at = CURRENT_TIMESTAMP 
	WHERE shortcode = $2
	RETURNING url_id, original_url, shortcode, created_at, updated_at, access_count`

	var url URL
	err := r.db.QueryRowContext(ctx, query, newURL, shortcode).Scan(
		&url.ID,
		&url.OriginalURL,
		&url.Shortcode,
		&url.CreatedAt,
		&url.UpdatedAt,
		&url.AccessCount,
	)
	if err != nil {
		return url, fmt.Errorf("failed to update URL: %w", err)
	}

	return url, nil
}

func (r *URLRepository) UpdateAccessCountByShortcode(shortcode string, ctx context.Context) error {
	query := `UPDATE shortened_urls SET access_count = access_count + 1,
	 updated_at = CURRENT_TIMESTAMP
	WHERE shortcode = $1`
	result, err := r.db.ExecContext(ctx, query, shortcode)
	if err != nil {
		return fmt.Errorf("failed to increment access count: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("shortcode %s not found", shortcode)
	}

	return nil

}

func (r *URLRepository) GetByShortcode(shortcode string, ctx context.Context) (URL, error) {
	query := `SELECT url_id, original_url, shortcode, created_at, updated_at, access_count
	FROM shortened_urls WHERE shortcode = $1`
	var url URL
	err := r.db.QueryRowContext(ctx, query, shortcode).Scan(
		&url.ID,
		&url.OriginalURL,
		&url.Shortcode,
		&url.CreatedAt,
		&url.UpdatedAt,
		&url.AccessCount,
	)
	if err != nil {
		return url, fmt.Errorf("error querying by shortcode: %v", err)
	}
	return url, nil
}

func (r *URLRepository) SaveURL(originalURL string, shortcode string, ctx context.Context) (URL, error) {
	query := `INSERT INTO shortened_urls (original_url,shortcode)
	VALUES ($1,$2)
	RETURNING url_id, original_url, shortcode, created_at, updated_at, access_count`

	var url URL
	err := r.db.QueryRowContext(ctx, query, originalURL, shortcode).Scan(
		&url.ID,
		&url.OriginalURL,
		&url.Shortcode,
		&url.CreatedAt,
		&url.UpdatedAt,
		&url.AccessCount,
	)
	if err != nil {
		return url, fmt.Errorf("error inserting URL: %v", err)
	}
	return url, nil
}

func (r *URLRepository) CountURLs(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM shortened_urls`
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *URLRepository) GetShortenedURLs(ctx context.Context) ([]URL, error) {
	query := `SELECT * FROM shortened_urls`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shortenedUrls []URL
	for rows.Next() {
		var url URL
		err := rows.Scan(
			&url.ID,
			&url.OriginalURL,
			&url.Shortcode,
			&url.CreatedAt,
			&url.UpdatedAt,
			&url.AccessCount,
		)
		if err != nil {
			return nil, err
		}
		shortenedUrls = append(shortenedUrls, url)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return shortenedUrls, nil

}
