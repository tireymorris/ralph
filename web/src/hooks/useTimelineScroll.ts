import { useEffect, type RefObject } from "react";
import type { TimelineEntry } from "../lib/timeline";

export function useTimelineScroll(
  ref: RefObject<HTMLElement | null>,
  entries: TimelineEntry[],
) {
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const distanceFromBottom =
      el.scrollHeight - el.scrollTop - el.clientHeight;
    if (distanceFromBottom <= 150) {
      if (typeof el.scrollTo === "function") {
        el.scrollTo({ top: el.scrollHeight, behavior: "instant" });
      } else {
        el.scrollTop = el.scrollHeight;
      }
    }
  }, [ref, entries]);
}
