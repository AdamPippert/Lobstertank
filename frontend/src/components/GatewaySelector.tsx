import type { Gateway } from "../types/gateway";

interface GatewaySelectorProps {
  gateways: Gateway[];
  loading: boolean;
  selectedId: string | null;
  onSelect: (id: string | null) => void;
}

export function GatewaySelector({
  gateways,
  loading,
  selectedId,
  onSelect,
}: GatewaySelectorProps) {
  return (
    <div className="gateway-selector">
      <label htmlFor="gateway-select" className="sr-only">
        Select Gateway
      </label>
      <select
        id="gateway-select"
        value={selectedId ?? ""}
        onChange={(e) => onSelect(e.target.value || null)}
        disabled={loading}
        aria-label="Select gateway"
      >
        <option value="">All Gateways</option>
        {gateways.map((gw) => (
          <option key={gw.id} value={gw.id}>
            {gw.name} — {gw.status}
          </option>
        ))}
      </select>

      <style>{`
        .gateway-selector select {
          padding: 0.5rem 2rem 0.5rem 0.75rem;
          border: 1px solid var(--color-border);
          border-radius: var(--radius);
          background: var(--color-surface);
          color: var(--color-text);
          font-size: 0.875rem;
          appearance: none;
          background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12'%3E%3Cpath fill='%238b8fa3' d='M6 8L1 3h10z'/%3E%3C/svg%3E");
          background-repeat: no-repeat;
          background-position: right 0.5rem center;
          min-width: 200px;
        }
        .gateway-selector select:focus {
          outline: 2px solid var(--color-primary);
          outline-offset: 1px;
        }
        .sr-only {
          position: absolute;
          width: 1px;
          height: 1px;
          padding: 0;
          margin: -1px;
          overflow: hidden;
          clip: rect(0,0,0,0);
          white-space: nowrap;
          border: 0;
        }
      `}</style>
    </div>
  );
}
