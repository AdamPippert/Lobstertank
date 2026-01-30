import { Outlet } from "react-router-dom";
import { GatewaySelector } from "./GatewaySelector";
import { useGateways } from "../hooks/useGateways";

export function Layout() {
  const { gateways, loading, selectedId, setSelectedId, refresh } =
    useGateways();

  return (
    <div className="layout">
      <header className="header">
        <div className="header-left">
          <h1 className="logo">Lobstertank</h1>
          <span className="tagline">OpenClaw Gateway Control Plane</span>
        </div>
        <div className="header-right">
          <GatewaySelector
            gateways={gateways}
            loading={loading}
            selectedId={selectedId}
            onSelect={setSelectedId}
          />
          <button className="btn btn-secondary" onClick={refresh}>
            Refresh
          </button>
        </div>
      </header>
      <main className="main">
        <Outlet />
      </main>
      <footer className="footer">
        <span>Lobstertank v0.1.0</span>
      </footer>

      <style>{`
        .layout {
          min-height: 100vh;
          display: flex;
          flex-direction: column;
        }
        .header {
          display: flex;
          align-items: center;
          justify-content: space-between;
          padding: 1rem 2rem;
          background: var(--color-surface);
          border-bottom: 1px solid var(--color-border);
        }
        .header-left {
          display: flex;
          align-items: baseline;
          gap: 1rem;
        }
        .logo {
          font-size: 1.5rem;
          font-weight: 700;
          color: var(--color-primary);
        }
        .tagline {
          font-size: 0.85rem;
          color: var(--color-text-muted);
        }
        .header-right {
          display: flex;
          align-items: center;
          gap: 0.75rem;
        }
        .main {
          flex: 1;
          padding: 2rem;
          max-width: 1200px;
          width: 100%;
          margin: 0 auto;
        }
        .footer {
          padding: 1rem 2rem;
          text-align: center;
          font-size: 0.8rem;
          color: var(--color-text-muted);
          border-top: 1px solid var(--color-border);
        }
        .btn {
          padding: 0.5rem 1rem;
          border: 1px solid var(--color-border);
          border-radius: var(--radius);
          background: var(--color-surface);
          color: var(--color-text);
          font-size: 0.875rem;
          transition: background 0.15s;
        }
        .btn:hover {
          background: var(--color-border);
        }
        .btn-primary {
          background: var(--color-primary);
          border-color: var(--color-primary);
        }
        .btn-primary:hover {
          background: var(--color-primary-hover);
        }
        .btn-secondary {
          background: transparent;
        }
      `}</style>
    </div>
  );
}
