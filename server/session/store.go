package session

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/you/lexsy-mvp/server/models"
	"github.com/you/lexsy-mvp/server/utils"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

// Store is a thread-safe in-memory session store
type Store struct {
	mu       sync.RWMutex
	sessions map[string]*models.Session
}

// NewStore creates a new session store
func NewStore() *Store {
	return &Store{
		sessions: make(map[string]*models.Session),
	}
}

// Create creates a new session and returns its ID
func (s *Store) Create(docBytes []byte, fields []string) (*models.Session, error) {
	id, err := generateID()
	if err != nil {
		return nil, err
	}

	// Infer field types
	fieldTypes := make(map[string]string)
	for _, field := range fields {
		fieldTypes[field] = utils.InferFieldType(field)
	}

	now := time.Now()
	session := &models.Session{
		ID:          id,
		OriginalDoc: docBytes,
		Fields:      fields,
		FieldTypes:  fieldTypes,
		Answers:     make(map[string]string),
		Questions:   make(map[string]string),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	s.mu.Lock()
	s.sessions[id] = session
	s.mu.Unlock()

	return session, nil
}

// Get retrieves a session by ID
func (s *Store) Get(id string) (*models.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[id]
	if !exists {
		return nil, ErrSessionNotFound
	}

	return session, nil
}

// Update updates a session (used for adding answers, questions)
func (s *Store) Update(id string, updateFn func(*models.Session)) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[id]
	if !exists {
		return ErrSessionNotFound
	}

	updateFn(session)
	session.UpdatedAt = time.Now()

	return nil
}

// Delete removes a session from the store
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.sessions[id]; !exists {
		return ErrSessionNotFound
	}

	delete(s.sessions, id)
	return nil
}

// generateID creates a random session ID
func generateID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
