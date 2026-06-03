import { useEffect, useState } from "react";
import { getRunPRD } from "../api/client";
import type { PRDDocument } from "../api/types";

export function usePRDLoader(
  id: string | undefined,
  runStatus: string | undefined,
) {
  const [prd, setPrd] = useState<PRDDocument | null>(null);

  useEffect(() => {
    if (!id || runStatus !== "waiting_review") {
      setPrd(null);
      return;
    }
    let cancelled = false;

    async function load() {
      try {
        const doc = await getRunPRD(id);
        if (!cancelled) setPrd(doc);
      } catch {
        if (!cancelled) setPrd(null);
      }
    }

    void load();
    return () => {
      cancelled = true;
    };
  }, [id, runStatus]);

  return { prd };
}
