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
    if (distanceFromBottom <= 100) {
      el.scrollTop = el.scrollHeight;
    }
  }, [ref, entries]);
}
