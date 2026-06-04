import { describe, expect, it } from "vitest";
import { groupTimelineEntries } from "./timeline-groups";
import type { TimelineEntry } from "./timeline";

function entry(id: string, variant: TimelineEntry["variant"], text: string): TimelineEntry {
  return { id, variant, text };
}

describe("groupTimelineEntries", () => {
  it("groups pre-story output under Setup", () => {
    const groups = groupTimelineEntries([
      entry("1", "assistant", "Analyzing…"),
      entry("2", "system", "Starting agent…"),
    ]);
    expect(groups).toHaveLength(1);
    expect(groups[0].label).toBe("Setup");
    expect(groups[0].entries).toHaveLength(2);
  });

  it("splits entries by story start and completion", () => {
    const groups = groupTimelineEntries([
      entry("1", "assistant", "prep"),
      entry("2", "system", "Started story: Auth API"),
      entry("3", "assistant", "coding"),
      entry("4", "system", "Story completed: Auth API"),
      entry("5", "system", "Started story: Admin UI"),
      entry("6", "assistant", "more work"),
    ]);
    expect(groups.map((g) => g.label)).toEqual([
      "Setup",
      "Auth API",
      "Admin UI",
    ]);
    expect(groups[1].state).toBe("completed");
    expect(groups[2].state).toBe("active");
  });
});
