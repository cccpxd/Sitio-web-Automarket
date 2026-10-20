package domain

import "errors"

// Domain errors are sentinels so transport adapters can map them to HTTP
// responses without depending on implementation details.
var (
	ErrNotFound = errors.New("resource not found")

	ErrNameRequired     = errors.New("name is required")
	ErrNameTooLong      = errors.New("name is too long")
	ErrEmailRequired    = errors.New("email is required")
	ErrEmailInvalid     = errors.New("email is invalid")
	ErrEmailTaken       = errors.New("email is already registered")
	ErrPasswordRequired = errors.New("password is required")
	ErrPasswordTooShort = errors.New("password is too short")
	ErrPasswordTooLong  = errors.New("password is too long")

	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidSession     = errors.New("invalid or expired session")

	ErrAdminEmailConflict = errors.New("bootstrap admin email belongs to a non-admin account")
)
