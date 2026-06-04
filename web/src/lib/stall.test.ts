import { describe, expect, it } from "vitest";
import { isRunAwaitingUser, isRunStalled, RUN_STALL_THRESHOLD_MS } from "./stall";

describe("isRunAwaitingUser", () => {
  it("returns true for waiting_review", () => {
    expect(isRunAwaitingUser("waiting_review", 0)).toBe(true);
  });

  it("returns true for waiting_clarify with questions", () => {
    expect(isRunAwaitingUser("waiting_clarify", 2)).toBe(true);
  });

  it("returns false for waiting_clarify without questions", () => {
    expect(isRunAwaitingUser("waiting_clarify", 0)).toBe(false);
  });

  it("returns false for implementing", () => {
    expect(isRunAwaitingUser("implementing", 0)).toBe(false);
  });
});

describe("isRunStalled", () => {
  it("returns false before threshold elapses", () => {
    const now = 1_000_000;
    expect(isRunStalled(now - RUN_STALL_THRESHOLD_MS + 1, now)).toBe(false);
  });

  it("returns true after threshold elapses", () => {
    const now = 1_000_000;
    expect(isRunStalled(now - RUN_STALL_THRESHOLD_MS, now)).toBe(true);
  });

  it("honors a custom threshold argument", () => {
    const now = 5_000;
    expect(isRunStalled(now - 500, now, 500)).toBe(true);
    expect(isRunStalled(now - 499, now, 500)).toBe(false);
  });
});
