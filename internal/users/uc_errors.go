package users

import "errors"

var wantErrors = []error{
	ErrUserNotFound,
	ErrPhoneAlreadyExists,
	ErrRequiredField,
}

var (
	// user
	ErrUserNotFound       = errors.New("user not found")
	ErrPhoneAlreadyExists = errors.New("phone number is already registered")
	ErrRequiredField      = errors.New("required field")
)
