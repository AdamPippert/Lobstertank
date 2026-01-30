package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/AdamPippert/Lobstertank/internal/audit"
	"github.com/AdamPippert/Lobstertank/internal/model"
	"github.com/AdamPippert/Lobstertank/internal/store"
	"github.com/google/uuid"
)

// Registry manages the lifecycle of gateway registrations.
type Registry struct {
	store   store.Store
	auditor *audit.Logger
}

// NewRegistry creates a Registry backed by the given store.
func NewRegistry(s store.Store, auditor *audit.Logger) *Registry {
	return &Registry{store: s, auditor: auditor}
}

// List returns all registered gateways.
func (r *Registry) List(ctx context.Context) ([]model.Gateway, error) {
	gateways, err := r.store.ListGateways(ctx)
	if err != nil {
		return nil, fmt.Errorf("list gateways: %w", err)
	}
	return gateways, nil
}

// Get returns a single gateway by ID.
func (r *Registry) Get(ctx context.Context, id string) (*model.Gateway, error) {
	gw, err := r.store.GetGateway(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get gateway %s: %w", id, err)
	}
	return gw, nil
}

// Create registers a new gateway and returns it.
func (r *Registry) Create(ctx context.Context, req model.CreateGatewayRequest) (*model.Gateway, error) {
	now := time.Now().UTC()
	gw := &model.Gateway{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Endpoint:    req.Endpoint,
		Transport:   req.Transport,
		Auth:        req.Auth,
		Status:      model.StatusUnknown,
		Labels:      req.Labels,
		EnrolledAt:  now,
		TTLSeconds:  req.TTLSeconds,
	}

	if err := r.store.CreateGateway(ctx, gw); err != nil {
		return nil, fmt.Errorf("create gateway: %w", err)
	}

	r.auditor.Log(ctx, audit.Event{
		Action:   "gateway.created",
		Resource: gw.ID,
		Detail:   fmt.Sprintf("registered gateway %q at %s", gw.Name, gw.Endpoint),
	})

	slog.Info("gateway registered", "id", gw.ID, "name", gw.Name)
	return gw, nil
}

// Update modifies a registered gateway.
func (r *Registry) Update(ctx context.Context, id string, req model.UpdateGatewayRequest) (*model.Gateway, error) {
	gw, err := r.store.GetGateway(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get gateway for update %s: %w", id, err)
	}

	if req.Name != nil {
		gw.Name = *req.Name
	}
	if req.Description != nil {
		gw.Description = *req.Description
	}
	if req.Endpoint != nil {
		gw.Endpoint = *req.Endpoint
	}
	if req.Transport != nil {
		gw.Transport = *req.Transport
	}
	if req.Auth != nil {
		gw.Auth = *req.Auth
	}
	if req.Labels != nil {
		gw.Labels = req.Labels
	}
	if req.TTLSeconds != nil {
		gw.TTLSeconds = req.TTLSeconds
	}

	if err := r.store.UpdateGateway(ctx, gw); err != nil {
		return nil, fmt.Errorf("update gateway %s: %w", id, err)
	}

	r.auditor.Log(ctx, audit.Event{
		Action:   "gateway.updated",
		Resource: gw.ID,
		Detail:   fmt.Sprintf("updated gateway %q", gw.Name),
	})

	return gw, nil
}

// Delete removes a gateway registration.
func (r *Registry) Delete(ctx context.Context, id string) error {
	if err := r.store.DeleteGateway(ctx, id); err != nil {
		return fmt.Errorf("delete gateway %s: %w", id, err)
	}

	r.auditor.Log(ctx, audit.Event{
		Action:   "gateway.deleted",
		Resource: id,
		Detail:   "gateway deregistered",
	})

	slog.Info("gateway deregistered", "id", id)
	return nil
}

// UpdateStatus records a new status for a gateway.
func (r *Registry) UpdateStatus(ctx context.Context, id string, status model.Status) error {
	now := time.Now().UTC()
	if err := r.store.UpdateGatewayStatus(ctx, id, string(status), &now); err != nil {
		return fmt.Errorf("update status for %s: %w", id, err)
	}
	return nil
}
