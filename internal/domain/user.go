package domain

import "time"

// Role represents the authorization level of a user.
type Role string

const (
	RoleVisitor Role = "visitor"
	RoleSeller  Role = "seller"
	RoleAdmin   Role = "admin"
)

// User is the domain representation of an authenticated account.
// PasswordHash must never be exposed through HTTP responses.
type User struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string `json:"-"`
	Role         Role
	CreatedAt    time.Time
}

func (r Role) IsValid() bool {
	switch r {
	case RoleVisitor, RoleSeller, RoleAdmin:
		return true
	default:
		return false
	}
}

// IsAllowed reports whether the role is part of the provided allow-list.
func (r Role) IsAllowed(allowed ...Role) bool {
	for _, candidate := range allowed {
		if r == candidate {
			return true
		}
	}
	return false
}
