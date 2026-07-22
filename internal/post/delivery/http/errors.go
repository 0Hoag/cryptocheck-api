package http

import (
	"github.com/0Hoag/cryptocheck-api/internal/post"
	pkgErrors "github.com/0Hoag/cryptocheck-api/pkg/errors"
)

var (
	errWrongPaginationQuery = pkgErrors.NewHTTPError(140001, "Wrong pagination query")
	errWrongQuery           = pkgErrors.NewHTTPError(140002, "Wrong query")
	errWrongBody            = pkgErrors.NewHTTPError(140003, "Wrong body")

	// Post errors
	errPostVersionNotFound = pkgErrors.NewHTTPError(143004, "Post version not found")
	errPostNotFound        = pkgErrors.NewHTTPError(143005, "Post not found")
	errPermissionDenied    = pkgErrors.NewForbiddenHTTPError()

	// Reaction errors
	errReactionNotFound = pkgErrors.NewHTTPError(143006, "Reaction not found")

	// Comment errors
	errCommentNotFound = pkgErrors.NewHTTPError(143007, "Comment not found")
)

func (h handler) mapError(err error) error {
	switch err {
	case post.ErrPostNotFound:
		return errPostNotFound
	case post.ErrPostVersionNotFound:
		return errPostVersionNotFound
	case post.ErrReactionNotFound:
		return errReactionNotFound
	case post.ErrCommentNotFound:
		return errCommentNotFound
	case post.ErrPermissionDenied:
		return errPermissionDenied
	default:
		return err
	}
}
