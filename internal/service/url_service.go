package service

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"math/big"
	"net/url"
	"strings"
	"time"

	qrcode "github.com/skip2/go-qrcode"
	"github.com/galanjabal3/linkas/internal/model"
	"github.com/galanjabal3/linkas/internal/repository"
)

const (
	slugLength = 6
	slugChars  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var (
	ErrInvalidURL      = errors.New("invalid URL")
	ErrURLNotFound     = errors.New("URL not found")
	ErrURLExpired      = errors.New("URL has expired")
	ErrSlugExists      = errors.New("slug already exists")
	ErrInvalidSlug     = errors.New("invalid slug format")
)

type URLService struct {
	repo *repository.URLRepository
}

func NewURLService(repo *repository.URLRepository) *URLService {
	return &URLService{repo: repo}
}

func (s *URLService) Shorten(originalURL, customSlug, expiresIn string) (*model.URL, error) {
	parsedURL, err := url.ParseRequestURI(originalURL)
	if err != nil || parsedURL.Host == "" {
		return nil, ErrInvalidURL
	}

	slug := customSlug
	if slug == "" {
		slug, err = s.generateUniqueSlug()
		if err != nil {
			return nil, err
		}
	} else {
		if !isValidSlug(slug) {
			return nil, ErrInvalidSlug
		}
		exists, err := s.repo.SlugExists(slug)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrSlugExists
		}
	}

	var expiresAt *time.Time
	if expiresIn != "" {
		duration, err := time.ParseDuration(expiresIn)
		if err != nil {
			return nil, err
		}
		t := time.Now().Add(duration)
		expiresAt = &t
	}

	urlRecord := &model.URL{
		Slug:        slug,
		OriginalURL: originalURL,
		ExpiresAt:   expiresAt,
	}

	if err := s.repo.Create(urlRecord); err != nil {
		return nil, err
	}

	return urlRecord, nil
}

func (s *URLService) Redirect(slug string) (string, error) {
	urlRecord, err := s.repo.GetBySlug(slug)
	if err != nil {
		return "", err
	}
	if urlRecord == nil {
		rawRecord, err := s.repo.GetBySlugRaw(slug)
		if err != nil {
			return "", err
		}
		if rawRecord != nil {
			return "", ErrURLExpired
		}
		return "", ErrURLNotFound
	}
	return urlRecord.OriginalURL, nil
}

func (s *URLService) GetStats(slug string) (*model.StatsResponse, error) {
	urlRecord, err := s.repo.GetBySlug(slug)
	if err != nil {
		return nil, err
	}
	if urlRecord == nil {
		rawRecord, err := s.repo.GetBySlugRaw(slug)
		if err != nil {
			return nil, err
		}
		if rawRecord != nil {
			return nil, ErrURLExpired
		}
		return nil, ErrURLNotFound
	}

	recentClicks, err := s.repo.GetRecentClicks(urlRecord.ID, 10)
	if err != nil {
		return nil, err
	}

	return &model.StatsResponse{
		Slug:         urlRecord.Slug,
		OriginalURL:  urlRecord.OriginalURL,
		ClickCount:   urlRecord.ClickCount,
		CreatedAt:    urlRecord.CreatedAt,
		RecentClicks: recentClicks,
	}, nil
}

func (s *URLService) Delete(slug string) error {
	err := s.repo.Delete(slug)
	if err == sql.ErrNoRows {
		return ErrURLNotFound
	}
	return err
}

func (s *URLService) GenerateQRCode(slug, baseURL string) ([]byte, error) {
	urlRecord, err := s.repo.GetBySlug(slug)
	if err != nil {
		return nil, err
	}
	if urlRecord == nil {
		rawRecord, err := s.repo.GetBySlugRaw(slug)
		if err != nil {
			return nil, err
		}
		if rawRecord != nil {
			return nil, ErrURLExpired
		}
		return nil, ErrURLNotFound
	}

	fullURL := baseURL + "/" + slug
	png, err := qrcode.Encode(fullURL, qrcode.Medium, 256)
	if err != nil {
		return nil, err
	}

	return png, nil
}

func (s *URLService) RecordClick(slug, ip, userAgent, referrer string) error {
	urlRecord, err := s.repo.GetBySlug(slug)
	if err != nil {
		return err
	}
	if urlRecord == nil {
		return ErrURLNotFound
	}

	click := &model.Click{
		URLID:     urlRecord.ID,
		IPAddress: ip,
		UserAgent: userAgent,
		Referrer:  referrer,
	}

	if err := s.repo.RecordClick(click); err != nil {
		return err
	}

	return s.repo.IncrementClickCount(slug)
}

func (s *URLService) generateUniqueSlug() (string, error) {
	for i := 0; i < 10; i++ {
		slug := generateSlug(slugLength)
		exists, err := s.repo.SlugExists(slug)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
	}
	return "", errors.New("failed to generate unique slug")
}

func generateSlug(length int) string {
	b := make([]byte, length)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(slugChars))))
		b[i] = slugChars[n.Int64()]
	}
	return string(b)
}

func isValidSlug(slug string) bool {
	if len(slug) < 3 || len(slug) > 255 {
		return false
	}
	// Each segment must be 1-50 chars
	segments := strings.Split(slug, "/")
	for _, seg := range segments {
		if len(seg) == 0 || len(seg) > 50 {
			return false
		}
		for _, c := range seg {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
				return false
			}
		}
	}
	return true
}
