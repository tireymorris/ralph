import { describe, expect, it } from "vitest";
import { entryFromEnvelope, stableEnvelopeEntryId } from "./timeline";

describe("stableEnvelopeEntryId", () => {
  it("returns the same id for the same envelope on replay", () => {
    const envelope = {
      type: "EventOutput",
      payload: { Text: "hello", IsErr: false, Verbose: false },
    };
    expect(stableEnvelopeEntryId(envelope)).toBe(stableEnvelopeEntryId(envelope));
  });

  it("returns different ids for different payloads", () => {
    const a = {
      type: "EventOutput",
      payload: { Text: "first", IsErr: false, Verbose: false },
    };
    const b = {
      type: "EventOutput",
      payload: { Text: "second", IsErr: false, Verbose: false },
    };
    expect(stableEnvelopeEntryId(a)).not.toBe(stableEnvelopeEntryId(b));
  });
});

describe("entryFromEnvelope implementation review", () => {
  it("maps started, findings, and completed events", () => {
    const started = entryFromEnvelope({
      type: "EventImplementationReviewStarted",
      payload: { Iteration: 1 },
    });
    expect(started).toEqual({
      id: stableEnvelopeEntryId({
        type: "EventImplementationReviewStarted",
        payload: { Iteration: 1 },
      }),
      variant: "system",
      text: "Implementation review started (iteration 1)",
    });

    const findings = entryFromEnvelope({
      type: "EventImplementationReview",
      payload: {
        Findings: [
          { ID: "a", Summary: "missing tests" },
          { ID: "b", Summary: "unsafe cast" },
        ],
      },
    });
    expect(findings?.variant).toBe("system");
    expect(findings?.text).toContain("missing tests");
    expect(findings?.text).toContain("unsafe cast");

    const completed = entryFromEnvelope({
      type: "EventImplementationReviewCompleted",
      payload: { Iteration: 1, Clean: true },
    });
    expect(completed?.text).toBe(
      "Implementation review completed (iteration 1, clean)",
    );
  });
});

describe("entryFromEnvelope slice events", () => {
  it("renders slice started and completed events as system entries", () => {
    const started = entryFromEnvelope({
      type: "EventSliceStarted",
      payload: { StoryID: "story-1", SliceID: "slice-1" },
    });
    expect(started).toEqual({
      id: stableEnvelopeEntryId({
        type: "EventSliceStarted",
        payload: { StoryID: "story-1", SliceID: "slice-1" },
      }),
      variant: "system",
      text: "Started slice: story-1/slice-1",
    });

    const completed = entryFromEnvelope({
      type: "EventSliceCompleted",
      payload: { StoryID: "story-1", SliceID: "slice-1" },
    });
    expect(completed).toEqual({
      id: stableEnvelopeEntryId({
        type: "EventSliceCompleted",
        payload: { StoryID: "story-1", SliceID: "slice-1" },
      }),
      variant: "system",
      text: "Completed slice: story-1/slice-1",
    });
  });
});

describe("entryFromEnvelope", () => {
  it("reuses stable ids across repeated calls", () => {
    const envelope = {
      type: "EventStoryStarted",
      payload: { id: "story-1", title: "Version metadata package" },
    };
    const first = entryFromEnvelope(envelope);
    const second = entryFromEnvelope(envelope);
    expect(first?.id).toBe(second?.id);
  });
});
