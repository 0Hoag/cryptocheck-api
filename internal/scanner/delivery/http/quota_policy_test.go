package http

import "testing"

func TestFreeDailyScanLimit(t *testing.T) {
	if freeDailyScanLimit != 2 {
		t.Fatalf("free daily scan limit = %d, want 2", freeDailyScanLimit)
	}
}
