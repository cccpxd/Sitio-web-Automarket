package domain

import "time"

// Session is a server-side session. Only a fingerprint of the raw token is
// persisted, so a database leak does not expose usable session tokens.
type Session struct {
	TokenHash string
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
}
