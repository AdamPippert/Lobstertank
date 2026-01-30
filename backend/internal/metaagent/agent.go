package metaagent

import (
	"context"

	"github.com/AdamPippert/Lobstertank/internal/audit"
	"github.com/AdamPippert/Lobstertank/internal/gateway"
	"github.com/AdamPippert/Lobstertank/internal/model"
)

// Agent orchestrates interactions across multiple OpenClaw gateways.
type Agent struct {
	registry      *gateway.Registry
	clientFactory *gateway.ClientFactory
	auditor       *audit.Logger
}

// New creates a meta-agent that can fan-out to multiple gateways.
func New(r *gateway.Registry, cf *gateway.ClientFactory, a *audit.Logger) *Agent {
	return &Agent{registry: r, clientFactory: cf, auditor: a}
}

// FanOutRequest describes a prompt to send to multiple gateways.
type FanOutRequest struct {
	GatewayIDs []string `json:"gateway_ids"` // Empty means all gateways.
	Prompt     string   `json:"prompt"`
}

// FanOutResponse aggregates responses from multiple gateways.
type FanOutResponse struct {
	Results []GatewayResult `json:"results"`
}

// GatewayResult holds the response (or error) from a single gateway.
type GatewayResult struct {
	GatewayID   string `json:"gateway_id"`
	GatewayName string `json:"gateway_name"`
	Response    string `json:"response,omitempty"`
	Error       string `json:"error,omitempty"`
}

// FanOut sends a prompt to the specified gateways concurrently and aggregates
// the results.
func (a *Agent) FanOut(ctx context.Context, req FanOutRequest) (*FanOutResponse, error) {
	gateways, err := a.resolveGateways(ctx, req.GatewayIDs)
	if err != nil {
		return nil, err
	}

	results := fanOutToGateways(ctx, a.clientFactory, gateways, req.Prompt)

	a.auditor.Log(ctx, audit.Event{
		Action: "metaagent.fanout",
		Detail: "fan-out completed",
	})

	return &FanOutResponse{Results: results}, nil
}

func (a *Agent) resolveGateways(ctx context.Context, ids []string) ([]model.Gateway, error) {
	if len(ids) == 0 {
		return a.registry.List(ctx)
	}

	gateways := make([]model.Gateway, 0, len(ids))
	for _, id := range ids {
		gw, err := a.registry.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		gateways = append(gateways, *gw)
	}
	return gateways, nil
}
