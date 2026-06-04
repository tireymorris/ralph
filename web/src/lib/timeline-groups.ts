import type { TimelineEntry } from "./timeline";

export type TimelineGroupState = "active" | "completed" | "failed";

export interface TimelineGroup {
  id: string;
  label: string;
  entries: TimelineEntry[];
  state: TimelineGroupState;
}

const storyStartRe = /^Started story:\s*(.+)$/i;
const storyDoneRe = /^Story (completed|failed):\s*(.+)$/i;

function makeSetupGroup(): TimelineGroup {
  return { id: "setup", label: "Setup", entries: [], state: "active" };
}

export function groupTimelineEntries(entries: TimelineEntry[]): TimelineGroup[] {
  const groups: TimelineGroup[] = [];
  let current = makeSetupGroup();

  function flush() {
    if (current.entries.length === 0) return;
    groups.push(current);
  }

  for (const entry of entries) {
    if (entry.variant === "system") {
      const start = storyStartRe.exec(entry.text);
      if (start) {
        flush();
        current = {
          id: entry.id,
          label: start[1].trim(),
          entries: [entry],
          state: "active",
        };
        continue;
      }
      const done = storyDoneRe.exec(entry.text);
      if (done && current.id !== "setup") {
        current.entries.push(entry);
        current.state = done[1].toLowerCase() === "completed" ? "completed" : "failed";
        flush();
        current = makeSetupGroup();
        continue;
      }
    }
    current.entries.push(entry);
  }

  flush();
  return groups;
}
