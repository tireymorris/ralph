import type {
  EventEnvelope,
  EventErrorPayload,
  EventOutputPayload,
} from "../api/types";

export type TimelineVariant = "assistant" | "system" | "error";

export interface TimelineEntry {
  id: string;
  variant: TimelineVariant;
  text: string;
}

export interface StoryPayload {
  id?: string;
  title?: string;
}

export interface StoryCompletedPayload {
  Story?: StoryPayload;
  Success?: boolean;
}

let entryCounter = 0;

function nextEntryId(): string {
  entryCounter += 1;
  return `entry-${entryCounter}`;
}

export function entryFromEnvelope(
  envelope: EventEnvelope,
): TimelineEntry | null {
  switch (envelope.type) {
    case "EventOutput": {
      const payload = envelope.payload as EventOutputPayload;
      if (payload.Verbose) {
        return null;
      }
      return {
        id: nextEntryId(),
        variant: payload.IsErr ? "error" : "assistant",
        text: payload.Text,
      };
    }
    case "EventStoryStarted": {
      const story = envelope.payload as StoryPayload;
      const label = story.title?.trim() || story.id || "unknown";
      return {
        id: nextEntryId(),
        variant: "system",
        text: `Started story: ${label}`,
      };
    }
    case "EventStoryCompleted": {
      const payload = envelope.payload as StoryCompletedPayload;
      const story = payload.Story;
      const label = story?.title?.trim() || story?.id || "unknown";
      const outcome = payload.Success ? "completed" : "failed";
      return {
        id: nextEntryId(),
        variant: "system",
        text: `Story ${outcome}: ${label}`,
      };
    }
    case "EventError": {
      const payload = envelope.payload as EventErrorPayload;
      return {
        id: nextEntryId(),
        variant: "error",
        text: payload.error || "Unknown error",
      };
    }
    default:
      return null;
  }
}

export function makeSystemEntry(text: string): TimelineEntry {
  return {
    id: nextEntryId(),
    variant: "system",
    text,
  };
}

export function resetTimelineEntryIdsForTests(): void {
  entryCounter = 0;
}
