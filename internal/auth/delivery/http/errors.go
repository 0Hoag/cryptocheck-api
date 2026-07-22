package http

import (
	"net/http"

	"github.com/0Hoag/cryptocheck-api/internal/auth"
	"github.com/0Hoag/cryptocheck-api/internal/users"
	pkgErrors "github.com/0Hoag/cryptocheck-api/pkg/errors"
)

var (
	errWrongBody = pkgErrors.NewHTTPError(140003, "Wrong body")

	// User errors
	errUserNotFound = pkgErrors.NewHTTPError(143005, "User not found")
	errInvalidCreds = &pkgErrors.HTTPError{Code: 141001, Message: "Invalid phone or password", StatusCode: http.StatusUnauthorized}
)

func (h handler) mapError(err error) error {
	switch err {
	case users.ErrUserNotFound:
		return errUserNotFound
	case auth.ErrInvalidCreds:
		return errInvalidCreds
	default:
		return err
	}
}
