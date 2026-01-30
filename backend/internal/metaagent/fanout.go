package metaagent

import (
	"context"
	"sync"

	"github.com/AdamPippert/Lobstertank/internal/gateway"
	"github.com/AdamPippert/Lobstertank/internal/model"
)

// fanOutToGateways sends a prompt to all gateways concurrently and collects results.
func fanOutToGateways(
	ctx context.Context,
	factory *gateway.ClientFactory,
	gateways []model.Gateway,
	prompt string,
) []GatewayResult {
	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		results = make([]GatewayResult, 0, len(gateways))
	)

	for i := range gateways {
		gw := gateways[i]
		wg.Add(1)
		go func() {
			defer wg.Done()

			client := factory.ClientFor(&gw)
			resp, err := client.SendPrompt(ctx, prompt)

			result := GatewayResult{
				GatewayID:   gw.ID,
				GatewayName: gw.Name,
			}
			if err != nil {
				result.Error = err.Error()
			} else {
				result.Response = string(resp)
			}

			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}()
	}

	wg.Wait()
	return results
}
