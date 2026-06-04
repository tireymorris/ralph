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

export interface ImplementationReviewStartedPayload {
  Iteration?: number;
}

export interface ImplementationFindingPayload {
  ID?: string;
  Summary?: string;
}

export interface ImplementationReviewPayload {
  Findings?: ImplementationFindingPayload[];
}

export interface ImplementationReviewCompletedPayload {
  Iteration?: number;
  Clean?: boolean;
}

let ephemeralEntryCounter = 0;

function nextEphemeralEntryId(): string {
  ephemeralEntryCounter += 1;
  return `local-${ephemeralEntryCounter}`;
}

function hashString(value: string): string {
  let hash = 0;
  for (let i = 0; i < value.length; i++) {
    hash = (hash << 5) - hash + value.charCodeAt(i);
    hash |= 0;
  }
  return (hash >>> 0).toString(36);
}

export function stableEnvelopeEntryId(envelope: EventEnvelope): string {
  const payload = JSON.stringify(envelope.payload ?? null);
  return `${envelope.type}:${hashString(payload)}`;
}

export function entryFromEnvelope(
  envelope: EventEnvelope,
): TimelineEntry | null {
  const id = stableEnvelopeEntryId(envelope);

  switch (envelope.type) {
    case "EventOutput": {
      const payload = envelope.payload as EventOutputPayload;
      if (payload.Verbose) {
        return null;
      }
      return {
        id,
        variant: payload.IsErr ? "error" : "assistant",
        text: payload.Text,
      };
    }
    case "EventStoryStarted": {
      const story = envelope.payload as StoryPayload;
      const label = story.title?.trim() || story.id || "unknown";
      return {
        id,
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
        id,
        variant: "system",
        text: `Story ${outcome}: ${label}`,
      };
    }
    case "EventImplementationReviewStarted": {
      const payload = envelope.payload as ImplementationReviewStartedPayload;
      const iteration = payload.Iteration ?? 0;
      return {
        id,
        variant: "system",
        text: `Implementation review started (iteration ${iteration})`,
      };
    }
    case "EventImplementationReview": {
      const payload = envelope.payload as ImplementationReviewPayload;
      const summaries = (payload.Findings ?? [])
        .map((f) => f.Summary?.trim())
        .filter((s): s is string => Boolean(s));
      if (summaries.length === 0) {
        return {
          id,
          variant: "system",
          text: "Implementation review reported findings",
        };
      }
      return {
        id,
        variant: "system",
        text: `Review findings: ${summaries.join("; ")}`,
      };
    }
    case "EventImplementationReviewCompleted": {
      const payload = envelope.payload as ImplementationReviewCompletedPayload;
      const iteration = payload.Iteration ?? 0;
      const outcome = payload.Clean ? "clean" : "findings";
      return {
        id,
        variant: "system",
        text: `Implementation review completed (iteration ${iteration}, ${outcome})`,
      };
    }
    case "EventError": {
      const payload = envelope.payload as EventErrorPayload;
      return {
        id,
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
    id: nextEphemeralEntryId(),
    variant: "system",
    text,
  };
}

export function resetTimelineEntryIdsForTests(): void {
  ephemeralEntryCounter = 0;
}
