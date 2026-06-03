import { useEffect, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { listRuns } from "../api/client";
import { formatStatus, relativeTime, statusBadgeClass } from "../lib/format";
import type { Run } from "../api/types";

const POLL_MS = 3000;
const RUN_PATH_RE = /^\/runs\/([^/]+)/;

export default function RunsList() {
  const location = useLocation();
  const match = RUN_PATH_RE.exec(location.pathname);
  const activeId = match?.[1];
  const [runs, setRuns] = useState<Run[] | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      try {
        const data = await listRuns();
        if (!cancelled) {
          setRuns(data);
        }
      } catch {
        if (!cancelled) {
          setRuns([]);
        }
      }
    }

    void load();
    const timer = setInterval(() => void load(), POLL_MS);
    return () => {
      cancelled = true;
      clearInterval(timer);
    };
  }, []);

  if (runs === null) {
    return <p className="runs-loading">Loading runs…</p>;
  }

  if (runs.length === 0) {
    return (
      <div className="runs-empty-state">
        <p className="runs-empty">No runs yet</p>
        <Link to="/new" className="btn btn--primary btn--sm">
          Start your first run
        </Link>
      </div>
    );
  }

  return (
    <div className="runs-section">
      <h2 className="runs-heading">Runs</h2>
      <ul className="runs-list">
        {runs.map((run) => (
          <li key={run.id} className={run.id === activeId ? "active" : ""}>
            <Link
              to={`/runs/${run.id}`}
              aria-current={run.id === activeId ? "page" : undefined}
            >
              <span className="run-prompt">{run.prompt}</span>
              <span className="run-meta">
                <span
                  className={`run-status-badge ${statusBadgeClass(run.status)}`}
                >
                  {formatStatus(run.status)}
                </span>
                <span className="run-time">{relativeTime(run.created_at)}</span>
              </span>
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}
