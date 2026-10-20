package domain

import (
	"net/mail"
	"strings"
	"unicode/utf8"
)

const (
	minPasswordLength = 6
	maxPasswordLength = 128
	maxNameLength     = 120
	maxEmailLength    = 254
)

// Registration contains the normalized data required to create an account.
type Registration struct {
	Name     string
	Email    string
	Password string
}

// NormalizeRegistration trims and validates user-provided registration data.
// Passwords are intentionally not trimmed: whitespace is part of the secret.
func NormalizeRegistration(name, email, password string) (Registration, error) {
	registration := Registration{
		Name:     normalizeName(name),
		Email:    NormalizeEmail(email),
		Password: password,
	}

	if err := validateName(registration.Name); err != nil {
		return Registration{}, err
	}
	if err := validateEmail(registration.Email); err != nil {
		return Registration{}, err
	}
	if err := validatePassword(registration.Password); err != nil {
		return Registration{}, err
	}

	return registration, nil
}

// NormalizeEmail produces the canonical form used for lookups and uniqueness.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizeName(name string) string {
	return strings.Join(strings.Fields(name), " ")
}

func validateName(name string) error {
	if name == "" {
		return ErrNameRequired
	}
	if utf8.RuneCountInString(name) > maxNameLength {
		return ErrNameTooLong
	}
	return nil
}

func validateEmail(email string) error {
	if email == "" {
		return ErrEmailRequired
	}
	if len(email) > maxEmailLength {
		return ErrEmailInvalid
	}

	at := strings.LastIndex(email, "@")
	if at <= 0 || at == len(email)-1 {
		return ErrEmailInvalid
	}
	domainPart := email[at+1:]
	if !strings.Contains(domainPart, ".") || strings.HasPrefix(domainPart, ".") || strings.HasSuffix(domainPart, ".") {
		return ErrEmailInvalid
	}

	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return ErrEmailInvalid
	}
	return nil
}

func validatePassword(password string) error {
	if password == "" {
		return ErrPasswordRequired
	}
	length := utf8.RuneCountInString(password)
	if length < minPasswordLength {
		return ErrPasswordTooShort
	}
	if length > maxPasswordLength {
		return ErrPasswordTooLong
	}
	return nil
}
