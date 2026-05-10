package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgtype"
)

func parseID(c *fiber.Ctx) (int64, error) {
	raw := c.Params("id")
	if raw == "" {
		return 0, errors.New("id is required")
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("id must be a positive integer")
	}
	return id, nil
}

func parseDate(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, errors.New("date is required")
	}
	return time.Parse("2006-01-02", value)
}

func parseTimestamp(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, errors.New("timestamp is required")
	}
	return time.Parse(time.RFC3339, value)
}

func randomToken(bytes int) (string, error) {
	buf := make([]byte, bytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func formatDate(value pgtype.Date) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format("2006-01-02")
}

func formatTime(value pgtype.Timestamptz) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format(time.RFC3339)
}

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case float64:
		return int64(val)
	case int:
		return int64(val)
	case string:
		i, _ := strconv.ParseInt(val, 10, 64)
		return i
	default:
		return 0
	}
}
