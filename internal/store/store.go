package store

import (
	"sync"

	"price-aggregation-service/internal/model"
)

type Store struct {
	mu    sync.RWMutex
	price model.Price
}

func New() *Store {
	return &Store{}
}

func (s *Store) Get() model.Price {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.price
}

func (s *Store) Update(p model.Price) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.price = p
}