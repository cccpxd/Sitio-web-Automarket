package service

import (
	"errors"
	"strings"
	"time"

	"automarket/internal/domain"
	"automarket/internal/repository"
	"automarket/internal/security"
)

type AuthService struct {
	repository repository.UserRepository
}

func NewAuthService(repo repository.UserRepository) *AuthService {
	return &AuthService{repository: repo}
}

func (s *AuthService) RegisterSeller(name, email, password string) (domain.User, error) {
	name = strings.TrimSpace(name)
	email = strings.TrimSpace(strings.ToLower(email))

	if name == "" {
		return domain.User{}, errors.New("el nombre es obligatorio")
	}
	if email == "" {
		return domain.User{}, errors.New("el email es obligatorio")
	}
	if len(password) < 6 {
		return domain.User{}, errors.New("la contraseña debe tener al menos 6 caracteres")
	}
	if !isValidEmail(email) {
		return domain.User{}, errors.New("el email no es válido")
	}
	if _, exists := s.repository.GetByEmail(email); exists {
		return domain.User{}, errors.New("ya existe un usuario con ese email")
	}

	hashValue, salt, err := security.HashPassword(password)
	if err != nil {
		return domain.User{}, err
	}

	user := domain.User{
		ID:        security.GenerateID(),
		Name:      name,
		Email:     email,
		Password:  hashValue,
		Salt:      salt,
		Role:      domain.RoleSeller,
		CreatedAt: time.Now(),
	}

	if err = s.repository.SaveUser(user); err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (s *AuthService) Login(email, password string) (string, domain.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	password = strings.TrimSpace(password)

	if email == "" {
		return "", domain.User{}, errors.New("el email es obligatorio")
	}
	if password == "" {
		return "", domain.User{}, errors.New("la contraseña es obligatoria")
	}

	user, exists := s.repository.GetByEmail(email)
	if !exists {
		return "", domain.User{}, errors.New("credenciales inválidas")
	}

	if !security.VerifyPassword(password, user.Salt, user.Password) {
		return "", domain.User{}, errors.New("credenciales inválidas")
	}

	token, err := security.GenerateToken()
	if err != nil {
		return "", domain.User{}, err
	}

	if err = s.repository.SaveSession(token, user.ID); err != nil {
		return "", domain.User{}, err
	}

	return token, user, nil
}

func (s *AuthService) Logout(token string) error {
	if token == "" {
		return nil
	}
	s.repository.DeleteSession(token)
	return nil
}

func (s *AuthService) UserFromToken(token string) (domain.User, bool) {
	return s.repository.GetUserBySession(token)
}

func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".") && len(email) > 5
}
