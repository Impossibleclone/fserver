package auth

import (
	"errors"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserInfo struct {
	LastActive time.Time
	IsActive   bool
}

type Authenticator interface {
	Authenticate(username, password string) error
	AddUser(username, password string) error
	RemoveUser(username string)
	GetUsers() []string
	GetUserInfo(username string) UserInfo
}

type MemoryAuth struct {
	mu          sync.RWMutex
	userHashes  map[string]string // Stores hashed passwords, NEVER plaintext
	activeUsers map[string]time.Time
}

func NewMemoryAuth() *MemoryAuth {
	return &MemoryAuth{
		userHashes:  make(map[string]string),
		activeUsers: make(map[string]time.Time),
	}
}

func (m *MemoryAuth) AddUser(username, password string) error {
	// Hash the password so we never store plain text!
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.userHashes[username] = string(hash)
	return nil
}

func (m *MemoryAuth) RemoveUser(username string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.userHashes, username)
	delete(m.activeUsers, username)
}

func (m *MemoryAuth) GetUsers() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var userList []string
	for u := range m.userHashes {
		userList = append(userList, u)
	}
	return userList
}

func (m *MemoryAuth) GetUserInfo(username string) UserInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	lastActive, exists := m.activeUsers[username]
	
	// Consider user active if they made a request in the last 15 minutes
	isActive := exists && time.Since(lastActive) < 15*time.Minute
	
	return UserInfo{
		LastActive: lastActive,
		IsActive:   isActive,
	}
}

func (m *MemoryAuth) Authenticate(username, password string) error {
	m.mu.RLock()
	expectedHash, exists := m.userHashes[username]
	m.mu.RUnlock()
	
	if !exists {
		return errors.New("user not found")
	}

	// Compare the hashes securely
	err := bcrypt.CompareHashAndPassword([]byte(expectedHash), []byte(password))
	if err != nil {
		return errors.New("invalid password")
	}

	// Update last active time on successful auth
	m.mu.Lock()
	m.activeUsers[username] = time.Now()
	m.mu.Unlock()

	return nil
}
