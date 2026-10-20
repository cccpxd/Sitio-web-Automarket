package domain_test

import (
	"errors"
	"testing"

	"automarket/internal/domain"
)

func TestNormalizeRegistration(t *testing.T) {
	registration, err := domain.NormalizeRegistration("  Juan   Pérez  ", "  JUAN@Example.COM ", " secret 123 ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if registration.Name != "Juan Pérez" {
		t.Fatalf("name = %q", registration.Name)
	}
	if registration.Email != "juan@example.com" {
		t.Fatalf("email = %q", registration.Email)
	}
	if registration.Password != " secret 123 " {
		t.Fatalf("password whitespace must be preserved, got %q", registration.Password)
	}
}

func TestNormalizeRegistrationRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		userName string
		email    string
		password string
		want     error
	}{
		{name: "empty name", userName: "  ", email: "a@b.com", password: "secret1", want: domain.ErrNameRequired},
		{name: "invalid email", userName: "Ana", email: "ana.example.com", password: "secret1", want: domain.ErrEmailInvalid},
		{name: "email without tld", userName: "Ana", email: "ana@example", password: "secret1", want: domain.ErrEmailInvalid},
		{name: "short password", userName: "Ana", email: "ana@example.com", password: "123", want: domain.ErrPasswordTooShort},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := domain.NormalizeRegistration(test.userName, test.email, test.password)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestRoleIsAllowed(t *testing.T) {
	if !domain.RoleAdmin.IsAllowed(domain.RoleSeller, domain.RoleAdmin) {
		t.Fatal("admin should be allowed")
	}
	if domain.RoleVisitor.IsAllowed(domain.RoleSeller, domain.RoleAdmin) {
		t.Fatal("visitor should not be allowed")
	}
}
