package wailslogto

import (
	"sync"

	"github.com/logto-io/go/v2/client"
)

// ensure that Store implements the LogTo client.Storage interface
var _ client.Storage = (*Store)(nil)

// Store is a simple in-memory key-value store that implements the logto client.Storage interface
type Store struct {
	sync.Mutex
	vals map[string]string
}

// NewStore creates a new Store instance
func NewStore() *Store {
	return &Store{
		vals: map[string]string{},
	}
}

// GetItem retrieves a value from the store
func (s *Store) GetItem(key string) string {
	s.Lock()
	defer s.Unlock()
	return s.vals[key]
}

// SetItem sets a value in the store
func (s *Store) SetItem(key, value string) {
	s.Lock()
	defer s.Unlock()
	s.vals[key] = value
}
