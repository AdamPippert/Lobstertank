package store

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/AdamPippert/Lobstertank/internal/model"
)

// SQLiteStore is an in-memory store that mimics SQLite behavior.
// TODO: Replace with actual database/sql + modernc.org/sqlite implementation.
type SQLiteStore struct {
	mu       sync.RWMutex
	gateways map[string]*model.Gateway
}

// NewSQLiteStore creates a new in-memory store.
func NewSQLiteStore(_ string) (*SQLiteStore, error) {
	return &SQLiteStore{
		gateways: make(map[string]*model.Gateway),
	}, nil
}

func (s *SQLiteStore) ListGateways(_ context.Context) ([]model.Gateway, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]model.Gateway, 0, len(s.gateways))
	for _, gw := range s.gateways {
		result = append(result, deepCopy(gw))
	}
	return result, nil
}

func (s *SQLiteStore) GetGateway(_ context.Context, id string) (*model.Gateway, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	gw, ok := s.gateways[id]
	if !ok {
		return nil, fmt.Errorf("gateway not found: %s", id)
	}
	cp := deepCopy(gw)
	return &cp, nil
}

func (s *SQLiteStore) CreateGateway(_ context.Context, gw *model.Gateway) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.gateways[gw.ID]; exists {
		return fmt.Errorf("gateway already exists: %s", gw.ID)
	}
	cp := deepCopy(gw)
	s.gateways[gw.ID] = &cp
	return nil
}

func (s *SQLiteStore) UpdateGateway(_ context.Context, gw *model.Gateway) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.gateways[gw.ID]; !exists {
		return fmt.Errorf("gateway not found: %s", gw.ID)
	}
	cp := deepCopy(gw)
	s.gateways[gw.ID] = &cp
	return nil
}

func (s *SQLiteStore) DeleteGateway(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.gateways[id]; !exists {
		return fmt.Errorf("gateway not found: %s", id)
	}
	delete(s.gateways, id)
	return nil
}

func (s *SQLiteStore) UpdateGatewayStatus(_ context.Context, id string, status string, lastSeen *time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	gw, ok := s.gateways[id]
	if !ok {
		return fmt.Errorf("gateway not found: %s", id)
	}
	gw.Status = model.Status(status)
	gw.LastSeenAt = lastSeen
	return nil
}

func (s *SQLiteStore) Close() error {
	return nil
}

// deepCopy produces a value copy via JSON round-trip to prevent aliasing.
func deepCopy(gw *model.Gateway) model.Gateway {
	data, _ := json.Marshal(gw)
	var cp model.Gateway
	_ = json.Unmarshal(data, &cp)
	return cp
}
