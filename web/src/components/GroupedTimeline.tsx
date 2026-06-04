import TimelineEntryBubble from "./TimelineEntry";
import { groupTimelineEntries, type TimelineGroup } from "../lib/timeline-groups";
import type { TimelineEntry } from "../lib/timeline";

interface GroupedTimelineProps {
  entries: TimelineEntry[];
}

function groupStateLabel(state: TimelineGroup["state"]): string {
  switch (state) {
    case "completed":
      return "Done";
    case "failed":
      return "Failed";
    default:
      return "In progress";
  }
}

export default function GroupedTimeline({ entries }: GroupedTimelineProps) {
  const groups = groupTimelineEntries(entries);
  const multiGroup = groups.length > 1 || groups[0]?.id !== "setup";

  if (!multiGroup) {
    return (
      <ul className="run-timeline" aria-live="polite">
        {entries.map((entry) => (
          <TimelineEntryBubble
            key={entry.id}
            variant={entry.variant}
            text={entry.text}
          />
        ))}
      </ul>
    );
  }

  return (
    <div className="grouped-timeline" aria-live="polite">
      {groups.map((group) => (
        <details
          key={group.id}
          className={`timeline-group timeline-group--${group.state}`}
          open={group.state === "active"}
        >
          <summary className="timeline-group-summary">
            <span className="timeline-group-label">{group.label}</span>
            <span className="timeline-group-meta">
              {groupStateLabel(group.state)} · {group.entries.length}{" "}
              {group.entries.length === 1 ? "entry" : "entries"}
            </span>
          </summary>
          <ul className="run-timeline">
            {group.entries.map((entry) => (
              <TimelineEntryBubble
                key={entry.id}
                variant={entry.variant}
                text={entry.text}
              />
            ))}
          </ul>
        </details>
      ))}
    </div>
  );
}
