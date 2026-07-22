package http

import (
	"github.com/0Hoag/cryptocheck-api/internal/comment"
	pkgErrors "github.com/0Hoag/cryptocheck-api/pkg/errors"
)

var (
	errWrongPaginationQuery = pkgErrors.NewHTTPError(140001, "Wrong pagination query")
	errWrongQuery           = pkgErrors.NewHTTPError(140002, "Wrong query")
	errWrongBody            = pkgErrors.NewHTTPError(140003, "Wrong body")

	// Comment errors
	errCommentNotFound = pkgErrors.NewHTTPError(143004, "Comment not found")
)

func (h handler) mapError(err error) error {
	switch err {
	case comment.ErrCommentNotFound:
		return errCommentNotFound
	default:
		return err
	}
}
