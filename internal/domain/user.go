package domain

import "time"

type Role string

const (
	RoleVisitor Role = "visitor"
	RoleSeller  Role = "seller"
	RoleAdmin   Role = "admin"
)

type User struct {
	ID        string
	Name      string
	Email     string
	Password  string
	Salt      string
	Role      Role
	CreatedAt time.Time
}

func (r Role) IsValid() bool {
	switch r {
	case RoleVisitor, RoleSeller, RoleAdmin:
		return true
	default:
		return false
	}
}
