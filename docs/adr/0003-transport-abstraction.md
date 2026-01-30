# ADR-0003: Transport Provider Abstraction

## Status

Accepted

## Context

Lobstertank users deploy OpenClaw gateways across diverse network environments:
Tailscale tailnets, Headscale servers, Cloudflare Tunnels, and plain HTTPS.
The control plane must connect to any gateway regardless of the underlying
network topology.

## Decision

Introduce a `TransportProvider` interface that returns a configured
`*http.Client` for a given transport type and parameters.

```go
type Provider interface {
    HTTPClient(transportType string, params map[string]string) *http.Client
}
```

Each gateway record stores its own `TransportConfig` specifying the type
and any transport-specific parameters.

## Rationale

- **Separation of concerns**: Gateway business logic does not need to know
  how the network connection is established.
- **Per-gateway configuration**: Different gateways can use different
  transports within the same Lobstertank instance.
- **Pluggability**: New transports (WireGuard, SSH tunnels, etc.) can be
  added by implementing the interface without modifying existing code.
- **Testability**: Tests can inject a mock transport that returns a
  test-server-backed `http.Client`.

## Consequences

- The initial implementation only provides a real HTTPS transport; Tailscale,
  Headscale, and Cloudflare are stubbed.
- Transport-specific dependencies (e.g., `tsnet`) will be added as needed.
- The abstraction assumes HTTP as the application protocol. If a future
  transport requires non-HTTP communication, the interface will need revision.
