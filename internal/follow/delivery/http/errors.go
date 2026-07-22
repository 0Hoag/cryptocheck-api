package http

import (
	"github.com/0Hoag/cryptocheck-api/internal/follow"
	pkgErrors "github.com/0Hoag/cryptocheck-api/pkg/errors"
)

var (
	errWrongPaginationQuery = pkgErrors.NewHTTPError(140001, "Wrong pagination query")
	errWrongQuery           = pkgErrors.NewHTTPError(140002, "Wrong query")
	errWrongBody            = pkgErrors.NewHTTPError(140003, "Wrong body")

	// Follow errors
	errFollowNotFound   = pkgErrors.NewHTTPError(143004, "Follow not found")
	errAlreadyFollowing = pkgErrors.NewHTTPError(143008, "Already following this user")
	errCannotFollowSelf = pkgErrors.NewHTTPError(143009, "Cannot follow yourself")
	errPermissionDenied = pkgErrors.NewForbiddenHTTPError()
)

func (h handler) mapError(err error) error {
	switch err {
	case follow.ErrFollowNotFound:
		return errFollowNotFound
	case follow.ErrAlreadyFollowing:
		return errAlreadyFollowing
	case follow.ErrCannotFollowSelf:
		return errCannotFollowSelf
	case follow.ErrPermissionDenied:
		return errPermissionDenied
	default:
		return err
	}
}
