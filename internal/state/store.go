package state

import (
	"sync"
	"time"
)

// Store manages backoff state for recurring errors to prevent notification spam.
type Store interface {
	InBackoff(fp string) bool
	RegisterFailure(fp string)
	RegisterSuccess(fp string)
	Reset()
}

// entry tracks failure count and next retry time for a fingerprint.
type entry struct {
	Failures int
	NextTry  time.Time
}

// MemoryStore is an in-memory implementation of Store with exponential backoff.
type MemoryStore struct {
	mu          sync.RWMutex
	data        map[string]*entry
	baseBackoff time.Duration
	maxBackoff  time.Duration
}

// NewMemoryStore creates a new in-memory state store.
func NewMemoryStore(baseBackoff, maxBackoff time.Duration) *MemoryStore {
	if baseBackoff == 0 {
		baseBackoff = 30 * time.Second
	}
	if maxBackoff == 0 {
		maxBackoff = 1 * time.Hour
	}
	return &MemoryStore{
		data:        make(map[string]*entry),
		baseBackoff: baseBackoff,
		maxBackoff:  maxBackoff,
	}
}

// InBackoff returns true if the fingerprint is currently in backoff period.
func (s *MemoryStore) InBackoff(fp string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, ok := s.data[fp]
	if !ok {
		return false
	}
	return time.Now().Before(e.NextTry)
}

// RegisterFailure increments the failure count and calculates next retry time with exponential backoff.
func (s *MemoryStore) RegisterFailure(fp string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	e := s.data[fp]
	if e == nil {
		e = &entry{}
		s.data[fp] = e
	}

	e.Failures++
	backoff := time.Duration(e.Failures) * s.baseBackoff
	if backoff > s.maxBackoff {
		backoff = s.maxBackoff
	}
	e.NextTry = time.Now().Add(backoff)
}

// RegisterSuccess removes the fingerprint from backoff state (error resolved).
func (s *MemoryStore) RegisterSuccess(fp string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, fp)
}

// Reset clears all backoff state (useful for testing or forced reconciliation).
func (s *MemoryStore) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]*entry)
}
