package service

import (
	"testing"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

func TestIsValidSlug(t *testing.T) {
	tests := []struct {
		name  string
		slug  string
		valid bool
	}{
		{"simple slug", "abc123", true},
		{"with hyphen", "my-link", true},
		{"with underscore", "my_link", true},
		{"with path", "video/01", true},
		{"nested path", "blog/posts/2024", true},
		{"mixed case", "MyLink", true},
		{"too short", "ab", false},
		{"empty", "", false},
		{"with space", "my link", false},
		{"with special char", "my@link", false},
		{"path with empty segment", "video//01", false},
		{"path ends with slash", "video/", false},
		{"path starts with slash", "/video/01", false},
		{"single char segment", "a/b", true},
		{"segment too long", "video/0123456789012345678901234567890123456789012345678901", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidSlug(tt.slug)
			if got != tt.valid {
				t.Errorf("isValidSlug(%q) = %v, want %v", tt.slug, got, tt.valid)
			}
		})
	}
}

func TestGenerateSlug(t *testing.T) {
	for i := 0; i < 100; i++ {
		slug := generateSlug(slugLength)
		if len(slug) != slugLength {
			t.Errorf("generateSlug(%d) returned length %d", slugLength, len(slug))
		}
		for _, c := range slug {
			valid := false
			for _, validChar := range slugChars {
				if c == validChar {
					valid = true
					break
				}
			}
			if !valid {
				t.Errorf("generateSlug returned invalid char: %c", c)
			}
		}
	}
}

func TestGenerateSlugLength(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"short", 4},
		{"default", 6},
		{"long", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slug := generateSlug(tt.length)
			if len(slug) != tt.length {
				t.Errorf("generateSlug(%d) = len %d", tt.length, len(slug))
			}
		})
	}
}

func TestGenerateSlugUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		slug := generateSlug(slugLength)
		if seen[slug] {
			t.Errorf("duplicate slug generated: %s", slug)
		}
		seen[slug] = true
	}
}

func TestQRCodeGeneration(t *testing.T) {
	png, err := qrcode.Encode("http://localhost:8080/test123", qrcode.Medium, 256)
	if err != nil {
		t.Fatalf("failed to generate QR code: %v", err)
	}

	if len(png) == 0 {
		t.Error("QR code PNG is empty")
	}

	if png[0] != 0x89 || png[1] != 0x50 {
		t.Error("output is not a valid PNG file")
	}
}

func TestExpirationParsing(t *testing.T) {
	tests := []struct {
		name      string
		expiresIn string
		valid     bool
	}{
		{"24 hours", "24h", true},
		{"30 minutes", "30m", true},
		{"7 days", "168h", true},
		{"empty (no expiry)", "", true},
		{"invalid format", "7d", false},
		{"invalid format days", "1week", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expiresIn == "" {
				if !tt.valid {
					t.Error("empty string should be valid (no expiry)")
				}
				return
			}

			_, err := time.ParseDuration(tt.expiresIn)
			if tt.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}
