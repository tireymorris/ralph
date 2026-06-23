import { useEffect, useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { listRuns } from "../api/client";
import { CLEAN_COMPLETED_EVENT } from "../lib/clean";
import {
  formatStatus,
  isTerminalRunStatus,
  relativeTime,
  statusBadgeClass,
} from "../lib/format";
import type { Run } from "../api/types";

const POLL_MS = 3000;
const RUN_PATH_RE = /^\/runs\/([^/]+)/;

export interface RunsListProps {
  activeOnly?: boolean;
  heading?: string;
  hideWhenEmpty?: boolean;
}

export default function RunsList({
  activeOnly = false,
  heading = "Runs",
  hideWhenEmpty = false,
}: RunsListProps) {
  const location = useLocation();
  const match = RUN_PATH_RE.exec(location.pathname);
  const activeId = match?.[1];
  const [runs, setRuns] = useState<Run[] | null>(null);
  const [fetchError, setFetchError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;

    async function load() {
      try {
        const data = await listRuns();
        if (!cancelled) {
          setRuns(data);
          setFetchError(null);
        }
      } catch (e) {
        if (!cancelled) {
          setFetchError(e instanceof Error ? e.message : "failed to load runs");
        }
      }
    }

    void load();
    const timer = setInterval(() => void load(), POLL_MS);
    const onClean = () => void load();
    window.addEventListener(CLEAN_COMPLETED_EVENT, onClean);
    return () => {
      cancelled = true;
      clearInterval(timer);
      window.removeEventListener(CLEAN_COMPLETED_EVENT, onClean);
    };
  }, []);

  const visibleRuns =
    runs?.filter((run) => !activeOnly || !isTerminalRunStatus(run.status)) ??
    null;

  if (hideWhenEmpty && visibleRuns !== null && visibleRuns.length === 0) {
    return null;
  }

  if (fetchError && runs === null) {
    return <p className="form-error">{fetchError}</p>;
  }

  if (runs === null) {
    if (hideWhenEmpty) {
      return null;
    }
    return <p className="content-body content-muted runs-loading">Loading runs…</p>;
  }

  const filteredRuns = visibleRuns ?? [];

  if (filteredRuns.length === 0) {
    return (
      <div className="runs-empty-state">
        <p className="content-body content-muted runs-empty">No runs yet</p>
        <Link to="/" className="btn btn--primary btn--sm">
          Start your first run
        </Link>
      </div>
    );
  }

  return (
    <div className="runs-section">
      <h2 className="runs-heading">{heading}</h2>
      <ul className="runs-list">
        {filteredRuns.map((run) => (
          <li key={run.id} className={run.id === activeId ? "active" : ""}>
            <Link
              to={`/runs/${run.id}`}
              aria-current={run.id === activeId ? "page" : undefined}
            >
              <span className="content-body run-prompt">
                {run.prompt}
                {run.source === "local_prd" ? (
                  <span className="run-source-tag">CLI</span>
                ) : null}
              </span>
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
