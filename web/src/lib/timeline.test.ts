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
