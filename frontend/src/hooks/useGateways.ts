import { useCallback, useEffect, useState } from "react";
import type { Gateway } from "../types/gateway";
import { api } from "../api/client";

interface UseGatewaysResult {
  gateways: Gateway[];
  loading: boolean;
  error: string | null;
  refresh: () => void;
  selectedId: string | null;
  setSelectedId: (id: string | null) => void;
}

export function useGateways(): UseGatewaysResult {
  const [gateways, setGateways] = useState<Gateway[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedId, setSelectedId] = useState<string | null>(null);

  const refresh = useCallback(() => {
    setLoading(true);
    setError(null);
    api.gateways
      .list()
      .then((data) => {
        setGateways(data);
        setLoading(false);
      })
      .catch((err: Error) => {
        setError(err.message);
        setLoading(false);
      });
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  return { gateways, loading, error, refresh, selectedId, setSelectedId };
}
