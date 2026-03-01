package utils

import (
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func ToPgText(s string) pgtype.Text {
	s = strings.TrimSpace(s)
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

func FormatTimestamptzOrEmpty(tz pgtype.Timestamptz) string {
	if !tz.Valid {
		return ""
	}
	return tz.Time.UTC().Format(time.RFC3339)
}

func TimeToPgTimestamptz(t time.Time) pgtype.Timestamptz {
	if t.IsZero() {
		return pgtype.Timestamptz{
			Time:  time.Time{},
			Valid: false,
		}
	}
	return pgtype.Timestamptz{
		Time:  t,
		Valid: true,
	}
}
