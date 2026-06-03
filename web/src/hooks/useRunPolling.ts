import { useEffect, useState } from "react";
import { getRun } from "../api/client";
import type { Run } from "../api/types";

const POLL_MS = 3000;
const TERMINAL_STATUSES = new Set(["completed", "failed", "cancelled"]);

export function useRunPolling(id: string | undefined) {
  const [run, setRun] = useState<Run | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!id) return;
    let cancelled = false;

    async function load() {
      try {
        const data = await getRun(id!);
        if (!cancelled) {
          setRun(data);
          setError(null);
        }
      } catch (e) {
        if (!cancelled) {
          setError(e instanceof Error ? e.message : "failed to load run");
        }
      }
    }

    void load();
    const interval = setInterval(() => {
      if (cancelled) return;
      if (run && TERMINAL_STATUSES.has(run.status)) return;
      void load();
    }, POLL_MS);
    return () => {
      cancelled = true;
      clearInterval(interval);
    };
  }, [id, run?.status]);

  const setRunDirectly = setRun;
  const setErrorDirectly = setError;

  return { run, loadError: error, setRun: setRunDirectly, setLoadError: setErrorDirectly };
}
