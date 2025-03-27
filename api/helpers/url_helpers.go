package helpers

import (
	"api/models"
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"
)

func ValidateURL(rawURL string) (bool, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false, err
	}

	return u.Scheme != "" && u.Host != "", nil

}

func GenerateShortcode(length int, charset string) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive")
	}

	if charset == "" {
		return "", fmt.Errorf("charset cannot be empty")
	}

	shortcode := strings.Builder{}
	shortcode.Grow(length)

	charsetLength := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		randomIndex, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		shortcode.WriteByte(charset[randomIndex.Int64()])
	}

	return shortcode.String(), nil
}

func ShortcodeInDatabase(shortcode string, urlRepo *models.URLRepository) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	databaseShortcodes, err := urlRepo.GetShortenedURLs(ctx)
	if err != nil {
		return false, fmt.Errorf("error getting shortened urls from database")
	}
	for _, u := range databaseShortcodes {
		if u.Shortcode == shortcode {
			return true, nil
		}
	}
	return false, nil
}

func SaveShortenedUrl(shortcode string, originalUrl string, urlRepo *models.URLRepository) (models.URL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	url, err := urlRepo.SaveURL(originalUrl, shortcode, ctx)
	if err != nil {
		return models.URL{}, err
	}
	return url, nil
}
