package domain

import (
	"errors"
	"sync"
	"time"
)

// User – учётная запись.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
}

// UserStore – потокобезопасное хранилище пользователей.
type UserStore interface {
	Create(user User) error
	GetByEmail(email string) (User, bool)
}

type inmemStore struct {
	mu    sync.RWMutex
	users map[string]User // email -> User
}

func NewInmemStore() UserStore {
	return &inmemStore{users: make(map[string]User)}
}

func (s *inmemStore) Create(user User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.users[user.Email]; exists {
		return ErrUserExists
	}
	s.users[user.Email] = user
	return nil
}

func (s *inmemStore) GetByEmail(email string) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[email]
	return u, ok
}

var ErrUserExists = errors.New("user already exists")
