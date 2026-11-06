package domain

import "errors"

// Domain errors are sentinels so transport adapters can map them to HTTP
// responses without depending on implementation details.
var (
	ErrNotFound = errors.New("resource not found")
	ErrConflict = errors.New("conflict")

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

	// Vehicle publishing errors.
	ErrBrandRequired    = errors.New("brand is required")
	ErrBrandTooLong     = errors.New("brand is too long")
	ErrModelRequired    = errors.New("model is required")
	ErrModelTooLong     = errors.New("model is too long")
	ErrYearInvalid      = errors.New("year is invalid")
	ErrDescriptionLong  = errors.New("description is too long")
	ErrPriceRequired    = errors.New("price is required")
	ErrPriceInvalid     = errors.New("price is invalid")
	ErrStatusInvalid    = errors.New("status is invalid")
	ErrEditWindowClosed = errors.New("edit window has expired")
	ErrDuplicate        = errors.New("duplicate submission")

	// Photo errors.
	ErrPhotoRequired      = errors.New("at least one photo is required")
	ErrPhotosTooMany      = errors.New("too many photos")
	ErrPhotoTooLarge      = errors.New("photo is too large")
	ErrPhotoTypeInvalid   = errors.New("photo type is not allowed")
	ErrPhotoDimensions    = errors.New("photo dimensions are invalid")
	ErrPhotoDecodeFailed  = errors.New("photo could not be decoded")
	ErrPhotoWriteFailed   = errors.New("photo could not be stored")

	// Messaging errors.
	ErrMessageEmpty     = errors.New("message is empty")
	ErrMessageTooLong   = errors.New("message is too long")
	ErrSelfMessage      = errors.New("cannot message your own listing")
	ErrConversationRead = errors.New("conversation is read only")

	// Moderation errors.
	ErrReasonRequired = errors.New("reason is required")
	ErrReasonTooLong  = errors.New("reason is too long")

	// Authorization error.
	ErrForbidden = errors.New("operation is not allowed")

	// Search errors.
	ErrFilterInvalid = errors.New("filter is invalid")
)
