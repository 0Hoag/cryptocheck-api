package http

import (
	"testing"

	"github.com/0Hoag/cryptocheck-api/internal/scanner"
)

func TestMapErrorMapsSolanaAndUnknownErrorsWithoutPanicking(t *testing.T) {
	h := handler{}
	if got := h.mapError(scanner.ErrSolanaMintUnavailable); got != errSolanaMintUnavailable {
		t.Fatalf("solana mint error mapping = %v, want %v", got, errSolanaMintUnavailable)
	}
	if got := h.mapError(assertionError{}); got != errScannerUnavailable {
		t.Fatalf("unknown scanner error mapping = %v, want %v", got, errScannerUnavailable)
	}
}

type assertionError struct{}

func (assertionError) Error() string { return "unexpected scanner failure" }
