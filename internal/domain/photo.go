package domain

import (
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MaxPhotoBytes     = 5 << 20 // 5 MiB per uploaded file
	MaxPhotoDimension = 6000
	MinPhotoDimension = 50
	MaxPhotoNameLen   = 120
)

// CarPhoto is a stored photograph. Data is only populated by queries that
// explicitly request the binary content.
type CarPhoto struct {
	ID           string
	CarID        string
	Position     int
	ContentType  string
	ByteSize     int64
	SHA256       string
	Width        int
	Height       int
	OriginalName string
	Data         []byte
	CreatedAt    time.Time

	CarStatus   CarStatus
	CarSellerID string
}

// IsAllowedPhotoType reports whether the detected content type is accepted.
func IsAllowedPhotoType(contentType string) bool {
	switch contentType {
	case "image/jpeg", "image/png":
		return true
	default:
		return false
	}
}

// ValidatePhotoUpload enforces the size, quantity and dimension limits.
func ValidatePhotoUpload(contentType string, size int64, width, height int) error {
	if !IsAllowedPhotoType(contentType) {
		return ErrPhotoTypeInvalid
	}
	if size <= 0 || size > MaxPhotoBytes {
		return ErrPhotoTooLarge
	}
	if width < MinPhotoDimension || height < MinPhotoDimension ||
		width > MaxPhotoDimension || height > MaxPhotoDimension {
		return ErrPhotoDimensions
	}
	return nil
}

// ValidatePhotoCount enforces the number of photos attached to a listing.
func ValidatePhotoCount(count int) error {
	if count < 1 {
		return ErrPhotoRequired
	}
	if count > MaxCarPhotos {
		return ErrPhotosTooMany
	}
	return nil
}

// SanitizeOriginalName keeps a safe, short basename for display only. It is
// never used as a storage key or filesystem path.
func SanitizeOriginalName(name string) string {
	name = strings.TrimSpace(filepath.Base(strings.ReplaceAll(name, "\\", "/")))
	name = strings.Map(func(char rune) rune {
		switch char {
		case '<', '>', ':', '"', '/', '\\', '|', '?', '*':
			return '-'
		default:
			return char
		}
	}, name)
	if utf8.RuneCountInString(name) > MaxPhotoNameLen {
		runes := []rune(name)
		name = string(runes[:MaxPhotoNameLen])
	}
	return name
}
