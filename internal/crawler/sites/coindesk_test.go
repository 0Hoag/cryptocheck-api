package sites

import (
	"testing"
	"time"
)

func TestParsePublishedAt(t *testing.T) {
	fallback := time.Date(2026, 7, 29, 0, 0, 0, 0, time.UTC)
	parsed := parsePublishedAt("2026-07-29T04:15:00Z", fallback)
	if !parsed.Equal(time.Date(2026, 7, 29, 4, 15, 0, 0, time.UTC)) {
		t.Fatalf("published time = %s", parsed)
	}
	if got := parsePublishedAt("not-a-date", fallback); !got.Equal(fallback) {
		t.Fatalf("invalid date fallback = %s, want %s", got, fallback)
	}
}
