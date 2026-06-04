import { useEffect, useState } from "react";
import { getRunPRD } from "../api/client";
import type { PRDDocument } from "../api/types";

const POLL_MS = 3000;

const PRD_REVIEW_STATUS = "waiting_review";
const PRD_POLL_STATUSES = new Set([
  PRD_REVIEW_STATUS,
  "implementing",
  "running",
  "completed",
  "failed",
]);

export function usePRDLoader(
  id: string | undefined,
  runStatus: string | undefined,
) {
  const [prd, setPrd] = useState<PRDDocument | null>(null);
  const [prdError, setPrdError] = useState<string | null>(null);
  const shouldLoad = !!id && !!runStatus && PRD_POLL_STATUSES.has(runStatus);
  const reviewOnly = runStatus === PRD_REVIEW_STATUS;

  useEffect(() => {
    if (!shouldLoad) {
      setPrd(null);
      setPrdError(null);
      return;
    }
    let cancelled = false;

    async function load() {
      try {
        const doc = await getRunPRD(id!);
        if (!cancelled) {
          setPrd(doc);
          setPrdError(null);
        }
      } catch (e) {
        if (!cancelled) {
          setPrd(null);
          if (reviewOnly) {
            setPrdError(
              e instanceof Error ? e.message : "failed to load PRD",
            );
          } else {
            setPrdError(null);
          }
        }
      }
    }

    void load();
    const interval = setInterval(() => {
      if (!cancelled) void load();
    }, POLL_MS);
    return () => {
      cancelled = true;
      clearInterval(interval);
    };
  }, [id, runStatus, shouldLoad, reviewOnly]);

  return { prd, prdError };
}
