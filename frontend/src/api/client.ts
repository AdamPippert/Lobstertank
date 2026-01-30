import type {
  Gateway,
  CreateGatewayRequest,
  HealthCheckResult,
  FanOutRequest,
  FanOutResponse,
} from "../types/gateway";

const BASE = "/api/v1";

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
    },
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(
      (body as Record<string, string>).error ?? `HTTP ${res.status}`,
    );
  }

  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export const api = {
  gateways: {
    list: () => request<Gateway[]>("/gateways"),

    get: (id: string) => request<Gateway>(`/gateways/${id}`),

    create: (data: CreateGatewayRequest) =>
      request<Gateway>("/gateways", {
        method: "POST",
        body: JSON.stringify(data),
      }),

    update: (id: string, data: Partial<CreateGatewayRequest>) =>
      request<Gateway>(`/gateways/${id}`, {
        method: "PUT",
        body: JSON.stringify(data),
      }),

    delete: (id: string) =>
      request<void>(`/gateways/${id}`, { method: "DELETE" }),

    healthCheck: (id: string) =>
      request<HealthCheckResult>(`/gateways/${id}/health`, {
        method: "POST",
      }),
  },

  meta: {
    fanOut: (data: FanOutRequest) =>
      request<FanOutResponse>("/meta/fanout", {
        method: "POST",
        body: JSON.stringify(data),
      }),
  },
};
