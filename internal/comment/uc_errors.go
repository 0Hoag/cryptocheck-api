package comment

import "errors"

var wantErrors = []error{
	ErrRequiredField,
	ErrPermissionDenied,
}

var (
	// follow
	ErrRequiredField    = errors.New("required field")
	ErrCommentNotFound  = errors.New("Comment not found")
	ErrPermissionDenied = errors.New("permission denied")
)
