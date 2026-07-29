package scanner

import "errors"

var wantErrors = []error{
	ErrTokenNotFound,
	ErrSourceCodeNotFound,
	ErrSolanaMintUnavailable,
}

var (
	// Scanner errors
	ErrTokenNotFound      = errors.New("token not found on DexScreener")
	ErrSourceCodeNotFound = errors.New("source code not found on supported networks")
	// ErrSolanaMintUnavailable is intentionally distinct from a DexScreener
	// lookup failure: a direct base58 address does not need a market listing.
	ErrSolanaMintUnavailable = errors.New("solana mint could not be verified")
)
