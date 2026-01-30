export type GatewayStatus = "online" | "offline" | "degraded" | "unknown";

export interface TransportConfig {
  type: "https" | "tailscale" | "headscale" | "cloudflare";
  params?: Record<string, string>;
}

export interface GatewayAuthConfig {
  type: "token" | "mtls" | "oidc";
  params?: Record<string, string>;
  secret_ref?: string;
}

export interface Gateway {
  id: string;
  name: string;
  description?: string;
  endpoint: string;
  transport: TransportConfig;
  auth: GatewayAuthConfig;
  status: GatewayStatus;
  labels?: Record<string, string>;
  enrolled_at: string;
  last_seen_at?: string;
  ttl_seconds?: number;
}

export interface CreateGatewayRequest {
  name: string;
  description?: string;
  endpoint: string;
  transport: TransportConfig;
  auth: GatewayAuthConfig;
  labels?: Record<string, string>;
  ttl_seconds?: number;
}

export interface HealthCheckResult {
  gateway_id: string;
  status: GatewayStatus;
  latency?: string;
  error?: string;
  checked_at: string;
}

export interface FanOutRequest {
  gateway_ids: string[];
  prompt: string;
}

export interface GatewayResult {
  gateway_id: string;
  gateway_name: string;
  response?: string;
  error?: string;
}

export interface FanOutResponse {
  results: GatewayResult[];
}
