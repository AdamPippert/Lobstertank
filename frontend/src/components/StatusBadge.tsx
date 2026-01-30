import type { GatewayStatus } from "../types/gateway";

const STATUS_COLORS: Record<GatewayStatus, string> = {
  online: "var(--color-success)",
  offline: "var(--color-danger)",
  degraded: "var(--color-warning)",
  unknown: "var(--color-text-muted)",
};

interface StatusBadgeProps {
  status: GatewayStatus;
}

export function StatusBadge({ status }: StatusBadgeProps) {
  return (
    <span
      className="status-badge"
      style={{ "--dot-color": STATUS_COLORS[status] } as React.CSSProperties}
    >
      <span className="status-dot" />
      {status}

      <style>{`
        .status-badge {
          display: inline-flex;
          align-items: center;
          gap: 0.4rem;
          font-size: 0.8rem;
          text-transform: capitalize;
          color: var(--color-text-muted);
        }
        .status-dot {
          width: 8px;
          height: 8px;
          border-radius: 50%;
          background: var(--dot-color);
        }
      `}</style>
    </span>
  );
}
