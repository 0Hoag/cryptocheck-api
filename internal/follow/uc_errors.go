package follow

import "errors"

var wantErrors = []error{
	ErrRequiredField,
	ErrAlreadyFollowing,
	ErrCannotFollowSelf,
	ErrPermissionDenied,
}

var (
	// follow
	ErrRequiredField    = errors.New("required field")
	ErrFollowNotFound   = errors.New("follow not found")
	ErrAlreadyFollowing = errors.New("already following this user")
	ErrCannotFollowSelf = errors.New("cannot follow yourself")
	ErrPermissionDenied = errors.New("permission denied")
)
