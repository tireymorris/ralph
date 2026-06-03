import { useEffect, useState } from "react";
import { getRunPRD } from "../api/client";
import type { PRDDocument } from "../api/types";

export function usePRDLoader(
  id: string | undefined,
  runStatus: string | undefined,
) {
  const [prd, setPrd] = useState<PRDDocument | null>(null);
  const [prdError, setPrdError] = useState<string | null>(null);

  useEffect(() => {
    if (!id || runStatus !== "waiting_review") {
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
          setPrdError(e instanceof Error ? e.message : "failed to load PRD");
        }
      }
    }

    void load();
    return () => {
      cancelled = true;
    };
  }, [id, runStatus]);

  return { prd, prdError };
}
