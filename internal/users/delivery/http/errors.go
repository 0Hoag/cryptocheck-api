package http

import (
	"github.com/0Hoag/cryptocheck-api/internal/users"
	pkgErrors "github.com/0Hoag/cryptocheck-api/pkg/errors"
	"net/http"
)

var (
	errWrongPaginationQuery = pkgErrors.NewHTTPError(140001, "Wrong pagination query")
	errWrongQuery           = pkgErrors.NewHTTPError(140002, "Wrong query")
	errWrongBody            = pkgErrors.NewHTTPError(140003, "Wrong body")

	// User errors
	errUserNotFound       = pkgErrors.NewHTTPError(143005, "User not found")
	errPhoneAlreadyExists = &pkgErrors.HTTPError{Code: 143006, Message: "Phone number is already registered", StatusCode: http.StatusConflict}
)

func (h handler) mapError(err error) error {
	switch err {
	case users.ErrUserNotFound:
		return errUserNotFound
	case users.ErrPhoneAlreadyExists:
		return errPhoneAlreadyExists
	default:
		return err
	}
}
