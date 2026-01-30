import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import type { Gateway } from "../types/gateway";
import { api } from "../api/client";
import { StatusBadge } from "./StatusBadge";

export function GatewayList() {
  const [gateways, setGateways] = useState<Gateway[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    api.gateways
      .list()
      .then(setGateways)
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <p className="loading">Loading gateways...</p>;
  if (error) return <p className="error">Error: {error}</p>;

  return (
    <section>
      <div className="section-header">
        <h2>Gateways</h2>
        <span className="badge">{gateways.length} registered</span>
      </div>

      {gateways.length === 0 ? (
        <div className="empty-state">
          <p>No gateways registered yet.</p>
          <p className="hint">
            Register an OpenClaw gateway via the API to get started.
          </p>
        </div>
      ) : (
        <div className="gateway-grid">
          {gateways.map((gw) => (
            <Link
              to={`/gateways/${gw.id}`}
              key={gw.id}
              className="gateway-card"
            >
              <div className="card-header">
                <h3>{gw.name}</h3>
                <StatusBadge status={gw.status} />
              </div>
              {gw.description && (
                <p className="card-desc">{gw.description}</p>
              )}
              <div className="card-meta">
                <code>{gw.endpoint}</code>
                <span className="transport-badge">{gw.transport.type}</span>
              </div>
            </Link>
          ))}
        </div>
      )}

      <style>{`
        .section-header {
          display: flex;
          align-items: center;
          gap: 1rem;
          margin-bottom: 1.5rem;
        }
        .section-header h2 {
          font-size: 1.25rem;
          font-weight: 600;
        }
        .badge {
          font-size: 0.75rem;
          background: var(--color-border);
          padding: 0.25rem 0.75rem;
          border-radius: 999px;
          color: var(--color-text-muted);
        }
        .empty-state {
          text-align: center;
          padding: 4rem 2rem;
          background: var(--color-surface);
          border-radius: var(--radius);
          border: 1px dashed var(--color-border);
        }
        .hint {
          margin-top: 0.5rem;
          color: var(--color-text-muted);
          font-size: 0.9rem;
        }
        .gateway-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
          gap: 1rem;
        }
        .gateway-card {
          display: block;
          padding: 1.25rem;
          background: var(--color-surface);
          border: 1px solid var(--color-border);
          border-radius: var(--radius);
          color: var(--color-text);
          transition: border-color 0.15s;
        }
        .gateway-card:hover {
          border-color: var(--color-primary);
        }
        .card-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 0.5rem;
        }
        .card-header h3 {
          font-size: 1rem;
          font-weight: 600;
        }
        .card-desc {
          font-size: 0.85rem;
          color: var(--color-text-muted);
          margin-bottom: 0.75rem;
        }
        .card-meta {
          display: flex;
          justify-content: space-between;
          align-items: center;
          font-size: 0.8rem;
        }
        .card-meta code {
          color: var(--color-text-muted);
        }
        .transport-badge {
          background: var(--color-border);
          padding: 0.15rem 0.5rem;
          border-radius: 4px;
          font-size: 0.75rem;
          text-transform: uppercase;
          letter-spacing: 0.05em;
        }
        .loading, .error {
          padding: 2rem;
          text-align: center;
        }
        .error {
          color: var(--color-danger);
        }
      `}</style>
    </section>
  );
}
