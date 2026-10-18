package repository

import (
	"sync"
	"time"

	"automarket/internal/domain"
	"automarket/internal/security"
)

type UserRepository interface {
	SaveUser(user domain.User) error
	GetByEmail(email string) (domain.User, bool)
	GetByID(id string) (domain.User, bool)
	SaveSession(token, userID string) error
	GetUserBySession(token string) (domain.User, bool)
	DeleteSession(token string)
}

type InMemoryRepository struct {
	mu           sync.RWMutex
	usersByID    map[string]domain.User
	usersByEmail map[string]domain.User
	sessions     map[string]string
}

func NewInMemoryRepository() *InMemoryRepository {
	repo := &InMemoryRepository{
		usersByID:    make(map[string]domain.User),
		usersByEmail: make(map[string]domain.User),
		sessions:     make(map[string]string),
	}
	repo.seedAdmin()
	return repo
}

func (r *InMemoryRepository) SaveUser(user domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.usersByID[user.ID] = user
	r.usersByEmail[user.Email] = user
	return nil
}

func (r *InMemoryRepository) GetByEmail(email string) (domain.User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.usersByEmail[email]
	return user, ok
}

func (r *InMemoryRepository) GetByID(id string) (domain.User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.usersByID[id]
	return user, ok
}

func (r *InMemoryRepository) SaveSession(token, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[token] = userID
	return nil
}

func (r *InMemoryRepository) GetUserBySession(token string) (domain.User, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	userID, ok := r.sessions[token]
	if !ok {
		return domain.User{}, false
	}
	user, ok := r.usersByID[userID]
	return user, ok
}

func (r *InMemoryRepository) DeleteSession(token string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, token)
}

func (r *InMemoryRepository) seedAdmin() {
	hash, salt, err := security.HashPassword("admin123")
	if err != nil {
		return
	}

	adminUser := domain.User{
		ID:        security.GenerateID(),
		Name:      "Administrador",
		Email:     "admin@automarket.com",
		Password:  hash,
		Salt:      salt,
		Role:      domain.RoleAdmin,
		CreatedAt: time.Now(),
	}

	r.usersByID[adminUser.ID] = adminUser
	r.usersByEmail[adminUser.Email] = adminUser
}
