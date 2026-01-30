import { useEffect, useState, useCallback } from "react";
import { useParams, Link } from "react-router-dom";
import type { Gateway, HealthCheckResult } from "../types/gateway";
import { api } from "../api/client";
import { StatusBadge } from "./StatusBadge";

export function GatewayDetail() {
  const { id } = useParams<{ id: string }>();
  const [gateway, setGateway] = useState<Gateway | null>(null);
  const [health, setHealth] = useState<HealthCheckResult | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;
    api.gateways
      .get(id)
      .then(setGateway)
      .catch((err: Error) => setError(err.message))
      .finally(() => setLoading(false));
  }, [id]);

  const runHealthCheck = useCallback(() => {
    if (!id) return;
    setHealth(null);
    api.gateways
      .healthCheck(id)
      .then(setHealth)
      .catch((err: Error) => setError(err.message));
  }, [id]);

  if (loading) return <p className="loading">Loading gateway...</p>;
  if (error) return <p className="error">Error: {error}</p>;
  if (!gateway) return <p className="error">Gateway not found</p>;

  return (
    <section className="detail">
      <Link to="/" className="back-link">
        &larr; All Gateways
      </Link>

      <div className="detail-header">
        <div>
          <h2>{gateway.name}</h2>
          {gateway.description && (
            <p className="detail-desc">{gateway.description}</p>
          )}
        </div>
        <StatusBadge status={gateway.status} />
      </div>

      <div className="detail-grid">
        <div className="detail-card">
          <h3>Connection</h3>
          <dl>
            <dt>Endpoint</dt>
            <dd>
              <code>{gateway.endpoint}</code>
            </dd>
            <dt>Transport</dt>
            <dd>{gateway.transport.type}</dd>
            <dt>Auth</dt>
            <dd>{gateway.auth.type}</dd>
          </dl>
        </div>

        <div className="detail-card">
          <h3>Lifecycle</h3>
          <dl>
            <dt>Enrolled</dt>
            <dd>{new Date(gateway.enrolled_at).toLocaleString()}</dd>
            <dt>Last Seen</dt>
            <dd>
              {gateway.last_seen_at
                ? new Date(gateway.last_seen_at).toLocaleString()
                : "Never"}
            </dd>
            {gateway.ttl_seconds != null && (
              <>
                <dt>TTL</dt>
                <dd>{gateway.ttl_seconds}s</dd>
              </>
            )}
          </dl>
        </div>

        <div className="detail-card">
          <h3>Labels</h3>
          {gateway.labels && Object.keys(gateway.labels).length > 0 ? (
            <div className="labels">
              {Object.entries(gateway.labels).map(([k, v]) => (
                <span key={k} className="label-tag">
                  {k}: {v}
                </span>
              ))}
            </div>
          ) : (
            <p className="muted">No labels</p>
          )}
        </div>
      </div>

      <div className="actions">
        <button className="btn btn-primary" onClick={runHealthCheck}>
          Run Health Check
        </button>
      </div>

      {health && (
        <div className="health-result">
          <h3>Health Check Result</h3>
          <dl>
            <dt>Status</dt>
            <dd>
              <StatusBadge status={health.status} />
            </dd>
            <dt>Latency</dt>
            <dd>{health.latency ?? "—"}</dd>
            <dt>Checked At</dt>
            <dd>{health.checked_at}</dd>
            {health.error && (
              <>
                <dt>Error</dt>
                <dd className="error-text">{health.error}</dd>
              </>
            )}
          </dl>
        </div>
      )}

      <style>{`
        .detail { max-width: 900px; }
        .back-link {
          display: inline-block;
          margin-bottom: 1.5rem;
          font-size: 0.9rem;
        }
        .detail-header {
          display: flex;
          justify-content: space-between;
          align-items: flex-start;
          margin-bottom: 2rem;
        }
        .detail-header h2 {
          font-size: 1.5rem;
          font-weight: 700;
        }
        .detail-desc {
          color: var(--color-text-muted);
          margin-top: 0.25rem;
        }
        .detail-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
          gap: 1rem;
          margin-bottom: 2rem;
        }
        .detail-card {
          background: var(--color-surface);
          border: 1px solid var(--color-border);
          border-radius: var(--radius);
          padding: 1.25rem;
        }
        .detail-card h3 {
          font-size: 0.85rem;
          text-transform: uppercase;
          letter-spacing: 0.05em;
          color: var(--color-text-muted);
          margin-bottom: 0.75rem;
        }
        dl {
          display: grid;
          grid-template-columns: auto 1fr;
          gap: 0.4rem 1rem;
          font-size: 0.9rem;
        }
        dt { color: var(--color-text-muted); }
        dd { color: var(--color-text); }
        .labels {
          display: flex;
          flex-wrap: wrap;
          gap: 0.5rem;
        }
        .label-tag {
          font-size: 0.8rem;
          background: var(--color-border);
          padding: 0.2rem 0.6rem;
          border-radius: 4px;
          font-family: var(--font-mono);
        }
        .muted { color: var(--color-text-muted); font-size: 0.9rem; }
        .actions { margin-bottom: 2rem; }
        .health-result {
          background: var(--color-surface);
          border: 1px solid var(--color-border);
          border-radius: var(--radius);
          padding: 1.25rem;
        }
        .health-result h3 {
          margin-bottom: 0.75rem;
        }
        .error-text { color: var(--color-danger); }
        .loading, .error {
          padding: 2rem;
          text-align: center;
        }
        .error { color: var(--color-danger); }
      `}</style>
    </section>
  );
}
